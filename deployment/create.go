package deployment

import (
	"context"
	"fmt"
	"time"

	manifestlib "github.com/streamingfast/sparkle/manifest"

	"github.com/jmoiron/sqlx"
	"github.com/streamingfast/sparkle/subgraph"
)

func CreateSubgraph(ctx context.Context, db *sqlx.DB, subgraphDef *subgraph.Definition, subgraphName string) error {
	manifest, err := manifestlib.DecodeYamlManifest(subgraphDef.Manifest)
	if err != nil {
		return fmt.Errorf("unable to decode manifest from subgraph definition: %w", err)
	}
	network := manifest.Network()
	ok, err := isChainSupported(ctx, db, network)
	if err != nil {
		return fmt.Errorf("unable to retrieve supported chaing: %w", err)
	}
	if !ok {
		return fmt.Errorf("chain %q is not suporter", network)
	}

	subgraphId := generateID()
	sql := fmt.Sprintf(
		"INSERT INTO subgraphs.subgraph (id, name, current_version, pending_version, created_at, vid, block_range) VALUES ('%s', '%s', null, null, %d, DEFAULT, '[-1,)')",
		subgraphId,
		subgraphName,
		time.Now().UTC().Unix(),
	)

	_, err = db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("unable to create subgraph: %w", err)
	}

	return nil
}

func isChainSupported(ctx context.Context, db *sqlx.DB, network string) (bool, error) {
	sql := "SELECT name FROM public.chains WHERE name=$1"
	row := struct {
		Name string
	}{}
	if err := db.GetContext(ctx, &row, sql, network); err != nil {
		return false, fmt.Errorf("unable to retrieve supported chain: %w", err)
	}

	if row.Name == network {
		return true, nil
	}
	return false, nil
}
