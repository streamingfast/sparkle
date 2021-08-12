package deployment

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func GetSubgraphVersions(ctx context.Context, db *sqlx.DB, subgraphName string) ([]*SubgraphVersion, error) {
	query := `
	SELECT
	subgraphs.subgraph.id AS subgraph_id,
	subgraphs.subgraph_version.id AS version_id,
	subgraphs.subgraph_version.deployment AS deployment_id,
	CASE (subgraphs.subgraph_version.id = subgraphs.subgraph.current_version) WHEN true THEN true ELSE false END AS is_current_version,
	CASE (subgraphs.subgraph_version.id = subgraphs.subgraph.pending_version)  WHEN true then true ELSE false END AS is_pending_version,
	public.deployment_schemas.name AS schema
 	FROM subgraphs.subgraph
 	LEFT JOIN subgraphs.subgraph_version ON (subgraph_version.subgraph = subgraph.id)
 	LEFT JOIN public.deployment_schemas ON (deployment_schemas.subgraph = subgraph_version.deployment)
 	WHERE subgraphs.subgraph.name = $1`
	rows := []*SubgraphVersion{}
	if err := db.SelectContext(ctx, &rows, query, subgraphName); err != nil {
		return nil, fmt.Errorf("fetch versions specs for %q: %w", subgraphName, err)
	}
	return rows, nil
}

func GetSubgraphVersion(ctx context.Context, db *sqlx.DB, subgraphName, versionID string) (*SubgraphVersion, error) {
	query := `
	SELECT
	subgraphs.subgraph.id AS subgraph_id,
	subgraphs.subgraph_version.id AS version_id,
	subgraphs.subgraph_version.deployment AS deployment_id,
	CASE (subgraphs.subgraph_version.id = subgraphs.subgraph.current_version) WHEN true THEN true ELSE false END AS is_current_version,
	CASE (subgraphs.subgraph_version.id = subgraphs.subgraph.pending_version)  WHEN true then true ELSE false END AS is_pending_version,
	public.deployment_schemas.name AS schema
 	FROM subgraphs.subgraph
 	LEFT JOIN subgraphs.subgraph_version ON (subgraph_version.subgraph = subgraph.id)
 	LEFT JOIN public.deployment_schemas ON (deployment_schemas.subgraph = subgraph_version.deployment)
 	WHERE subgraphs.subgraph.name = $1 AND subgraph_version.id = $2`
	row := &SubgraphVersion{}
	if err := db.GetContext(ctx, row, query, subgraphName, versionID); err != nil {
		return nil, fmt.Errorf("fetch versions specs for %q: %w", subgraphName, err)
	}
	return row, nil
}
