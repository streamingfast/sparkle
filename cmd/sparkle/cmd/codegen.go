package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/sparkle/codegen"
)

var codegenCmd = &cobra.Command{
	Use:   "codegen <yaml file> <go module path>",
	Short: "generate full skeleton for a subgraph",
	RunE:  runSubgraphCodegenE,
	Args:  cobra.ExactArgs(2),
}

func init() {
	RootCmd.AddCommand(codegenCmd)
	codegenCmd.Flags().Bool("no-gomod", false, "Don't generate go.mod")
}

func runSubgraphCodegenE(cmd *cobra.Command, args []string) error {
	yamlFilePath := args[0]
	goModulePath := args[1]

	userLog.Printf("Running code generation for subgraph")
	engine, err := codegen.NewEngine(yamlFilePath, goModulePath, userLog)
	if err != nil {
		return fmt.Errorf("initiating template engine: %w", err)
	}

	err = engine.GenerateCode(viper.GetBool("codegen-cmd-no-gomod"))
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	userLog.Printf("All done, goodbye")
	return nil
}
