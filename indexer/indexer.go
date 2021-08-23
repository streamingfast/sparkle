package indexer

import (
	"context"
	"math"
	"sync"

	"github.com/streamingfast/eth-go/rpc"
	pbbstream "github.com/streamingfast/pbgo/dfuse/bstream/v1"
	"github.com/streamingfast/shutter"
	"github.com/streamingfast/sparkle/blocks"
	"github.com/streamingfast/sparkle/entity"
	"github.com/streamingfast/sparkle/metrics"
	"github.com/streamingfast/sparkle/storage"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

//go:generate go-enum -f=$GOFILE --marshal --names

//
// ENUM(
//   Batch
//   Live
// )
//
type Mode uint

type Option interface {
	apply(i *Indexer)
}

type optionFunc func(i *Indexer)

func (f optionFunc) apply(i *Indexer) {
	f(i)
}

func StartBlock(startBlock int64) Option {
	return optionFunc(func(i *Indexer) {
		i.startBlock = startBlock
	})
}

func StopBlock(stopBlock uint64) Option {
	return optionFunc(func(i *Indexer) {
		i.stopBlock = stopBlock
	})
}

func WithReversible() Option {
	return optionFunc(func(i *Indexer) {
		i.withReversible = true
	})
}
func WithPOI(networkName string) Option {
	return optionFunc(func(i *Indexer) {
		i.enablePOI = true
		i.networkName = networkName
	})
}
func WithNonArchiveNode() Option {
	return optionFunc(func(i *Indexer) {
		i.nonArchiveNode = true
	})
}

func UseTransactionalFlush() Option {
	return optionFunc(func(i *Indexer) { i.useTransactionalFlush = true })
}

type Indexer struct {
	*shutter.Shutter

	forkSteps     []pbbstream.ForkStep
	rpcClient     *rpc.Client
	streamFactory blocks.Firehose
	subgraph      *subgraph.Definition

	mode                  Mode
	step                  int
	startBlock            int64
	stopBlock             uint64
	useTransactionalFlush bool
	withReversible        bool
	nonArchiveNode        bool

	stateLock      sync.RWMutex
	subgraphStream *subgraphStream

	networkName string
	enablePOI   bool
}

func NewBatch(
	step int,
	startBlock uint64,
	stopBlock uint64,
	rpcClient *rpc.Client,
	streamFactory blocks.Firehose,
	subgraph *subgraph.Definition,
	opts ...Option,
) *Indexer {
	indexer := &Indexer{
		Shutter: shutter.New(),

		forkSteps:     []pbbstream.ForkStep{pbbstream.ForkStep_STEP_IRREVERSIBLE},
		rpcClient:     rpcClient,
		streamFactory: streamFactory,
		subgraph:      subgraph,
		startBlock:    int64(startBlock),
		stopBlock:     stopBlock,

		mode: ModeBatch,
		step: step,

		enablePOI: false,
	}

	for _, opt := range opts {
		opt.apply(indexer)
	}

	return indexer
}

func New(
	rpcClient *rpc.Client,
	streamFactory blocks.Firehose,
	subgraph *subgraph.Definition,
	opts ...Option,
) *Indexer {
	indexer := &Indexer{
		Shutter: shutter.New(),

		forkSteps:     []pbbstream.ForkStep{pbbstream.ForkStep_STEP_IRREVERSIBLE},
		rpcClient:     rpcClient,
		streamFactory: streamFactory,
		subgraph:      subgraph,

		mode: ModeLive,
		step: math.MaxInt32,
	}

	for _, opt := range opts {
		opt.apply(indexer)
	}

	return indexer
}

type StoreFactory func(context.Context, *zap.Logger, *metrics.BlockMetrics, *entity.Registry) (storage.Store, error)

func (i *Indexer) Start(makeStore StoreFactory) error {
	streamCtx, cancel := context.WithCancel(context.Background())

	logger := zlog.With(zap.String("subgraph", i.subgraph.PackageName))
	logger.Info("instantiating new store for a graph")

	intrinsics := newDefaultIntrinsic(streamCtx, i.step, i.rpcClient)
	zlog.Info("intrinsics initiated")
	/// read db to get the corresponing Qz.... iD
	///

	if i.nonArchiveNode {
		intrinsics.nonArchiveNode = true
	}
	if i.enablePOI {
		intrinsics.enablePOI = true
		intrinsics.networkName = i.networkName
		intrinsics.aggregatePOI = i.mode == ModeLive
	}

	subgraphInst := i.subgraph.New(subgraph.Base{
		Log:        logger,
		Intrinsics: intrinsics,
		Definition: i.subgraph,
	})

	zlog.Info("subgraph instance initiated")
	metric := metrics.NewBlockMetrics()
	zlog.Info("metric initiated")
	store, err := makeStore(streamCtx, logger, metric, i.subgraph.Entities)
	if err != nil {
		return err
	}
	intrinsics.store = store
	zlog.Info("store initiated")

	// FIXME: This starts to receives lots of stuff that are actually created right here by the indexer.
	//        I think most of the creation above should actually be moved down stream in the `newSubgraphStream`
	//        method directly.
	stream := newSubgraphStream(
		logger,
		intrinsics,
		store,
		metric,
		i.streamFactory, // this is the blockFactory, we want inputStoreFactory ???
		i.subgraph,
		subgraphInst,
		i.step,
		i.resolveStartBlock(),
		i.stopBlock,
		i.withReversible,
	)
	zlog.Info("stream initiated")

	// we only want to shut down the indexer, once the stream
	// is completed terminated (i.e. when the Unmarshal blocks have been drained)
	stream.OnTerminated(func(err error) {
		i.Shutdown(err)
	})

	i.OnTerminating(func(err error) {
		if err != nil {
			zlog.Error("indexer shutting down due to error", zap.Error(err))
		}

		zlog.Info("shutting down subgraph stream")
		i.subgraphStream.Shutdown(nil)

		if err := store.Close(); err != nil {
			zlog.Error("failed closing the store", zap.Error(err))
		}
		cancel()
	})

	go func() {
		zlog.Info("starting block stream")
		stream.Start()
		logger.Info("stream run loop ended")
	}()
	i.subgraphStream = stream

	return nil
}

func (i *Indexer) resolveStartBlock() (start int64) {
	if i.mode == ModeBatch || i.startBlock != 0 {
		return i.startBlock
	}
	return int64(i.subgraph.StartBlock)
}
