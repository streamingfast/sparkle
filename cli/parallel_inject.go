package cli

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/streamingfast/sparkle/deployment"

	"github.com/abourget/llerrgroup"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dstore"
	"github.com/streamingfast/sparkle/entity"
	"github.com/streamingfast/sparkle/storage/postgres"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

var parallelInjectCmd = &cobra.Command{
	Use:   "inject <subgraph-name>@<version> <psql-dsn>",
	Short: "Injects generated CSV entities for <subgraph-name>'s deployment version <version> into the database pointed by <psql-dsn> argument",
	Args:  cobra.ExactArgs(2),
	RunE:  runParallelInject,
}

func init() {
	parallelInjectCmd.Flags().StringSlice("only-tables", []string{}, "Inject only the following tables")
	parallelInjectCmd.Flags().Bool("enable-index-creation", false, "Create indexes after injection completed")

	parallelCmd.AddCommand(parallelInjectCmd)
}

func runParallelInject(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	subgraphVersionedName := args[0]
	psqlDSN := args[1]

	inputPath := viper.GetString("parallel-cmd-input-path")
	startBlockNum := viper.GetUint64("parallel-cmd-start-block")
	stopBlockNum := viper.GetUint64("parallel-cmd-stop-block")
	onlyTables := viper.GetStringSlice("parallel-inject-cmd-only-tables")
	enableIndexesCreation := viper.GetBool("parallel-inject-cmd-enable-index-creation")
	onlyTablesMap := map[string]bool{}

	zlog.Info("injecting into postgres",
		zap.String("subgraph_versioned_name", subgraphVersionedName),
		zap.String("psql_dsn", psqlDSN),
		zap.String("input_path", inputPath),
		zap.Strings("only_tables", onlyTables),
		zap.Uint64("start_block_num", startBlockNum),
		zap.Uint64("stop_block_num", stopBlockNum),
		zap.Bool("enable_indexes_creation", enableIndexesCreation),
	)

	versionedSubgraph, err := parseSubgraphVersionedName(subgraphVersionedName)
	if err != nil {
		return fmt.Errorf("unable to parse subgraph versioned name %q: %w", subgraphVersionedName, err)
	}
	zlog.Info("injecting subgraph",
		zap.String("subgraph_name", versionedSubgraph.name),
		zap.String("version", versionedSubgraph.version),
	)

	postgresDSN, err := postgres.ParseDSN(psqlDSN)
	if err != nil {
		return fmt.Errorf("invalid postgres DSN %q: %w", psqlDSN, err)
	}

	zlog.Info("connecting to input store")
	inputStore, err := dstore.NewStore(inputPath, "", "", false)
	if err != nil {
		return fmt.Errorf("unable to create input store: %w", err)
	}

	subgraphDef := subgraph.MainSubgraphDef

	entityCount := subgraphDef.Entities.Len()
	sqlxDB, err := createPostgresDB(ctx, postgresDSN)
	if err != nil {
		return fmt.Errorf("creating postgres db: %w", err)
	}

	zlog.Info("connecting to postgres")
	pool, err := pgxpool.Connect(ctx, fmt.Sprintf("%s pool_min_conns=%d pool_max_conns=%d", postgresDSN.DSN(), entityCount+1, entityCount+2))
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquiring pg connection: %w", err)
	}

	specs, err := deployment.GetSubgraphVersion(ctx, sqlxDB, versionedSubgraph.name, versionedSubgraph.version)
	if err != nil {
		return fmt.Errorf("unable to retrieve specs: %q", err)
	}

	for _, tbl := range onlyTables {
		onlyTablesMap[tbl] = true
	}

	zlog.Info("creating progress table ")
	// TODO: handler error
	_, _ = conn.Exec(ctx, fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.progress$ (filename text PRIMARY KEY, table_name text, injected_at timestamp without time zone NOT NULL);`, specs.Schema))
	conn.Release()

	eg := llerrgroup.New(20)
	t0 := time.Now()

	for tableName, ent := range subgraphDef.Entities.Data() {
		if eg.Stop() {
			continue // short-circuit the loop if we got an error
		}

		if len(onlyTablesMap) != 0 && !onlyTablesMap[tableName] {
			continue
		}

		tableShard := NewTblShard(pool, specs.Schema, tableName, ent, startBlockNum, stopBlockNum, inputStore)
		theTableName := tableName
		eg.Go(func() error {
			err = tableShard.Run(ctx)
			if err != nil {
				return fmt.Errorf("table shard %q: %w", theTableName, err)
			}
			if enableIndexesCreation {
				zlog.Info("creating indexes")
				return postgres.CreateIndexes(ctx, sqlxDB, subgraphDef, specs.Schema, []string{tableName}, zlog)
			}
			return nil

		})
	}

	zlog.Info("waiting for all shards")
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("processEntity shards: %w", err)
	}

	zlog.Info("all shards done", zap.Duration("total", time.Since(t0)))
	// for tbl, fields := range shardStats {
	// 	zlog.Info(fmt.Sprintf("shard completed - %s", tbl), fields...)
	// }
	return nil
}

type ProgressMarker struct {
	Filename   string    `db:"filename"`
	TableName  string    `db:"table_name"`
	InjectedAt time.Time `db:"injected_at"`
}

type TblShard struct {
	pqSchema string
	tblName  string

	ent           reflect.Type
	in            dstore.Store
	startBlockNum uint64
	stopBlockNum  uint64
	pool          *pgxpool.Pool
}

func NewTblShard(pool *pgxpool.Pool, pqSchema, tblName string, ent reflect.Type, startBlockNum, stopBlockNum uint64, inStore dstore.Store) *TblShard {
	return &TblShard{
		tblName:       tblName,
		pqSchema:      pqSchema,
		pool:          pool,
		ent:           ent,
		startBlockNum: startBlockNum,
		stopBlockNum:  stopBlockNum,
		in:            inStore,
	}
}

type sortedFilenames []string

func (p sortedFilenames) Len() int           { return len(p) }
func (p sortedFilenames) Less(i, j int) bool { return p[i] > p[j] }
func (p sortedFilenames) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (t *TblShard) pruneFiles(ctx context.Context, filenames []string) (out sortedFilenames, err error) {
	conn, err := t.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to acquire conn: %w", err)
	}
	defer conn.Release()
	filesToPrune := map[string]bool{}
	for _, filename := range filenames {
		filesToPrune[filename] = true
	}

	result, err := conn.Query(ctx, fmt.Sprintf("SELECT * FROM %s.progress$ WHERE (table_name=$1)", t.pqSchema), t.tblName)
	if err != nil {
		return nil, err
	}

	for result.Next() {
		values, err := result.Values()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve result row: %w", err)
		}
		injectedFilename := values[0].(string)
		if _, found := filesToPrune[injectedFilename]; found {
			delete(filesToPrune, injectedFilename)
		}
	}
	for filename := range filesToPrune {
		out = append(out, filename)
	}
	sort.Sort(out)
	return out, nil
}

func (t *TblShard) markProgress(ctx context.Context, filename string, timestamp time.Time) error {
	conn, err := t.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("unable to acquire conn: %w", err)
	}
	defer conn.Release()

	query := fmt.Sprintf("INSERT INTO %s.progress$ (filename, table_name, injected_at) VALUES ($1, $2, $3)", t.pqSchema)
	_, err = conn.Exec(ctx, query, filename, t.tblName, timestamp)
	if err != nil {
		return err
	}
	return nil
}

func s(str string) *string {
	return &str
}

func (t *TblShard) Run(ctx context.Context) error {
	zlog.Info("table shard", zap.String("table", t.tblName))

	dbFields := []string{}
	nonNullableFields := []string{}
	for _, fieldTag := range entity.DBFields(t.ent) {
		if fieldTag.Name == "VID" {
			continue
		}
		dbFields = append(dbFields, fieldTag.ColumnName)

		if !fieldTag.Optional {
			nonNullableFields = append(nonNullableFields, fieldTag.ColumnName)
		}
	}

	loadFiles, err := injectFilesToLoad(t.in, t.tblName, t.stopBlockNum, t.startBlockNum)
	if err != nil {
		return fmt.Errorf("listing files: %w", err)
	}

	prunedFilenames, err := t.pruneFiles(ctx, loadFiles)
	if err != nil {
		return fmt.Errorf("unable to prune filename list: %w", err)
	}

	zlog.Info("files to load",
		zap.String("table", t.tblName),
		zap.Int("file_count", len(loadFiles)),
		zap.Int("pruned_file_count", len(prunedFilenames)),
	)

	for _, filename := range prunedFilenames {
		zlog.Info("opening file", zap.String("file", filename))

		if err := t.injectFile(ctx, filename, dbFields, nonNullableFields); err != nil {
			return fmt.Errorf("failed to inject file %q: %w", filename, err)
		}

		if err := t.markProgress(ctx, filename, time.Now()); err != nil {
			return fmt.Errorf("failed to mark progress file %q: %w", filename, err)
		}
	}

	return nil
}

func (t *TblShard) injectFile(ctx context.Context, filename string, dbFields, nonNullableFields []string) error {
	fl, err := t.in.OpenObject(ctx, filename)
	if err != nil {
		return fmt.Errorf("opening csv: %w", err)
	}
	defer fl.Close()

	query := fmt.Sprintf(`COPY %s.%s ("%s") FROM STDIN WITH (FORMAT CSV, FORCE_NOT_NULL ("%s"))`, t.pqSchema, t.tblName, strings.Join(dbFields, `","`), strings.Join(nonNullableFields, `","`))
	zlog.Info("loading file into sql", zap.String("filename", filename), zap.String("table_name", t.tblName), zap.Strings("db_fields", dbFields), zap.Strings("non_nullable_fields", nonNullableFields))

	t0 := time.Now()

	conn, err := t.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("pool acquire: %w", err)
	}
	defer conn.Release()

	tag, err := conn.Conn().PgConn().CopyFrom(ctx, fl, query)
	if err != nil {
		return fmt.Errorf("failed COPY FROM for %q: %w", t.tblName, err)
	}
	count := tag.RowsAffected()
	elapsed := time.Since(t0)
	zlog.Info("loaded file into sql",
		zap.String("filename", filename),
		zap.String("table_name", t.tblName),
		zap.Int64("rows_affected", count),
		zap.Duration("elapsed", elapsed),
	)

	return nil
}

func injectFilesToLoad(inputStore dstore.Store, tableName string, stopBlockNum, desiredStartBlockNum uint64) (out []string, err error) {
	err = inputStore.Walk(context.Background(), tableName+"/", func(filename string) (err error) {
		startBlockNum, _, err := getBlockRange(filename)
		if err != nil {
			return fmt.Errorf("fail reading block range in %q: %w", filename, err)
		}

		if stopBlockNum != 0 && startBlockNum >= stopBlockNum {
			return dstore.StopIteration
		}

		if startBlockNum < desiredStartBlockNum {
			return nil
		}

		if strings.Contains(filename, ".csv") {
			out = append(out, filename)
		}

		return nil
	})
	return
}
