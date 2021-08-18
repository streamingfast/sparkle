package indexer

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/streamingfast/eth-go"
	"github.com/streamingfast/sparkle/entity"
	pbcodec "github.com/streamingfast/sparkle/pb/dfuse/ethereum/codec/v1"
	"github.com/streamingfast/sparkle/storage/dryrunstore"
	"github.com/streamingfast/sparkle/subgraph"
	"github.com/streamingfast/sparkle/testgraph/testgraph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStore_Cache(t *testing.T) {
	ctx := context.Background()
	subgraph.MainSubgraphDef = testgraph.Definition
	store := dryrunstore.New(ctx, testgraph.Definition, zap.NewNop(), 0, "/tmp/bob.json")
	obj := &testgraph.TestEntity{
		Base: entity.Base{
			ID: "1",
		},
		Name: "test-set",
		Set1: entity.NewIntFromLiteral(2),
	}
	obj.SetExists(true)
	store.Cache["test_entity"]["1"] = obj
	int := &defaultIntrinsic{
		store:   store,
		current: make(map[string]map[string]entity.Interface),
		updates: make(map[string]map[string]entity.Interface),
	}
	int.startBlock(&pbcodec.Block{
		Hash:   eth.MustNewHash("0x00"),
		Number: 100,
		Header: &pbcodec.BlockHeader{
			Timestamp: timestamppb.Now(),
		},
	}, 99999999999)
	testEnt := testgraph.NewTestEntity("1")
	err := int.Load(testEnt)
	require.NoError(t, err)
	assert.Equal(t, entity.NewIntFromLiteral(2), testEnt.Set1)

	testEnt.Set1 = entity.NewIntFromLiteral(82)

	testEntOther := testgraph.NewTestEntity("1")
	err = int.Load(testEntOther)
	require.NoError(t, err)
	assert.Equal(t, entity.NewIntFromLiteral(2), testEntOther.Set1)
}

func TestPOI(t *testing.T) {
	updates := map[string]map[string]entity.Interface{
		"friends": map[string]entity.Interface{
			"three": testgraph.NewTestEntity("3"),
			"four":  testgraph.NewTestEntity("4"),
		},
	}

	poi := entity.NewPOI("eth/testnet")
	// one table
	err := computePOI(poi, updates, &Blk{
		Id:  "aa",
		Num: 10,
	})
	require.NoError(t, err)
	digest := hex.EncodeToString(poi.Digest)
	assert.Equal(t, "2479d5bb22d6648b56cda5e5d461bf85", digest)

	updates["accounts"] = map[string]entity.Interface{
		"one": testgraph.NewTestEntity("1"),
		"two": testgraph.NewTestEntity("2"),
	}

	// two tables
	poi.Clear()
	err = computePOI(poi, updates, &Blk{
		Id:  "bb",
		Num: 11,
	})
	require.NoError(t, err)
	digest = hex.EncodeToString(poi.Digest)
	assert.Equal(t, "8289d942b6b23399fc27555585ce619f", digest)

	poi.Clear()
	err = computePOI(poi, updates, &Blk{
		Id:  "cc",
		Num: 22,
	})
	require.NoError(t, err)
	digest = hex.EncodeToString(poi.Digest)
	assert.Equal(t, "0566a532606d0387bc069bd9128b90d7", digest)

	// with blockref
	poi.Clear()
	err = computePOI(poi, updates, &Blk{
		Id:  "deadbeef",
		Num: 234,
	})
	require.NoError(t, err)
	digest = hex.EncodeToString(poi.Digest)
	assert.Equal(t, "08207c821d70a46918431e7f5d1872d6", digest)

}
