package deployment

import "github.com/streamingfast/sparkle/entity"

type Subgraph struct {
	ID             string            `db:"id"`
	Name           string            `db:"name"`
	CurrentVersion string            `db:"current_version,nullable"`
	PendingVersion string            `db:"pending_version,nullable"`
	CreatedAt      string            `db:"created_at"`
	VID            int64             `db:"vid"`
	BlockRange     entity.BlockRange `db:"block_range"`
}

type SubgraphVersion struct {
	SubgraphID       string `db:"subgraph_id"`
	DeploymentID     string `db:"deployment_id"`
	VersionID        string `db:"version_id"`
	Schema           string `db:"schema"`
	IsCurrentVersion bool   `db:"is_current_version"`
	IsPendingVersion bool   `db:"is_pending_version"`
}
