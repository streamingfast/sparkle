package main

import (
	"github.com/streamingfast/sparkle/cli"
	"github.com/streamingfast/sparkle/subgraph"
	"github.com/streamingfast/sparkle/testgraph/testgraph"
)

func main() {
	subgraph.MainSubgraphDef = testgraph.Definition
	cli.Execute()
}
