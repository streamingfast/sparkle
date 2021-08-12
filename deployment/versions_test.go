package deployment

import (
	"context"
	"os"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestGetSubgraphVersion(t *testing.T) {
	ctx := context.Background()
	dsn := os.Getenv("TEST_DEPLOYMENT_DSN")
	if dsn == "" {
		t.Skip("skipping deployment TestGetSubgraphVersion set 'TEST_DEPLOYMENT_DSN' to you psql DSB to run test")
		return
	}
	subgraphName := "pancakeswap/exchange-v2"
	version := "52630e30c4452a700d971ae888373185"

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	require.NoError(t, err)

	db.SetMaxOpenConns(100)

	dep, err := GetSubgraphVersion(ctx, db, subgraphName, version)
	require.NoError(t, err)

	assert.Equal(t, &SubgraphVersion{
		DeploymentID:     "QmQKsZbFbYuQQPMHBzParhUm3Dpbd9T6bwwk4j5znWyMDS",
		SubgraphID:       "1c338a0e7d072e91c044fcee8c8bd7e5",
		VersionID:        "52630e30c4452a700d971ae888373185",
		Schema:           "sgd5",
		IsCurrentVersion: true,
		IsPendingVersion: false,
	}, dep)
}
