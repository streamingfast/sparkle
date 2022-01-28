// Copyright 2021 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blocks

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	pbcodec "github.com/streamingfast/sparkle/pb/sf/ethereum/codec/v1"
)

func BlockDecoder(blk *bstream.Block) (interface{}, error) {
	if blk.Kind() != pbbstream.Protocol_ETH {
		return nil, fmt.Errorf("expected kind %s, got %s", pbbstream.Protocol_ETH, blk.Kind())
	}

	if blk.Version() != 1 {
		return nil, fmt.Errorf("this decoder only knows about version 1, got %d", blk.Version())
	}

	block := new(pbcodec.Block)
	pl, err := blk.Payload.Get()
	if err != nil {
		return nil, fmt.Errorf("unable to get payload: %s", err)
	}

	err = proto.Unmarshal(pl, block)
	if err != nil {
		return nil, fmt.Errorf("unable to decode payload: %s", err)
	}

	// This whole BlockDecoder method is being called through the `bstream.Block.ToNative()`
	// method. Hence, it's a great place to add temporary data normalization calls to backport
	// some features that were not in all blocks yet (because we did not re-process all blocks
	// yet).
	//
	// Thoughts for the future: Ideally, we would leverage the version information here to take
	// a decision, like `do X if version <= 2.1` so we would not pay the performance hit
	// automatically instead of having to re-deploy a new version of bstream (which means
	// rebuild everything mostly)
	//
	// We reconstruct the state reverted value per call, for each transaction traces. We also
	// normalize signature curve points since we were not setting to be alwasy 32 bytes long and
	// sometimes, it would have been only 31 bytes long.
	for _, trx := range block.TransactionTraces {
		trx.PopulateStateReverted()
		trx.PopulateTrxStatus()

		if len(trx.R) > 0 && len(trx.R) != 32 {
			trx.R = normalizeSignaturePoint(trx.R)
		}

		if len(trx.S) > 0 && len(trx.S) != 32 {
			trx.S = normalizeSignaturePoint(trx.S)
		}
	}

	// We leverage StateReverted field inside the `PopulateLogBlockIndices`
	// and as such, it must be invoked after the `PopulateStateReverted` has
	// been executed.
	block.PopulateLogBlockIndices()

	return block, nil
}

func normalizeSignaturePoint(value []byte) []byte {
	if len(value) == 0 {
		return value
	}

	if len(value) < 32 {
		offset := 32 - len(value)

		out := make([]byte, 32)
		copy(out[offset:32], value)

		return out
	}

	return value[0:32]
}
