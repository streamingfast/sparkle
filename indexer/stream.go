package indexer

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/bstream/stream"
	"github.com/streamingfast/dmetrics"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	"github.com/streamingfast/shutter"
	"github.com/streamingfast/sparkle/blocks"
	"github.com/streamingfast/sparkle/metrics"
	pbcodec "github.com/streamingfast/sparkle/pb/sf/ethereum/type/v2"
	"github.com/streamingfast/sparkle/storage"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

var MetricSet = dmetrics.NewSet()

type subgraphStream struct {
	*shutter.Shutter

	cleanupNeeded bool
	forkHeadRef   bstream.BlockRef
	lastBlockRef  bstream.BlockRef

	in            privateIntrinsic
	store         storage.Store
	streamFactory blocks.Firehose
	subgraph      *subgraph.Definition
	subgraphInst  subgraph.Subgraph

	headBlockTimeDrift *dmetrics.HeadTimeDrift
	headBlockNumber    *dmetrics.HeadBlockNum

	step           int
	startBlock     int64
	stopBlock      uint64
	withReversible bool

	metrics         *metrics.BlockMetrics
	statusFrequency time.Duration
	nextStatus      time.Time
	syncStatus      *SyncStatus
	doneFlush       chan struct{}

	logger *zap.Logger
}

func newSubgraphStream(
	logger *zap.Logger,
	in privateIntrinsic,
	store storage.Store,
	metrics *metrics.BlockMetrics,
	streamFactory blocks.Firehose,
	subgraph *subgraph.Definition,
	subgraphInst subgraph.Subgraph,
	step int,
	startBlock int64,
	stopBlock uint64,
	withReversible bool,
) *subgraphStream {
	return &subgraphStream{
		Shutter:       shutter.New(),
		in:            in,
		store:         store,
		streamFactory: streamFactory,
		subgraph:      subgraph,
		subgraphInst:  subgraphInst,

		headBlockTimeDrift: MetricSet.NewHeadTimeDrift(subgraph.PackageName),
		headBlockNumber:    MetricSet.NewHeadBlockNumber(subgraph.PackageName),

		step:           step,
		startBlock:     startBlock,
		stopBlock:      stopBlock,
		withReversible: withReversible,

		metrics:         metrics,
		statusFrequency: 10 * time.Second,
		syncStatus:      NewSyncStatus(),

		logger: logger,
	}
}

func (s *subgraphStream) Start() {
	s.logger.Info("starting blocks stream")

	streamCtx, cancelStream := context.WithCancel(context.Background())
	unmarshalledBlocks := make(chan *unmarshalledBlock, 2)
	unmarshalProcessDone := make(chan bool, 1)

	cursor, err := s.in.loadCursor()
	if err != nil {
		s.Shutdown(fmt.Errorf("unable to load cursor of subgraph %q: %w", s.subgraph.PackageName, err))
		return
	}

	startBlock, err := func(cursor string, blockNum uint64) (uint64, error) {
		if cursor == "" {
			return blockNum, nil
		}

		cursorObj, err := bstream.CursorFromOpaque(cursor)
		if err != nil {
			return 0, fmt.Errorf("cursor from opaque: %w", err)
		}

		if s.withReversible {
			return cursorObj.HeadBlock.Num(), nil //headblock never goes backwards in our irreversibility logic, safer than just using cursor.Block.Num()
		}
		return cursorObj.LIB.Num(), nil
	}(cursor, uint64(s.startBlock))

	if err != nil {
		s.Shutdown(fmt.Errorf("could not resolve start block: %w", err))
		return
	}

	if cursor != "" {
		s.logger.Info("loaded initial cursor, cleaning store", zap.String("cursor", cursor))
		if err := s.in.cleanStoreAtBlock(startBlock); err != nil {
			s.Shutdown(fmt.Errorf("unable to clean before cursor %q with start block %d of subgraph %q: %w", cursor, s.startBlock, s.subgraph.PackageName, err))
			return
		}
	} else {
		s.logger.Info("no cursor found, skipping cleanup")
	}

	s.logger.Info("cleaned store, loading dynamic data sources")
	if err := s.subgraphInst.LoadDynamicDataSources(startBlock); err != nil {
		s.Shutdown(fmt.Errorf("unable to load dynamic datasources for subgraph %q: %w", s.subgraph.PackageName, err))
		return
	}

	if err := s.subgraphInst.Init(); err != nil {
		s.Shutdown(fmt.Errorf("unable to init subgraph %q: %w", s.subgraph.PackageName, err))
		return
	}

	s.lastBlockRef = bstream.BlockRefEmpty
	s.nextStatus = time.Now().Add(s.statusFrequency)

	s.OnTerminating(func(err error) {
		if err != nil {
			s.logger.Error("subgraph shutting down due to error", zap.Error(err))
		}
		s.logger.Info("subgraph is terminating, cancelling block stream")
		cancelStream()
		s.logger.Info("waiting on unmarshall process completion")
		<-unmarshalProcessDone
		s.logger.Info("unmarshall process completed")
	})

	go func() {
		zlog.Info("starting unmarhsall process loop")
		err := s.unmarshallProcessLoop(unmarshalledBlocks, unmarshalProcessDone)
		if err != nil {
			s.Shutdown(err)
		}
	}()

	s.Shutdown(s.streamLoop(streamCtx, unmarshalledBlocks))
}

