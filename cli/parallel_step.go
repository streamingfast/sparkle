package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	manifestlib "github.com/streamingfast/sparkle/manifest"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dstore"
	"github.com/streamingfast/eth-go/rpc"
	"github.com/streamingfast/sparkle/blocks"
	"github.com/streamingfast/sparkle/entity"
	"github.com/streamingfast/sparkle/indexer"
	"github.com/streamingfast/sparkle/metrics"
	"github.com/streamingfast/sparkle/storage"
	"github.com/streamingfast/sparkle/storage/squashable"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

var parallelStepCmd = &cobra.Command{
	Use:   "step",
	Short: "Run a step",
	Args:  cobra.NoArgs,
	RunE:  runParallelStep,
}

func init() {
	parallelStepCmd.Flags().StringP("rpc-endpoint", "e", "http://localhost:8545", "ETH JSON-RPC Endpoint")
	parallelStepCmd.Flags().String("blocks-store-url", "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1", "dfuse Blocks Store URL")
	parallelStepCmd.Flags().IntP("step", "s", 1, "First step in parallel loader")
	parallelStepCmd.Flags().BoolP("flush-entities", "b", false, "Flush entities to 'output-path'")
	parallelStepCmd.Flags().Bool("store-snapshot", true, "Enables snapshot storage in 'output_path' at the end of the batch")
	parallelStepCmd.Flags().Bool("debug-cache", false, "Enables a cache dump after the preload, and before the batch is run in 'tmp/content.json'")
	parallelStepCmd.Flags().Bool("enable-poi", false, "Enable POI injection")
	parallelStepCmd.Flags().Bool("enable-aggregate-snapshot-save", false, "Enable the save of the aggregate snapshot")
	parallelStepCmd.Flags().Bool("non-archive-node", false, "Remove the requirement for an archive node. RPC Calls will be called on LATEST (breaks consistency and POI)")
	parallelStepCmd.Flags().String("rpc-cache-load-path", "", "If non-empty, will load and use rpc request/response cache from this path")
	parallelStepCmd.Flags().String("rpc-cache-save-path", "", "If non-empty, will save rpc request/response cache to this path (can be the same as load-path to append data)")

	parallelCmd.AddCommand(parallelStepCmd)
}

