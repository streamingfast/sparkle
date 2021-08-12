package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/sparkle/deployment"
	"github.com/streamingfast/sparkle/storage/postgres"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

var createIndexesCmd = &cobra.Command{
	Use:   "create-indexes <subgraph-name>@<version>",
	Short: "Create subgraph's indexes for a given deployment version",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateIndexes,
}

func init() {
	createIndexesCmd.Flags().String("psql-dsn", "postgresql://graph:${PG_PASSWORD}@127.0.0.1:5432/graph?enable_incremental_sort=off&sslmode=disable", "Postgres DSN to connect to, ${} variables are expanded against enviornment variables")
	createIndexesCmd.Flags().StringSlice("only-tables", []string{}, "Create indexes only the following tables")

	RootCmd.AddCommand(createIndexesCmd)
}

func runCreateIndexes(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	subgraphVersionedName := args[0]

	psqlDSN := viper.GetString("create-indexes-cmd-psql-dsn")
	onlyTables := viper.GetStringSlice("create-indexes-cmd-only-tables")

	zlog.Info("creating indexes",
		zap.String("subgraph_versioned_name", subgraphVersionedName),
		zap.String("psql_dsn", psqlDSN),
		zap.Strings("only_tables", onlyTables),
	)

	versionedSubgraph, err := parseSubgraphVersionedName(subgraphVersionedName)
	if err != nil {
		return fmt.Errorf("unable to parse subgraph versioned name %q: %w", subgraphVersionedName, err)
	}

	postgresDSN, err := postgres.ParseDSN(psqlDSN)
	if err != nil {
		return fmt.Errorf("invalid postgres DSN %q: %w", psqlDSN, err)
	}

	db, err := createPostgresDB(ctx, postgresDSN)
	if err != nil {
		return fmt.Errorf("creating postgres db: %w", err)
	}

	spec, err := deployment.GetSubgraphVersion(ctx, db, versionedSubgraph.name, versionedSubgraph.version)
	if err != nil {
		return fmt.Errorf("unable to retrieve specs: %q", err)
	}

	zlog.Info("creating indexes", zap.Reflect("subgraph_deployment", versionedSubgraph))
	err = postgres.CreateIndexes(ctx, db, subgraph.MainSubgraphDef, spec.Schema, onlyTables, zlog)
	if err != nil {
		return fmt.Errorf("creating index: %w", err)
	}
	return nil
}
