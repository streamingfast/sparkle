package blocks

import (
	"context"
	"crypto/tls"
	"fmt"

	dfuse "github.com/streamingfast/client-go"
	"github.com/streamingfast/dgrpc"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// This contains an idea about an easier way to create Firehose stream. The factory is always
// fully aware of all dependencies and is able

type Firehose interface {
	StreamBlocks(ctx context.Context, req *pbfirehose.Request) (pbfirehose.Stream_BlocksClient, error)
}

type StreamingFastFirehoseFactory struct {
	dfuseClient  dfuse.Client
	streamClient pbfirehose.StreamClient
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
		streamClient: pbfirehose.NewStreamClient(conn),
	}, nil
}

func (f *StreamingFastFirehoseFactory) StreamBlocks(ctx context.Context, req *pbfirehose.Request) (pbfirehose.Stream_BlocksClient, error) {
	tokenInfo, err := f.dfuseClient.GetAPITokenInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve StreamingFast API initialToken: %w", err)
	}

	zlog.Info("retrieved dfuse API token", zap.String("token", tokenInfo.Token))
	creds := oauth.NewOauthAccess(&oauth2.Token{AccessToken: tokenInfo.Token, TokenType: "Bearer"})

	return f.streamClient.Blocks(ctx, req, grpc.PerRPCCredentials(creds))
	//		block := &pbcodec.Block{}
	//		if err := ptypes.UnmarshalAny(in, block); err != nil {
	//			panic(err)
	//		}
	//		return block
}
