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

var deployCmd = &cobra.Command{
	Use:   "deploy <subgraph-name>",
	Short: "Deploy a new subgraph on the given database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeploy,
}

func init() {
	deployCmd.Flags().String("psql-dsn", "postgresql://postgres:${PG_PASSWORD}@127.0.0.1:5432/graph-node?enable_incremental_sort=off&sslmode=disable", "Postgres DSN where to connect, expands environment variable in the form '${}'")
	deployCmd.Flags().String("ipfs-address", "http://localhost:5001", "IPFS server to upload manfiest")
	RootCmd.AddCommand(deployCmd)
}

func runDeploy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	subgraphName := args[0]
	psqlDSN := viper.GetString("deploy-cmd-psql-dsn")
	ipfsAddr := viper.GetString("deploy-cmd-ipfs-address")

	zlog.Info("starting subgraph deploy",
		zap.String("psql_dsn", psqlDSN),
		zap.String("subgraph_name", subgraphName),
		zap.String("ipfs_address", ipfsAddr),
	)

	postgresDSN, err := postgres.ParseDSN(psqlDSN)
	if err != nil {
		return fmt.Errorf("invalid postgres DSN %q: %w", psqlDSN, err)
	}

	db, err := createPostgresDB(ctx, postgresDSN)
	if err != nil {
		return fmt.Errorf("creating postgres db: %w", err)
	}

	ipfsNode := deployment.NewIPFSNode(ipfsAddr)

	if err := deployment.DeploySubgraph(ctx, db, subgraph.MainSubgraphDef, ipfsNode, subgraphName); err != nil {
		return fmt.Errorf("deploying subgraph: %w", err)
	}

	zlog.Info("completed deployment of subgraph")
	return nil

}