func runParallelStep(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	startBlock := viper.GetUint64("parallel-cmd-start-block")
	stopBlock := viper.GetUint64("parallel-cmd-stop-block")
	rpcEndpoint := viper.GetString("parallel-step-cmd-rpc-endpoint")
	step := viper.GetInt("parallel-step-cmd-step")
	blocksStoreURL := viper.GetString("parallel-step-cmd-blocks-store-url")
	flushEntities := viper.GetBool("parallel-step-cmd-flush-entities")
	outputPath := viper.GetString("parallel-cmd-output-path")
	inputPath := viper.GetString("parallel-cmd-input-path")
	rpcCacheLoadPath := viper.GetString("parallel-step-cmd-rpc-cache-load-path")
	rpcCacheSavePath := viper.GetString("parallel-step-cmd-rpc-cache-save-path")
	entitiesPath := viper.GetString("parallel-step-cmd-entities-path")
	storeSnapshot := viper.GetBool("parallel-step-cmd-store-snapshot")
	debugCache := viper.GetBool("parallel-step-cmd-debug-cache")
	enablePOI := viper.GetBool("parallel-step-cmd-enable-poi")
	enableAggregateSnapshotSave := viper.GetBool("parallel-step-cmd-enable-aggregate-snapshot-save")
	nonArchiveNode := viper.GetBool("parallel-step-cmd-non-archive-node")

	zlog.Info("fetching transactions for network",
		zap.String("rpc_endpoint", rpcEndpoint),
		zap.Uint64("start_block", startBlock),
		zap.Uint64("stop_block", stopBlock),
		zap.String("output_path", outputPath),
		zap.String("input_path", inputPath),
		zap.String("rpc_cache_load_path", rpcCacheLoadPath),
		zap.String("rpc_cache_save_path", rpcCacheSavePath),
		zap.String("entities_path", entitiesPath),
		zap.String("blocks_store_url", blocksStoreURL),
		zap.Int("step", step),
		zap.Bool("flush_entities", flushEntities),
		zap.Bool("store_snapshots", storeSnapshot),
		zap.Bool("debug_cache", debugCache),
		zap.Bool("enable_poi", enablePOI),
		zap.Bool("enable_aggregate_snapshot_save", enableAggregateSnapshotSave),
		zap.Bool("non_archive_node", nonArchiveNode),
	)

	zlog.Info("creating rpc client")
	rpcClient := rpc.NewClient(rpcEndpoint, rpc.WithHttpClient(&http.Client{
		Timeout: 10 * time.Second,
	}))

	var inputStore dstore.Store
	var err error
	if inputPath != "" {
		inputStore, err = dstore.NewStore(inputPath, "", "", true)
		if err != nil {
			return fmt.Errorf("unable to create input store: %w", err)
		}
	}
	outputStore, err := dstore.NewStore(outputPath, "", "", true)
	if err != nil {
		return fmt.Errorf("unable to create output store: %w", err)
	}

	blocksStore, err := dstore.NewDBinStore(blocksStoreURL)
	if err != nil {
		return fmt.Errorf("unable to create blocks store: %w", err)
	}

	subgraphDef := subgraph.MainSubgraphDef

	// TODO: rename to `mergeableStore`. for consistency.
	squashableStore := squashable.New(ctx, zlog, subgraphDef, step, startBlock, stopBlock)

	if flushEntities {
		zlog.Info("defaulting entities store to output store, since 'entities_path' is not defined", zap.String("entities_store_path", outputPath))
		squashableStore.FlushEntities(outputStore)
	}

	if inputStore != nil {
		startBlockNum, stopBlockNum, err := squashableStore.Preload(ctx, inputStore)
		if err != nil {
			return fmt.Errorf("unable to load snapshots: %w", err)
		}

		if enableAggregateSnapshotSave {
			aggregateFilename := fmt.Sprintf("%010d-%010d.jsonl", startBlockNum, stopBlockNum)
			zlog.Info("flushing aggregate snapshot to input path",
				zap.String("input_path", inputPath),
				zap.Int("step", step),
				zap.String("snapshot_filename", aggregateFilename),
			)
			_, err = squashableStore.WriteSnapshot(inputStore, aggregateFilename)
			if err != nil {
				return fmt.Errorf("unable store aggregate snapshot: %w", err)
			}
		}
	}

	if debugCache {
		zlog.Info("dumping Dumping squashed cache content into /tmp/content.json")
		cnt, _ := json.MarshalIndent(squashableStore.GetCache(), "", "  ")
		ioutil.WriteFile("/tmp/content.json", cnt, 0644)
	}

	firehoseFactory := blocks.NewLocalFirehoseFactory(blocksStore)

	sf := func(_ context.Context, _ *zap.Logger, _ *metrics.BlockMetrics, _ *entity.Registry) (storage.Store, error) {
		return squashableStore, nil
	}

	// TODO: this can be stored in the generated subgraph
	manifest, err := manifestlib.DecodeYamlManifest(subgraphDef.Manifest)
	if err != nil {
		return fmt.Errorf("unable to decode manifest: %w", err)
	}

	var indexerOpts []indexer.Option
	if enablePOI {
		indexerOpts = append(indexerOpts, indexer.WithPOI(manifest.Network()))
	}

	if nonArchiveNode {
		indexerOpts = append(indexerOpts, indexer.WithNonArchiveNode())
	}

	var rpcCache *indexer.RPCCache
	if rpcCacheLoadPath != "" || rpcCacheSavePath != "" {
		var loadStore, saveStore dstore.Store
		if rpcCacheSavePath != "" {
			saveStore, err = dstore.NewStore(rpcCacheSavePath, "json.zst", "zstd", true)
			if err != nil {
				return fmt.Errorf("cannot open RPC Cache save store:%w", err)
			}
		}
		if rpcCacheLoadPath != "" {
			loadStore, err = dstore.NewStore(rpcCacheLoadPath, "json.zst", "zstd", false)
			if err != nil {
				return fmt.Errorf("cannot open RPC Cache load store:%w", err)
			}
		}
		rpcCache = indexer.NewCache(loadStore, saveStore, startBlock, stopBlock)

		rpcCache.Load(ctx)
		indexerOpts = append(indexerOpts, indexer.WithRPCCache(rpcCache))
	}

	indexer := indexer.NewBatch(
		step,
		startBlock,
		stopBlock,
		rpcClient,
		firehoseFactory,
		subgraphDef,
		indexerOpts...,
	)

	err = indexer.Start(sf)

	if err != nil {
		return fmt.Errorf("start indexer returned error: %w", err)
	}

	signalHandler := derr.SetupSignalHandler(viper.GetDuration("common-system-shutdown-signal-delay"))
	select {
	case <-signalHandler:
		zlog.Info("Received termination signal, quitting")
		indexer.Shutdown(nil)
	case <-indexer.Terminating():
	}

	zlog.Info("wait for complete shutdown")
	<-indexer.Terminated()
	zlog.Info("terminated")

	if indexer.Err() != nil {
		return indexer.Err()
	}

	if flushEntities {
		zlog.Info("closing entities flushing pipeline")
		if err := squashableStore.CloseEntities(); err != nil {
			return fmt.Errorf("unable to close squasahble store entities writer: %w", err)
		}

	}
	if storeSnapshot {
		snapshotFilename := fmt.Sprintf("%010d-%010d.jsonl", squashableStore.StartBlock, squashableStore.EndBlock)
		zlog.Info("flushing snapshot to output store",
			zap.Int("step", step),
			zap.String("output_path", outputPath),
			zap.String("snapshot_filename", snapshotFilename),
		)
		_, err := squashableStore.WriteSnapshot(outputStore, snapshotFilename)
		if err != nil {
			return err
		}
	}
	if rpcCache != nil {
		zlog.Info("saving rpc cache", zap.String("rpc_cache_save_path", rpcCacheSavePath))
		rpcCache.Save(ctx)
	}

	return nil
}
