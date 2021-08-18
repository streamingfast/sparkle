package indexer

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

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
	assert.Equal(t, "033eec8ab31dece34b8a6bc4bd7dc3d1", digest)

	updates["accounts"] = map[string]entity.Interface{
		"one": testgraph.NewTestEntity("1"),
		"two": testgraph.NewTestEntity("2"),
	}

	// two tables
	poi.Clear()
	err = computePOI(poi, updates, nil)
	require.NoError(t, err)
	digest = hex.EncodeToString(poi.Digest)
	assert.Equal(t, "8daa2db36c925c265204b88268aa8d4a", digest)

	// two tables
	poi.Clear()
	err = computePOI(poi, updates, nil)
	require.NoError(t, err)
	digest = hex.EncodeToString(poi.Digest)
	assert.Equal(t, "8daa2db36c925c265204b88268aa8d4a", digest)

	// with blockref
	poi.Clear()
	err = computePOI(poi, updates, &testBlockRef{
		id:     "deadbeef",
		number: 234,
	})
	require.NoError(t, err)
	digest = hex.EncodeToString(poi.Digest)
	assert.Equal(t, "66a67edd4c3fdd4a2f8bf7182d8f60e8", digest)

}

type testBlockRef struct {
	id        string
	number    uint64
	timestamp time.Time
}

func (b *testBlockRef) ID() string {
	return b.id
}

func (b *testBlockRef) Number() uint64 {
	return b.number
}

func (b *testBlockRef) Timestamp() time.Time {
	return b.timestamp
}