func (s *subgraphStream) streamLoop(streamCtx context.Context, unmarshalledBlocks chan *unmarshalledBlock) error {
	defer func() {
		s.logger.Info("closing unmarshalled block channel")
		close(unmarshalledBlocks) // since you are shutting down, you should not be writting more blocks
	}()
	retryDelay := 0 * time.Second
	for {
		if s.IsTerminating() {
			s.logger.Info("not retrying stream, subgraph has is terminating")
			return nil
		}

		time.Sleep(retryDelay)

		retryDelay = 5 * time.Second

		// FIXME: Quand je vais m'apprêter à reconnecter ici, est-ce
		// que ma func qui consomme `unmarshalledBlock` est well
		// finished ou flushed ou completedly processed? Est-ce que
		// mon cursor va être legit? Est-ceq e mon lastBlockRef va
		// réfléter vraiment tout ce que j'ai pié dans
		// `unmarshalledBlock`
		waitForEmptyChannel(unmarshalledBlocks)

		zlog.Info("connecting new streaming")
		err := s.processNewStream(streamCtx, unmarshalledBlocks)
		if err != nil {
			if err == stream.ErrStopBlockReached {
				return nil
			}
			return err
		}
	}
}

func (s *subgraphStream) unmarshallProcessLoop(unmarshalledBlocks chan *unmarshalledBlock, unmarshalProcessDone chan bool) error {
	defer close(unmarshalProcessDone)
	for {
		t0 := time.Now()

		response, ok := <-unmarshalledBlocks
		if !ok {
			s.logger.Warn("unmarshal channel closed")
			return nil
		}

		if err := s.processUnmarshalledBlock(response); err != nil {
			return err
		}

		s.headBlockNumber.SetUint64(response.block.Number)
		if response.block.Header != nil && response.block.Header.Timestamp != nil {
			s.headBlockTimeDrift.SetBlockTime(response.block.Header.Timestamp.AsTime())
		}

		s.metrics.Exec.WaitForBlock += time.Since(t0)
	}

}

// processNewStream returns `nil` to indicate we need to continue, an error if there was a fatal error.
func (s *subgraphStream) processNewStream(ctx context.Context, unmarshalledBlocks chan *unmarshalledBlock) error {
	s.logger.Info("retrieving cursor")
	cursor, err := s.in.loadCursor()
	if err != nil {
		return fmt.Errorf("unable to load cursor of subgraph %q: %w", s.subgraph.PackageName, err)
	}

	// Launch streaming thing

	forkSteps := []pbfirehose.ForkStep{pbfirehose.ForkStep_STEP_IRREVERSIBLE}
	if s.withReversible {
		forkSteps = []pbfirehose.ForkStep{
			pbfirehose.ForkStep_STEP_NEW,
			pbfirehose.ForkStep_STEP_UNDO,
		}
	}

	s.logger.Info("requesting blocks", zap.String("cursor", cursor), zap.Int64("start_block_num", s.startBlock), zap.Uint64("stop_block_num", s.stopBlock))
	myStream, err := s.streamFactory.StreamBlocks(ctx, &pbfirehose.Request{
		StartBlockNum:     s.startBlock,
		StartCursor:       cursor,
		StopBlockNum:      s.stopBlock,
		ForkSteps:         forkSteps,
		IncludeFilterExpr: s.subgraph.IncludeFilter,
	})
	if err != nil {
		return fmt.Errorf("unable to create blocks stream of subgraph %q: %w", s.subgraph.PackageName, err)
	}

	s.logger.Info("streaming blocks", zap.Uint64("stop_block_num", s.stopBlock))

	defer s.metrics.BlockRate.Clean()

	for {
		if s.IsTerminating() {
			s.logger.Info("stopping streaming loop")
			return nil
		}

		response, err := myStream.Recv()
		if (response == nil) && (err == nil) {
			err = io.EOF // FIXME in bstream lib, stepd hack
		}

		if err == io.EOF {
			if s.stopBlock != 0 {
				if s.lastBlockRef.Num() == s.stopBlock {
					s.logger.Info("reached our stop block", zap.Stringer("last_block_ref", s.lastBlockRef))
					return stream.ErrStopBlockReached
				}
				return fmt.Errorf("stream ended with EOF but last block seen %q does not match expected stop block %q", s.lastBlockRef.Num(), s.stopBlock)
			}

			var cursor string
			if response != nil {
				cursor = response.Cursor
			}
			s.logger.Error("received EOF when not expected, reconnecting",
				zap.String("cursor", cursor),
				zap.Stringer("last_block", s.lastBlockRef),
				zap.Uint64("stop_block", s.stopBlock),
			)
			return nil
		}

		if err != nil {
			s.logger.Error("stream encountered a remote error, retrying",
				zap.Stringer("last_block", s.lastBlockRef),
				zap.Error(err),
			)
			return nil
		}

		block := &pbcodec.Block{}
		if err := ptypes.UnmarshalAny(response.Block, block); err != nil {
			panic(err)
		}

		s.lastBlockRef = block.AsRef()

		unmarshalledBlocks <- &unmarshalledBlock{
			block:  block,
			cursor: response.Cursor,
			step:   response.Step,
		}
	}
}

