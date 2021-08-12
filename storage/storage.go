package storage

import (
	"context"

	pbcodec "github.com/streamingfast/sparkle/pb/dfuse/ethereum/codec/v1"

	"github.com/streamingfast/sparkle/entity"
)

type Store interface {
	// FIXME: get rid of *pbcodec, an ethereum-centric dependency, we only use `Number` and `Time` in here..
	BatchSave(ctx context.Context, block *pbcodec.Block, updates map[string]map[string]entity.Interface, cursor string) (err error)
	Load(ctx context.Context, id string, entity entity.Interface, blockNum uint64) error
	LoadAllDistinct(ctx context.Context, model entity.Interface, blockNum uint64) ([]entity.Interface, error)

	LoadCursor(ctx context.Context) (string, error)

	CleanDataAtBlock(ctx context.Context, blockNum uint64) error
	CleanUpFork(ctx context.Context, newHeadBlock uint64) error

	Close() error
}
