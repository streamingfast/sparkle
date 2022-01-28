package indexer

import (
	"context"

	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	"google.golang.org/grpc/metadata"
)

type TestBlocksClient struct {
	blocks []*pbfirehose.Response
}

// func (pbfirehose.Stream_BlocksClient).Recv() (*pbfirehose.Response, error)
func (t *TestBlocksClient) Recv() (*pbfirehose.Response, error) {
	var res *pbfirehose.Response
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