func (s *subgraphStream) processUnmarshalledBlock(response *unmarshalledBlock) (err error) {
	// Accumulate UNDO steps, and clean-up on the very next NEW
	if response.step == pbfirehose.ForkStep_STEP_UNDO {
		s.cleanupNeeded = true
		if s.forkHeadRef == nil {
			s.forkHeadRef = response.block.AsRef()
		}
		return nil
	}

	if s.cleanupNeeded {
		zlog.Info("fork detected", zap.Stringer("this is a new step block", response.block.AsRef()), zap.Stringer("this ", s.forkHeadRef))
		err = s.in.cleanUpFork(response.block.Number)
		if err != nil {
			return err
		}
		s.forkHeadRef = nil
		s.cleanupNeeded = false
	}

	cursor := response.cursor
	block := response.block
	// HeadBlockNumber.SetUint64(block.Number)
	if block.Header != nil && block.Header.Timestamp != nil {
		// HeadBlockTimeDrift.SetBlockTime(block.Header.Timestamp.AsTime())
	}

	if err = s.dispatchToSubgraph(block, cursor); err != nil {
		return fmt.Errorf("unable to dispatch received block to subgraph: %w", err)
	}

	now := time.Now()
	if now.After(s.nextStatus) {
		s.nextStatus = now.Add(s.statusFrequency)
		s.logger.Info(fmt.Sprintf("loader stats each %s", s.statusFrequency), zap.Object("stats", s.metrics))
		s.metrics.Exec.Clean()
		s.metrics.BlockRate.Clean()

		s.subgraphInst.LogStatus()
	}

	s.syncStatus.Update(block.Number)

	return nil
}

func (s *subgraphStream) dispatchToSubgraph(block *pbcodec.Block, cursor string) error {
	t0 := time.Now()
	s.metrics.LastBlockRef = block.AsRef()

	s.in.startBlock(block, s.step)

	tblkProc0 := time.Now()
	err := s.subgraphInst.HandleBlock(block)
	if err != nil {
		return fmt.Errorf("could not handle block %s: %w", block.AsRef(), err)
	}

	s.metrics.Exec.BlockProc += time.Since(tblkProc0)

	tsflush0 := time.Now()

	if err = s.in.flushBlock(cursor); err != nil {
		return fmt.Errorf("flush block %s: %w", block.AsRef(), err)
	}

	s.metrics.Exec.StoreFlush += time.Since(tsflush0)
	s.metrics.Exec.Finalize(time.Since(t0))

	s.metrics.BlockRate.Inc()

	return nil
}

type unmarshalledBlock struct {
	block  *pbcodec.Block
	err    error
	cursor string
	step   pbfirehose.ForkStep
}

func waitForEmptyChannel(ch chan *unmarshalledBlock) {
	wait := 0 * time.Second
	for {
		time.Sleep(wait)
		wait = 100 * time.Millisecond

		if len(ch) != 0 {
			continue
		}
		return
	}
}
