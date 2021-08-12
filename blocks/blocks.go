package blocks

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	pbany "github.com/golang/protobuf/ptypes/any"
	blockstream "github.com/streamingfast/bstream/blockstream/v2"
	dfuse "github.com/streamingfast/client-go"
	"github.com/streamingfast/dgrpc"
	"github.com/streamingfast/dstore"
	pbbstream "github.com/streamingfast/pbgo/dfuse/bstream/v1"
	pbcodec "github.com/streamingfast/sparkle/pb/dfuse/ethereum/codec/v1"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// This contains an idea about an easier way to create Firehose stream. The factory is always
// fully aware of all dependencies and is able

func init() {
	blockstream.StreamBlocksParallelFiles = 2
}

type Firehose interface {
	StreamBlocks(ctx context.Context, req *pbbstream.BlocksRequestV2) (blockstream.UnmarshalledBlockStreamV2Client, error)
}

type StreamingFastFirehoseFactory struct {
	dfuseClient  dfuse.Client
	streamClient pbbstream.BlockStreamV2Client
}

type LocalFirehoseFactory struct {
	blockstreamServer *blockstream.Server
}

func (f *LocalFirehoseFactory) StreamBlocks(ctx context.Context, req *pbbstream.BlocksRequestV2) (blockstream.UnmarshalledBlockStreamV2Client, error) {
	return f.blockstreamServer.UnmarshalledBlocksFromLocal(ctx, req), nil
}

func NewLocalFirehoseFactory(store dstore.Store) *LocalFirehoseFactory {
	stores := []dstore.Store{store}
	return &LocalFirehoseFactory{
		blockstreamServer: blockstream.NewServer(zlog, stores, nil, nil, nil, nil),
	}

}

func NewStreamingFastFirehoseFactory(apiKey string, endpoint string) (*StreamingFastFirehoseFactory, error) {

	zlog.Info("getting API initialToken")
	dfuseClient, err := dfuse.NewClient("api.streamingfast.io", apiKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create streamingfast tokenMetaClient: %w", err)
	}

	zlog.Info("connecting to streaming fast", zap.String("endpoint", endpoint))
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))}

	conn, err := dgrpc.NewExternalClient(endpoint, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to create external gRPC StreamingFast client: %w", err)
	}

	return &StreamingFastFirehoseFactory{
		dfuseClient:  dfuseClient,
		streamClient: pbbstream.NewBlockStreamV2Client(conn),
	}, nil
}

func (f *StreamingFastFirehoseFactory) StreamBlocks(ctx context.Context, req *pbbstream.BlocksRequestV2) (blockstream.UnmarshalledBlockStreamV2Client, error) {
	tokenInfo, err := f.dfuseClient.GetAPITokenInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve StreamingFast API initialToken: %w", err)
	}

	zlog.Info("retrieved dfuse API token", zap.String("token", tokenInfo.Token))
	creds := oauth.NewOauthAccess(&oauth2.Token{AccessToken: tokenInfo.Token, TokenType: "Bearer"})

	blocksClient, err := f.streamClient.Blocks(ctx, req, grpc.PerRPCCredentials(creds))
	if err != nil {
		return nil, err
	}
	return blockstream.GetUnmarshalledBlockClient(ctx, blocksClient, func(in *pbany.Any) interface{} {
		block := &pbcodec.Block{}
		if err := ptypes.UnmarshalAny(in, block); err != nil {
			panic(err)
		}
		return block
	}), nil
}
