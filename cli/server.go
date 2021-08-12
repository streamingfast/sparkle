package cli

import (
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:           "server",
	Short:         "HTTP & GraphQL server specific commands",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
