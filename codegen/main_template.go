package codegen

var templateSubgraphMain = `
package main

import (
	"github.com/streamingfast/sparkle/cli"
	"{{ .GoModulePath }}/{{ .PackageName }}"
	"github.com/streamingfast/sparkle/subgraph"
	"github.com/streamingfast/sparkle/entity"
)

func main() {
	subgraph.MainSubgraphDef = {{ .PackageName }}.Definition
	cli.Execute()
}
`

var templateSubgraphMainGoMod = `
module {{ .GoModulePath }}
go 1.15
require (
	github.com/streamingfast/eth-go master
	github.com/streamingfast/sparkle master
	go.uber.org/zap v1.16.0
)
`
