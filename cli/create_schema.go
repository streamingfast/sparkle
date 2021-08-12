package cli

import (
	"fmt"
	_ "net/http/pprof"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/sparkle/deployment"
	"github.com/streamingfast/sparkle/storage/postgres"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

var createSchemaCmd = &cobra.Command{
	Use:   "create-schema <schema>",
	Short: "Create new schema and tables for a subgraph",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateSchema,
}

func init() {
	createSchemaCmd.Flags().String("psql-dsn", "postgresql://postgres:${PG_PASSWORD}@127.0.0.1:5432/graph-node?enable_incremental_sort=off&sslmode=disable", "Postgres DSN where to connect, expands environment variable in the form '${}'")
	RootCmd.AddCommand(createSchemaCmd)
}

func runCreateSchema(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	zlog.Info("Warning by simply creating the schema, you will not be creating all the associated subgraph objects")
	schema := args[0]
	psqlDSN := viper.GetString("create-schema-cmd-psql-dsn")

	zlog.Info("starting subgraph deploy",
		zap.String("psql_dsn", psqlDSN),
		zap.String("schema", schema),
	)

	postgresDSN, err := postgres.ParseDSN(psqlDSN)
	if err != nil {
		return fmt.Errorf("invalid postgres DSN %q: %w", psqlDSN, err)
	}

	db, err := createPostgresDB(ctx, postgresDSN)
	if err != nil {
		return fmt.Errorf("creating postgres db: %w", err)
	}

	if err = deployment.SetupDBSchema(ctx, db, subgraph.MainSubgraphDef, schema); err != nil {
		return fmt.Errorf("unable to create schema: %w", err)
	}
	zlog.Info("completed schema deployment of subgraph")
	return nil

}
