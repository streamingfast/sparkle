package cli

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/streamingfast/sparkle/deployment"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/sparkle/storage/postgres"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

var createCmd = &cobra.Command{
	Use:   "create <subgraph-name>",
	Short: "create a new subgraph",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().String("psql-dsn", "postgresql://postgres:${PG_PASSWORD}@127.0.0.1:5432/graph-node?enable_incremental_sort=off&sslmode=disable", "Postgres DSN where to connect, expands environment variable in the form '${}'")
	RootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	subgraphName := args[0]
	psqlDSN := viper.GetString("create-cmd-psql-dsn")

	zlog.Info("starting subgraph create",
		zap.String("psql_dsn", psqlDSN),
		zap.String("subgraph_name", subgraphName),
	)

	postgresDSN, err := postgres.ParseDSN(psqlDSN)
	if err != nil {
		return fmt.Errorf("invalid postgres DSN %q: %w", psqlDSN, err)
	}

	db, err := createPostgresDB(ctx, postgresDSN)
	if err != nil {
		return fmt.Errorf("creating postgres db: %w", err)
	}

	if err := deployment.CreateSubgraph(ctx, db, subgraph.MainSubgraphDef, subgraphName); err != nil {
		return fmt.Errorf("unable to create subgraph: %w", err)
	}
	// TODO: better user error handling
	userLog.Printf("Subgraph %s created successfully!", subgraphName)
	return nil

}
