package indexer

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"time"

	"github.com/streamingfast/eth-go"
	"github.com/streamingfast/eth-go/rpc"
	"github.com/streamingfast/sparkle/entity"
	pbcodec "github.com/streamingfast/sparkle/pb/dfuse/ethereum/codec/v1"
	"github.com/streamingfast/sparkle/storage"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

var debugPOI = os.Getenv("DEBUG_POI") == "true"

type privateIntrinsic interface {
	subgraph.Intrinsics

	startBlock(block *pbcodec.Block, step int)
	flushBlock(cursor string) error
	setStep(step int)

	loadCursor() (string, error)
	cleanStoreAtBlock(startBlock uint64) error
	cleanUpFork(longestChainStartBlock uint64) error
}

var _ privateIntrinsic = (*defaultIntrinsic)(nil)

type defaultIntrinsic struct {
	ctx       context.Context
	rpcClient *rpc.Client
	store     storage.Store

	enablePOI      bool
	nonArchiveNode bool
	rpcCache       *RPCCache
	aggregatePOI   bool
	networkName    string

	step int

	block    *pbcodec.Block
	blockRef *blockRef

	// cached entities
	current map[string]map[string]entity.Interface
	updates map[string]map[string]entity.Interface
}

//newDefaultIntrinsic does not set the store right away, set it afterwards
func newDefaultIntrinsic(ctx context.Context, step int, rpcClient *rpc.Client) *defaultIntrinsic {
	return &defaultIntrinsic{
		ctx:       ctx,
		rpcClient: rpcClient,
		step:      step,
		enablePOI: false,

		current: make(map[string]map[string]entity.Interface),
		updates: make(map[string]map[string]entity.Interface),
	}
}

func (d *defaultIntrinsic) Save(ent entity.Interface) error {
	tableName := entity.GetTableName(ent)
	tbl, found := d.updates[tableName]
	if !found {
		tbl = make(map[string]entity.Interface)
		d.updates[tableName] = tbl
	}

	// WARN: what's the impact of setting this is we HAVEN'T
	// fetched it from the DB. Don't we rely on the `Exists()` to
	// mean "it exists in the DB" ?
	ent.SetExists(true)
	ent.SetMutated(d.step)

	tbl[ent.GetID()] = ent

	return nil
}

func (d *defaultIntrinsic) Load(ent entity.Interface) error {
	tableName := entity.GetTableName(ent)
	id := ent.GetID()
	zlog.Debug("loading entity",
		zap.String("id", id),
		zap.String("table", tableName),
		zap.Uint64("vid", ent.GetVID()),
	)
	if id == "" {
		return fmt.Errorf("id was not set before calling load")
	}

	// First check from updates
	updateTbl, found := d.updates[tableName]
	if !found {
		updateTbl = make(map[string]entity.Interface)
		d.updates[tableName] = updateTbl
	}

	cachedEntity, found := updateTbl[id]
	if found {
		if cachedEntity == nil {
			return nil
		}
		ve := reflect.ValueOf(ent).Elem()
		ve.Set(reflect.ValueOf(cachedEntity).Elem())
		return nil
	}

	// Load from DB otherwise
	currentTbl, found := d.current[tableName]
	if !found {
		currentTbl = make(map[string]entity.Interface)
		d.current[tableName] = currentTbl
	}

	cachedEntity, found = currentTbl[id]
	if found {
		if cachedEntity == nil {
			return nil
		}
		ve := reflect.ValueOf(ent).Elem()
		ve.Set(reflect.ValueOf(cachedEntity).Elem())
		return nil
	}

	if err := d.store.Load(d.ctx, id, ent, d.blockRef.Number()); err != nil {
		return fmt.Errorf("failed loading entity: %w", err)
	}

	if ent.Exists() {
		reflectType, ok := subgraph.MainSubgraphDef.Entities.GetType(tableName)
		if !ok {
			return fmt.Errorf("unable to retrieve entity type")
		}
		clone := reflect.New(reflectType).Interface()
		ve := reflect.ValueOf(clone).Elem()
		ve.Set(reflect.ValueOf(ent).Elem())
		currentTbl[id] = clone.(entity.Interface)
	} else {
		currentTbl[id] = nil
	}

	return nil
}

func (d *defaultIntrinsic) LoadAllDistinct(model entity.Interface, blockNum uint64) ([]entity.Interface, error) {
	return d.store.LoadAllDistinct(d.ctx, model, blockNum)
}

func (d *defaultIntrinsic) Remove(e entity.Interface) error {
	tableName := entity.GetTableName(e)

	tbl, found := d.updates[tableName]
	if !found {
		tbl = make(map[string]entity.Interface)
		d.updates[tableName] = tbl
	}

	tbl[e.GetID()] = nil
	return nil
}

func (d *defaultIntrinsic) Block() subgraph.BlockRef {
	return d.blockRef
}

func (d *defaultIntrinsic) StepBelow(step int) bool {
	return d.step < step
}

func (d *defaultIntrinsic) StepAbove(step int) bool {
	return d.step > step
}

