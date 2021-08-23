package indexer

import (
	"github.com/streamingfast/sparkle/entity"
	pbcodec "github.com/streamingfast/sparkle/pb/dfuse/ethereum/codec/v1"
	"github.com/streamingfast/sparkle/subgraph"
)

type TestIntrinsics struct {
}

func (t TestIntrinsics) Save(entity entity.Interface) error {
	panic("implement me")
}

func (t TestIntrinsics) Load(entity entity.Interface) error {
	panic("implement me")
}

func (t TestIntrinsics) Remove(entity entity.Interface) error {
	panic("implement me")
}

func (t TestIntrinsics) Block() subgraph.BlockRef {
	panic("implement me")
}

func (t TestIntrinsics) StepBelow(step int) bool {
	panic("implement me")
}

func (t TestIntrinsics) StepAbove(step int) bool {
	panic("implement me")
}

func (t TestIntrinsics) SqlSelect(dest interface{}, query string, args ...interface{}) error {
	panic("implement me")
}

func (t TestIntrinsics) startBlock(block *pbcodec.Block, step int) {
	panic("implement me")
}

func (t TestIntrinsics) flushBlock(cursor string) error {
	panic("implement me")
}

func (t TestIntrinsics) setStep(step int) {
	panic("implement me")
}

func (t TestIntrinsics) waitForFlush() {
	panic("implement me")
}

func (t TestIntrinsics) loadCursor() (string, error) {
	panic("implement me")
}

func (t TestIntrinsics) cleanStoreAtCursor(cursor string, startBlock uint64) error {
	panic("implement me")
}

func (t TestIntrinsics) cleanUpFork(newHeadBlock uint64) error {
	panic("implement me")
}
