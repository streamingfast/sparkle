package cli

import (
	"fmt"

	"github.com/streamingfast/sparkle/deployment"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/sparkle/storage/postgres"
	"go.uber.org/zap"
)

// new `indexes` command with `--drop` and `--create`.
// drops can be linear, but creates need to be linear
// takes as input the `backup.sql` file.. and FILTERS OUT
// only the things we need: INDEXES, and CONSTRAINTS.
// We don't want to touch to tabels or sequences.
// Parse the `backup.sql`, extract what you need.
// For each table ?  For each index, without regards to the tables?
//  Take a connection from the `pgxpool` and run those index creation.s.. up to a max of whatever postgres can saturate.

// For --drop, we take the CREATE INDEXES, and regexp the table name out of it, and `DROP IF EXISTS`
// For --create, we just insert `IF NOT EXISTS` in the CREATE INDEX and CREATE CONSTRAINT statement.

var versionsCmd = &cobra.Command{
	Use:   "versions <subgraph-name>",
	Short: "Listed the versions for a given <subgraph-name>",
	Args:  cobra.ExactArgs(1),
	RunE:  runListVersionsE,
}

func init() {
	versionsCmd.Flags().String("psql-dsn", "postgresql://graph:${PG_PASSWORD}@127.0.0.1:5432/graph?enable_incremental_sort=off&sslmode=disable", "Postgres DSN to connect to, ${} variables are expanded against enviornment variables")

	RootCmd.AddCommand(versionsCmd)
}

func runListVersionsE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	subgraphName := args[0]

	psqlDSN := viper.GetString("versions-cmd-psql-dsn")

	zlog.Info("list versions",
		zap.String("subgraph_name", subgraphName),
		zap.String("psql_dsn", psqlDSN),
	)

	postgresDSN, err := postgres.ParseDSN(psqlDSN)
	if err != nil {
		return fmt.Errorf("invalid postgres DSN %q: %w", psqlDSN, err)
	}

	db, err := createPostgresDB(ctx, postgresDSN)
	if err != nil {
		return fmt.Errorf("creating postgres db: %w", err)
	}

	versions, err := deployment.GetSubgraphVersions(ctx, db, subgraphName)
	if err != nil {
		return fmt.Errorf("unable to retrieve subgraph verions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Printf("No available versions for subgraph %q, make sure you ran deploy at-least once.", subgraphName)
		return nil
	}

	fmt.Printf("Subgraph %q available versions:\n", subgraphName)
	for _, version := range versions {
		tag := ""
		if version.IsCurrentVersion {
			tag = "[current]"
		} else if version.IsPendingVersion {
			tag = "[pending]"
		}
		fmt.Printf("	* %s (schema: %s, depoyment_id: %s) %s\n", version.VersionID, version.Schema, version.DeploymentID, tag)
	}
	return nil
}
