package indexer

import (
	"context"

	pbbstream "github.com/streamingfast/pbgo/dfuse/bstream/v1"
	"google.golang.org/grpc/metadata"
)

type TestBlocksClient struct {
	blocks []*pbbstream.BlockResponseV2
}

func (t *TestBlocksClient) Recv() (*pbbstream.BlockResponseV2, error) {
	var res *pbbstream.BlockResponseV2
	res, t.blocks = t.blocks[len(t.blocks)-1], t.blocks[:len(t.blocks)-1]
	return res, nil
}

func (t TestBlocksClient) Header() (metadata.MD, error) {
	panic("implement me")
}

func (t TestBlocksClient) Trailer() metadata.MD {
	panic("implement me")
}

func (t TestBlocksClient) CloseSend() error {
	panic("implement me")
}

func (t TestBlocksClient) Context() context.Context {
	panic("implement me")
}

func (t TestBlocksClient) SendMsg(m interface{}) error {
	panic("implement me")
}

func (t TestBlocksClient) RecvMsg(m interface{}) error {
	panic("implement me")
}
