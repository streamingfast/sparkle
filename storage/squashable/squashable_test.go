package squashable

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/streamingfast/dstore"
	"github.com/streamingfast/sparkle/entity"
	"github.com/streamingfast/sparkle/testgraph/testgraph"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStore_SnapshotWriter(t *testing.T) {
	t.Skipf("skip snapshot writer")
	ctx := context.Background()
	testSquasheableStore := &store{
		logger: zap.NewNop(),
		cache: map[string]map[string]entity.Interface{
			"test_entity": {
				"1": &testgraph.TestEntity{
					Base: entity.Base{
						ID: "1",
						BlockRange: &entity.BlockRange{
							StartBlock: 100,
							EndBlock:   101,
						},
						UpdatedBlockNum: 101,
					},
					Name:                    "test entity name",
					Set1:                    entity.NewIntFromLiteral(2),
					Set3:                    "testify",
					Counter1:                entity.NewIntFromLiteral(4),
					Counter2:                entity.NewFloatFromLiteral(3.6159),
					Counter3:                entity.NewIntFromLiteral(2932).Ptr(),
					DerivedFromCounter1And2: entity.NewFloatFromLiteral(1.666666),
				},
			},
		},
	}
	testStore, err := dstore.NewStore("/tmp/squash-test", "", "", false)
	require.NoError(t, err)
	filename, err := testSquasheableStore.WriteSnapshot(testStore)
	require.NoError(t, err)

	newTestSquasheableStore := &store{
		cache:    make(map[string]map[string]entity.Interface),
		subgraph: testgraph.Definition,
		logger:   zap.NewNop(),
	}

	err = newTestSquasheableStore.loadSnapshotFile(ctx, testStore, filename)
	require.NoError(t, err)
	spew.Dump(newTestSquasheableStore.cache)
}