// RPCCalls retries eternally until it gets deterministic results / errors...
// However, it will return an error if your input calls cannot be encoded.
// Good luck!
func (d *defaultIntrinsic) RPC(calls []*subgraph.RPCCall) ([]*subgraph.RPCResponse, error) {

	opts := []rpc.ETHCallOption{}
	if !d.nonArchiveNode {
		opts = append(opts, rpc.AtBlockNum(d.blockRef.num))
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
	if d.rpcCache != nil {
		var cacheKeyParts []interface{}
		if !d.nonArchiveNode {
			cacheKeyParts = append(cacheKeyParts, d.blockRef.num)
		}
		for _, call := range calls {
			cacheKeyParts = append(cacheKeyParts, call.ToString())
		}
		cacheKey = d.rpcCache.Key("rpc", cacheKeyParts...)

		if fromCache, found := d.rpcCache.GetRaw(cacheKey); found {
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

		out, err := d.rpcClient.DoRequests(reqs)
		if err != nil {
			zlog.Warn("retrying RPCCall on RPC error", zap.Error(err), zap.Uint64("at_block", d.blockRef.num))
			continue
		}

		for _, resp := range out {
			if !resp.Deterministic() {
				zlog.Warn("retrying RPCCall on non-deterministic RPC call error", zap.Error(resp.Err), zap.Uint64("at_block", d.blockRef.num))
				continue
			}
		}
		if d.rpcCache != nil {
			d.rpcCache.Set(cacheKey, out)
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

func (d *defaultIntrinsic) setStep(step int) {
	d.step = step
}

func (d *defaultIntrinsic) startBlock(block *pbcodec.Block, step int) {
	d.block = block
	d.blockRef = asBlockRef(block)

	d.step = step

	d.current = make(map[string]map[string]entity.Interface)
	d.updates = make(map[string]map[string]entity.Interface)
}

func (d *defaultIntrinsic) flushBlock(cursor string) error {
	if d.enablePOI {
		poi, err := d.generatePOI()
		if err != nil {
			return fmt.Errorf("unable to generate POI")
		}

		err = d.Save(poi)
		if err != nil {
			return fmt.Errorf("unable to save generated POI: %w", err)
		}
	}

	return d.store.BatchSave(d.ctx, d.block, d.updates, cursor)
}

func (d *defaultIntrinsic) generatePOI() (*entity.POI, error) {
	blk := &Blk{
		Id:  d.block.ID(),
		Num: d.block.Number,
	}
	if debugPOI {
		zlog.Info("generating POI", zap.Reflect("block", blk), zap.Reflect("updates", d.updates))
	}
	poi := entity.NewPOI(d.networkName)
	if err := d.Load(poi); err != nil {
		return nil, err
	}

	if !d.aggregatePOI {
		poi.Clear() // discard md5 and digest information...
	}

	if err := computePOI(poi, d.updates, blk); err != nil {
		return nil, err
	}

	return poi, nil
}

func (d *defaultIntrinsic) loadCursor() (string, error) {
	return d.store.LoadCursor(d.ctx)
}

func (d *defaultIntrinsic) cleanStoreAtBlock(startBlock uint64) error {
	return d.store.CleanDataAtBlock(d.ctx, startBlock)
}

func (d *defaultIntrinsic) cleanUpFork(longestChainStartBlock uint64) error {
	return d.store.CleanUpFork(d.ctx, longestChainStartBlock)
}

type blockRef struct {
	id        string
	num       uint64
	timestamp time.Time
}

func (b blockRef) ID() string {
	return b.id
}

func (b blockRef) Number() uint64 {
	return b.num
}

func (b blockRef) Timestamp() time.Time {
	return b.timestamp
}

func asBlockRef(block *pbcodec.Block) *blockRef {
	return &blockRef{id: block.ID(), num: block.Number, timestamp: block.Header.Timestamp.AsTime()}
}

func computePOI(poi *entity.POI, updates map[string]map[string]entity.Interface, blockRef *Blk) error {
	count := 0

	previousPOIDigest := poi.Digest
	poi.Clear()

	tblNames := make([]string, 0, len(updates))
	for k := range updates {
		tblNames = append(tblNames, k)
	}
	sort.Strings(tblNames)

	for _, tblName := range tblNames {
		tblUpdates := updates[tblName]
		rowIDs := make([]string, 0, len(tblUpdates))
		for k := range tblUpdates {
			rowIDs = append(rowIDs, k)
		}
		sort.Strings(rowIDs)

		for _, id := range rowIDs {
			row := tblUpdates[id]
			count++

			if row == nil {
				if err := poi.RemoveEnt(tblName, id); err != nil {
					return fmt.Errorf("unable to write entity in POI: %w", err)
				}
				continue
			}

			if err := poi.AddEnt(tblName, row); err != nil {
				return fmt.Errorf("unable to write entity in POI: %w", err)
			}
		}
	}

	zlog.Debug("encoded update in POI",
		zap.Int("update_count", count),
		zap.Reflect("block", blockRef),
	)

	if err := poi.AddEnt("blocks", blockRef); err != nil {
		return fmt.Errorf("unable to write block ref POI: %w", err)
	}

	poi.Apply()
	zlog.Debug("unit poi", zap.String("digest", hex.EncodeToString(poi.Digest)))
	if previousPOIDigest != nil { // we are aggregating
		poi.AggregateDigest(previousPOIDigest)
		zlog.Debug("aggregated poi", zap.String("digest", hex.EncodeToString(poi.Digest)))
	}
	return nil
}

type Blk struct {
	Id  string
	Num uint64
}
