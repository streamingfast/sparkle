package cli

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	manifestlib "github.com/streamingfast/sparkle/manifest"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dmetrics"
	"github.com/streamingfast/eth-go/rpc"
	"github.com/streamingfast/sparkle/blocks"
	"github.com/streamingfast/sparkle/deployment"
	"github.com/streamingfast/sparkle/entity"
	"github.com/streamingfast/sparkle/indexer"
	"github.com/streamingfast/sparkle/metrics"
	"github.com/streamingfast/sparkle/storage"
	"github.com/streamingfast/sparkle/storage/dryrunstore"
	"github.com/streamingfast/sparkle/storage/postgres"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

var indexCmd = &cobra.Command{
	Use:   "index <subgraph-name>@<version>",
	Short: "Starts the indexer for <subgraph-name>'s deployment version <version>",
	Args:  cobra.ExactArgs(1),
	Example: ExampleSparkle(`
		index exchange@52630e30c4452a700d971ae888373185
	`),
	RunE: runIndex,
}

func init() {
	indexCmd.Flags().String("psql-dsn", "postgresql://postgres:${PG_PASSWORD}@127.0.0.1:5432/graph-node?enable_incremental_sort=off&sslmode=disable", "Postgres DSN where to connect, expands environment variable in the form '${}'")
	indexCmd.Flags().String("rpc-endpoint", "http://localhost:8545", "RPC endpoint to use to perform Ethereum JSON-RPC.")
	indexCmd.Flags().Int64("start-block-num", 0, "start-block-num")
	indexCmd.Flags().String("sf-api-key", "", "StreamingFast API key")
	indexCmd.Flags().String("sf-endpoint", "bsc.streamingfast.io:443", "StreamingFast API endpoint")
	indexCmd.Flags().Bool("with-reversible", false, "get reversible block segments from stream")
	indexCmd.Flags().Bool("flush-without-transaction", false, "skip DB transactions when flushing to improve speed")
	indexCmd.Flags().Bool("dry-run", false, "Dry run and spit out as .csv files. Implies --with-reversible=false")
	indexCmd.Flags().Int64("dry-run-blocks", 1000, "Number of blocks for a dry-run")
	indexCmd.Flags().String("dry-run-output", "./dry_run", "Path to output dry-run CSVs")
	indexCmd.Flags().String("schema-override", "", "Override schema")
	indexCmd.Flags().String("pprof-listen-addr", ":6060", "If non-empty, the process will listen on this address for pprof analysis (see https://golang.org/pkg/net/http/pprof/)")
	indexCmd.Flags().Bool("enable-poi", false, "Enable POI injection")
	indexCmd.Flags().Bool("non-archive-node", false, "Remove the requirement for an archive node. RPC Calls will be called on LATEST (breaks consistency and POI)")
	RootCmd.AddCommand(indexCmd)
}

