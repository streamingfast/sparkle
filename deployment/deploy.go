package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	manifestlib "github.com/streamingfast/sparkle/manifest"
	"github.com/streamingfast/sparkle/storage/postgres"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

func DeploySubgraph(ctx context.Context, db *sqlx.DB, subgraphDef *subgraph.Definition, ipfsNode *IPFSNode, subgraphName string) error {
	zlog.Info("looking up subgraph", zap.String("name", subgraphName))
	subgraph := struct{ ID string }{}

	query := "SELECT id FROM subgraphs.subgraph WHERE name=$1 LIMIT 1"
	if err := db.GetContext(ctx, &subgraph, query, subgraphName); err != nil {
		return fmt.Errorf("getting row from subgraphs.subgraph sql: %w", err)
	}
	zlog.Info("found subgraph", zap.String("subgraph_id", subgraph.ID))

	zlog.Info("parsing manifest")
	manifest, err := manifestlib.DecodeYamlManifest(subgraphDef.Manifest)
	if err != nil {
		return fmt.Errorf("unable to decode manifest from subgraph definition: %w", err)
	}

	zlog.Info("uploading maniest to IPFS")
	ipfsHash, err := ipfsNode.UploadManifest(subgraphDef)
	if err != nil {
		return fmt.Errorf("uploading to ipfs: %w", err)
	}
	deployemntId := ipfsHash

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to begin transaction")
	}

	rollbackFunc := func(err error) {
		zlog.Info("rolling back db", zap.String("cause", err.Error()))
		if err := tx.Rollback(); err != nil {
			panic("failed to rollback transaction")
		}
	}

	zlog.Info("deploying to schema", zap.String("deployment_id", deployemntId))
	deploymentSchema := struct {
		Id     int64  `db:"id"`
		Schema string `db:"name"`
	}{}

	query = "INSERT INTO public.deployment_schemas (subgraph, version, shard, network, active) VALUES ($1, 'relational', 'primary', $2, true) returning id, name"
	res := tx.QueryRowContext(ctx, query, deployemntId, manifest.Network())
	if res.Err() != nil {
		rollbackFunc(res.Err())
		return fmt.Errorf("inserting into public.deployment_schemas sql: %w", err)
	}
	if err = res.Scan(&deploymentSchema.Id, &deploymentSchema.Schema); err != nil {
		rollbackFunc(err)
		return fmt.Errorf("unable to retrieve newly created deployment schema: %w", err)
	}

	zlog.Info("deployment schema created",
		zap.Int64("deployment_schema_id", deploymentSchema.Id),
		zap.String("deployment_schema", deploymentSchema.Schema),
	)

	query = "INSERT INTO subgraphs.subgraph_deployment (deployment, failed, synced, earliest_ethereum_block_hash, earliest_ethereum_block_number, latest_ethereum_block_hash, latest_ethereum_block_number, entity_count, graft_base, graft_block_hash, graft_block_number, fatal_error, non_fatal_errors, health, reorg_count, current_reorg_depth, max_reorg_depth, last_healthy_ethereum_block_hash, last_healthy_ethereum_block_number, id) VALUES ($1, false, false, null, null, null, null, 0, null, null, null, null, DEFAULT, 'healthy', DEFAULT, DEFAULT, DEFAULT, null, null, $2)"
	_, err = tx.ExecContext(ctx, query, ipfsHash, deploymentSchema.Id)
	if err != nil {
		rollbackFunc(err)
		return fmt.Errorf("inserting into public.subgraph_deployment sql: %w", err)
	}

	versionIds, err := GetRandomNames(1)
	if err != nil {
		rollbackFunc(err)
		return fmt.Errorf("generating version id: %w", err)
	}
	versionId := versionIds[0]
	zlog.Info("inserting deployment version", zap.String("version_id", versionId))
	query = "INSERT INTO subgraphs.subgraph_version (id, subgraph, deployment, created_at, block_range) VALUES ($1, $2, $3, $4, '[-1,)')"
	_, err = tx.ExecContext(ctx, query, versionId, subgraph.ID, ipfsHash, time.Now().UTC().Unix())
	if err != nil {
		rollbackFunc(err)
		return fmt.Errorf("inserting into public.subgraph_version sql: %w", err)
	}

	// TODO: check this is correct
	query = "INSERT INTO subgraphs.subgraph_deployment_assignment (node_id, id) VALUES ('default', $1)"
	_, err = tx.ExecContext(ctx, query, deploymentSchema.Id)
	if err != nil {
		rollbackFunc(err)
		return fmt.Errorf("inserting into public.subgraph_deployment_assignment sql: %w", err)
	}

	query = "INSERT INTO subgraphs.subgraph_manifest (spec_version, description, repository, schema, features, id) VALUES ($1, $2, $3, $4, '{}', $5)"
	_, err = tx.ExecContext(ctx, query, manifest.SpecVersion, manifest.Description, manifest.Repository, subgraphDef.GraphQLSchema, deploymentSchema.Id)
	if err != nil {
		rollbackFunc(err)
		return fmt.Errorf("inserting into subgraphs.subgraph_manifest sql: %w", err)
	}

	// THIS IS DANGEROUS, we want the graph-node to figure this out
	//query = "UPDATE subgraphs.subgraph SET current_version = $1 WHERE id = $2"
	//_, err = tx.ExecContext(ctx, query, versionId, subgraph.Id)
	//if err != nil {
	//	rollbackFunc(err)
	//	return fmt.Errorf("inserting into subgraphs.subgraph_manifest sql: %w", err)
	//}

	if err := SetupDBSchema(ctx, db, subgraphDef, deploymentSchema.Schema); err != nil {
		rollbackFunc(err)
		return fmt.Errorf("unable to setup schema db: %w", err)
	}

	return tx.Commit()
}

func SetupDBSchema(ctx context.Context, db *sqlx.DB, subgraphDef *subgraph.Definition, schema string) error {
	if err := postgres.InitiateSchema(ctx, db, subgraphDef, schema, zlog); err != nil {
		return fmt.Errorf("initiating schema: %w", err)
	}

	if err := postgres.CreateTables(ctx, db, subgraphDef, schema, zlog); err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}

	//if err := postgres.CreateIndexes(ctx, db, subgraphDef, schema, nil, zlog); err != nil {
	//	return fmt.Errorf("creating indexes: %w", err)
	//}
	return nil
}
