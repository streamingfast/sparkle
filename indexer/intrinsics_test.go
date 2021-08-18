package indexer

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/streamingfast/sparkle/subgraph"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/streamingfast/eth-go"
	pbcodec "github.com/streamingfast/sparkle/pb/dfuse/ethereum/codec/v1"

	"github.com/streamingfast/sparkle/storage/dryrunstore"
	"go.uber.org/zap"

	"github.com/streamingfast/sparkle/testgraph/testgraph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/streamingfast/sparkle/entity"
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
	err := computePOI(poi, updates, nil)
	require.NoError(t, err)
	digest := hex.EncodeToString(poi.Digest)
	assert.Equal(t, "5a2e4f825a79779529815f99d6964c0a", digest)

	updates["accounts"] = map[string]entity.Interface{
		"one": testgraph.NewTestEntity("1"),
		"two": testgraph.NewTestEntity("2"),
	}

	// two tables
	poi.Clear()
	err = computePOI(poi, updates, nil)
	require.NoError(t, err)
	digest = hex.EncodeToString(poi.Digest)
	assert.Equal(t, "70a4f35f99fff6d7ad7559d247f8c9bc", digest)

	// two tables
	poi.Clear()
	err = computePOI(poi, updates, nil)
	require.NoError(t, err)
	digest = hex.EncodeToString(poi.Digest)
	assert.Equal(t, "70a4f35f99fff6d7ad7559d247f8c9bc", digest)

	// with blockref
	poi.Clear()
	err = computePOI(poi, updates, &Blk{
		Id:  "deadbeef",
		Num: 234,
	})
	require.NoError(t, err)
	digest = hex.EncodeToString(poi.Digest)
	assert.Equal(t, "3fa16e30e49cbe386774fd5ec9663f1f", digest)

}
