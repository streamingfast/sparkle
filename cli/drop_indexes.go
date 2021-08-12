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

var dropIndexesCmd = &cobra.Command{
	Use:   "drop-indexes <subgraph-name>@<version>",
	Short: "Drop subgraph's indexes for a given deployment version",
	Args:  cobra.ExactArgs(1),
	RunE:  runDropIndexes,
}

func init() {
	dropIndexesCmd.Flags().String("psql-dsn", "postgresql://graph:${PG_PASSWORD}@127.0.0.1:5432/graph?enable_incremental_sort=off&sslmode=disable", "Postgres DSN to connect to, ${} variables are expanded against enviornment variables")
	dropIndexesCmd.Flags().StringSlice("only-tables", []string{}, "Drop indexes only the following tables")
	RootCmd.AddCommand(dropIndexesCmd)
}

func runDropIndexes(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	subgraphVersionedName := args[0]

	psqlDSN := viper.GetString("drop-indexes-cmd-psql-dsn")
	onlyTables := viper.GetStringSlice("drop-indexes-cmd-only-tables")

	zlog.Info("dropping indexes for  subgraph",
		zap.String("subgraph_versioned_name", subgraphVersionedName),
		zap.String("psql_dsn", psqlDSN),
		zap.Strings("only_tables", onlyTables),
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

	db, err := createPostgresDB(ctx, postgresDSN)
	if err != nil {
		return fmt.Errorf("creating postgres db: %w", err)
	}

	spec, err := deployment.GetSubgraphVersion(ctx, db, versionedSubgraph.name, versionedSubgraph.version)
	if err != nil {
		return fmt.Errorf("unable to retrieve specs: %q", err)
	}

	zlog.Info("droping indexes", zap.Reflect("subgraph_deployment", spec))
	err = postgres.DropIndexes(ctx, db, subgraph.MainSubgraphDef, spec.Schema, onlyTables, zlog)
	if err != nil {
		return fmt.Errorf("dropping index: %w", err)
	}
	return nil
}