func runIndex(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	subgraphVersionedName := args[0]

	psqlDSN := viper.GetString("index-cmd-psql-dsn")
	rpcEndpoint := viper.GetString("index-cmd-rpc-endpoint")
	startBlockNum := viper.GetInt64("index-cmd-start-block-num")
	sfAPIKey := viper.GetString("index-cmd-sf-api-key")
	sfEndpoint := viper.GetString("index-cmd-sf-endpoint")
	flushWithoutTransaction := viper.GetBool("index-cmd-flush-without-transaction")
	withReversible := viper.GetBool("index-cmd-with-reversible")
	dryRun := viper.GetBool("index-cmd-dry-run")
	dryRunBlocks := viper.GetInt64("index-cmd-dry-run-blocks")
	dryRunOutput := viper.GetString("index-cmd-dry-run-output")
	pprofListenAddr := viper.GetString("index-cmd-pprof-listen-addr")
	enablePOI := viper.GetBool("index-cmd-enable-poi")
	nonArchiveNode := viper.GetBool("index-cmd-non-archive-node")

	if dryRun {
		withReversible = false
	}

	zlog.Info("starting sparkle graph indexer",
		zap.String("psql_dsn", psqlDSN),
		zap.String("rpc_endpoint", rpcEndpoint),
		zap.Int64("start_block_num", startBlockNum),
		zap.String("streaming_fast_api_key", sfAPIKey),
		zap.String("streaming_fast_api_endpoint", sfEndpoint),
		zap.Bool("flush_without_transaction", flushWithoutTransaction),
		zap.String("subgraph_versioned_name", subgraphVersionedName),
		zap.Bool("with_reversible", withReversible),
		zap.Bool("dry_run", dryRun),
		zap.Int64("dry_run_blocks", dryRunBlocks),
		zap.String("dry_run_output", dryRunOutput),
		zap.Bool("enable_poi", enablePOI),
		zap.Bool("non_archive_node", nonArchiveNode),
	)

	versionedSubgraph, err := parseSubgraphVersionedName(subgraphVersionedName)
	if err != nil {
		return fmt.Errorf("unable to parse subgraph versioned name %q: %w", subgraphVersionedName, err)
	}

	subgraphDef := subgraph.MainSubgraphDef

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true, // don't reuse connections
		},
		Timeout: 3 * time.Second,
	}

	rpcClient := rpc.NewClient(rpcEndpoint, rpc.WithHttpClient(httpClient))

	apiKey := sfAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("STREAMINGFAST_API_KEY")
	}

	if apiKey == "" {
		return fmt.Errorf("pass an API key as `STREAMINGFAST_API_KEY` env var, or with --api-key")
	}

	firehoseFactory, err := blocks.NewStreamingFastFirehoseFactory(apiKey, sfEndpoint)
	if err != nil {
		return fmt.Errorf("unable to create firehose factory: %w", err)
	}

	var indexerOpts []indexer.Option
	if !flushWithoutTransaction {
		indexerOpts = append(indexerOpts, indexer.UseTransactionalFlush())
	}

	if startBlockNum != 0 {
		indexerOpts = append(indexerOpts, indexer.StartBlock(startBlockNum))
	}

	if dryRun {
		indexerOpts = append(indexerOpts, indexer.StopBlock(subgraphDef.StartBlock+uint64(dryRunBlocks)))
	}

	if withReversible {
		indexerOpts = append(indexerOpts, indexer.WithReversible())
	}

	// TODO: this can be stored in the generated subgraph
	manifest, err := manifestlib.DecodeYamlManifest(subgraphDef.Manifest)
	if err != nil {
		return fmt.Errorf("unable to decode manifest")
	}

	if enablePOI {
		indexerOpts = append(indexerOpts, indexer.WithPOI(manifest.Network()))
	}

	if nonArchiveNode {
		indexerOpts = append(indexerOpts, indexer.WithNonArchiveNode())
	}

	indexerInst := indexer.New(rpcClient, firehoseFactory, subgraphDef, indexerOpts...)

	zlog.Info("launching subgraph indexing",
		zap.Uint64("subgrpah_start_block", subgraphDef.StartBlock),
		zap.String("subgraph_name", versionedSubgraph.name),
		zap.String("subgraph_version", versionedSubgraph.version),
	)

	var storeFactory indexer.StoreFactory

	if dryRun {
		zlog.Info("setting up dry run store factory")
		storeFactory = func(streamCtx context.Context, logger *zap.Logger, metrics *metrics.BlockMetrics, registry *entity.Registry) (storage.Store, error) {
			drStore := dryrunstore.New(ctx, subgraphDef, logger, subgraphDef.StartBlock+uint64(dryRunBlocks), dryRunOutput)

			if err := drStore.OpenOutputFiles(); err != nil {
				return nil, fmt.Errorf("opening up dry run store: %w", err)
			}

			return drStore, nil
		}
	} else {
		zlog.Info("setting up postgre store factory")
		storeFactory = func(streamCtx context.Context, logger *zap.Logger, metrics *metrics.BlockMetrics, registry *entity.Registry) (storage.Store, error) {

			postgresDSN, err := postgres.ParseDSN(psqlDSN)
			if err != nil {
				return nil, fmt.Errorf("invalid postgres DSN %q: %w", postgresDSN, err)
			}

			db, err := createPostgresDB(ctx, postgresDSN)
			if err != nil {
				return nil, fmt.Errorf("creating postgres db: %w", err)
			}

			spec, err := deployment.GetSubgraphVersion(ctx, db, versionedSubgraph.name, versionedSubgraph.version)
			if err != nil {
				return nil, err
			}
			zlog.Info("specs resolved", zap.Reflect("spec", spec))

			zlog.Info("creating postgres store")
			psqlStore, err := postgres.New(logger, metrics, db, spec.Schema, spec.DeploymentID, subgraphDef, map[string]bool{}, true)
			if err != nil {
				return nil, err
			}
			zlog.Info("postgres store created")

			zlog.Info("registering entities")
			if err := psqlStore.RegisterEntities(); err != nil {
				return nil, fmt.Errorf("unable to start store: %w", err)
			}
			zlog.Info("entities registered")

			psqlStore.StartLogger(streamCtx)
			return psqlStore, nil
		}
	}

	go dmetrics.Serve(":9102")

	go func() {
		err := http.ListenAndServe(pprofListenAddr, nil)
		if err != nil {
			zlog.Info("unable to start profiling server", zap.Error(err), zap.String("listen_addr", pprofListenAddr))
		}
	}()

	zlog.Info("starting indexer")
	err = indexerInst.Start(storeFactory)
	if err != nil {
		return fmt.Errorf("starting indexer: %w", err)
	}
	zlog.Info("indexer started")

	signalHandler := derr.SetupSignalHandler(viper.GetDuration("common-system-shutdown-signal-delay"))
	select {
	case <-signalHandler:
		zlog.Info("received termination signal, shutting down indexer")
		indexerInst.Shutdown(nil)
	case <-indexerInst.Terminating():
		zlog.Info("indexer is terminating, quitting")
	}

	zlog.Info("wait for indexer to completely shutdown")
	<-indexerInst.Terminated()
	zlog.Info("terminated")
	return nil
}
