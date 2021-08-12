package cli

import (
	"github.com/spf13/cobra"
)

var parallelCmd = &cobra.Command{
	Use:           "parallel",
	Short:         "Parallel reprocessing specific commands",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	parallelCmd.PersistentFlags().String("input-path", "", "Data input store")
	parallelCmd.PersistentFlags().String("output-path", "", "Data output store")
	parallelCmd.PersistentFlags().Uint64("start-block", 0, "Start subgraph at block (inclusive)")
	parallelCmd.PersistentFlags().Uint64("stop-block", 0, "Stop batches at block (inclusive)")

	RootCmd.AddCommand(parallelCmd)
}
