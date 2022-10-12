package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/streamingfast/eth-go"
	"github.com/streamingfast/eth-go/rpc"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

func DoRPCCalls(noArchiveNode bool, blockNum uint64, rpcClient *rpc.Client, rpcCache *RPCCache, calls []*subgraph.RPCCall) ([]*subgraph.RPCResponse, error) {
	opts := []rpc.ETHCallOption{}
	if noArchiveNode {
		opts = append(opts, rpc.AtBlockNum(blockNum))
	}

	var reqs []*rpc.RPCRequest
	for _, call := range calls {
		method, err := eth.NewMethodDef(call.MethodSignature)
		if err != nil {
			return nil, fmt.Errorf("invalid method signature %s: %w", call.MethodSignature, err)
		}
		addr, err := eth.NewAddress(call.ToAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid address %s: %w", call.ToAddr, err)
		}
		reqs = append(reqs, rpc.NewETHCall(addr, method, opts...).ToRequest())
	}

	var cacheKey RPCCacheKey
	if rpcCache != nil {
		var cacheKeyParts []interface{}
		if noArchiveNode {
			cacheKeyParts = append(cacheKeyParts, blockNum)
		}
		for _, call := range calls {
			cacheKeyParts = append(cacheKeyParts, call.ToString())
		}
		cacheKey = rpcCache.Key("rpc", cacheKeyParts...)

		if fromCache, found := rpcCache.GetRaw(cacheKey); found {
			rpcResp := []*rpc.RPCResponse{}
			err := json.Unmarshal(fromCache, &rpcResp)
			if err != nil {
				zlog.Warn("cannot unmarshal cache response for rpc call", zap.Error(err))
			} else {
				for i, resp := range rpcResp {
					resp.CopyDecoder(reqs[i])
				}
				resps := toSubgraphRPCResponse(rpcResp)
				return resps, nil
			}
		}
	}

	var delay time.Duration
	var attemptNumber int
	for {
		time.Sleep(delay)

		attemptNumber += 1
		delay = minDuration(time.Duration(attemptNumber*500)*time.Millisecond, 10*time.Second)

		out, err := rpcClient.DoRequests(context.Background(), reqs)
		if err != nil {
			zlog.Warn("retrying RPCCall on RPC error", zap.Error(err), zap.Uint64("at_block", blockNum))
			continue
		}

		var nonDeterministicResp bool
		for _, resp := range out {
			if !resp.Deterministic() {
				zlog.Warn("retrying RPCCall on non-deterministic RPC call error", zap.Error(resp.Err), zap.Uint64("at_block", blockNum))
				nonDeterministicResp = true
				break
			}
		}
		if nonDeterministicResp {
			continue
		}
		if rpcCache != nil {
			rpcCache.Set(cacheKey, out)
		}
		resp := toSubgraphRPCResponse(out)
		return resp, nil
	}

}

func toSubgraphRPCResponse(in []*rpc.RPCResponse) (out []*subgraph.RPCResponse) {
	for _, rpcResp := range in {
		decoded, decodingError := rpcResp.Decode()
		out = append(out, &subgraph.RPCResponse{
			Decoded:       decoded,
			DecodingError: decodingError,
			CallError:     rpcResp.Err,
			Raw:           rpcResp.Content,
		})
	}
	return
}

func minDuration(a, b time.Duration) time.Duration {
	if a <= b {
		return a
	}
	return b
}
