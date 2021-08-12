package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/streamingfast/sparkle/entity"
	"github.com/streamingfast/sparkle/subgraph"
)

var RootCmd = &cobra.Command{
	Use:           "subgraph",
	Short:         "Live & Parallel Processor for Subgraphs",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().String("log-format", "text", "Format for logging to stdout. Either 'text' or 'stackdriver'")
	RootCmd.PersistentFlags().String("log-level-switcher-listen-addr", "localhost:1065", "If non-empty, the process will listen on this address for json-formatted requests to change different logger levels (see DEBUG.md for more info)")
	RootCmd.PersistentFlags().CountP("verbose", "v", "Enables verbose output (-vvvv for max verbosity)")
}

func Execute() {
	if subgraph.MainSubgraphDef == nil {
		fmt.Println("Error subgraph is not set, you are mostly likely running the subgraph cli commang via the sparkle repo. This command is meant to be run withing a subgraph repo.")
		os.Exit(1)
	}

	subgraphDef := subgraph.MainSubgraphDef
	lines := []string{
		RootCmd.Short,
		"",
		fmt.Sprintf("Subgraph: %s)\n", subgraphDef.PackageName),
		fmt.Sprintf("Start block: %d)\n", subgraphDef.StartBlock),
	}

	sg := []string{}
	for _, ent := range subgraphDef.Entities.Entities() {
		var add string
		if _, ok := ent.(entity.Finalizable); ok {
			add += ", finalizeable"
		}
		if _, ok := ent.(entity.Mergeable); ok {
			add += ", mergeable"
		}
		if _, ok := ent.(entity.Sanitizable); ok {
			add += ", sanitizable"
		}
		sg = append(sg, fmt.Sprintf("  * %s%s", entity.GetTableName(ent), add))
	}
	lines = append(lines, sg...)
	lines = append(lines, "")
	RootCmd.Long = strings.Join(lines, "\n")

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetEnvPrefix("SPARKLE")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	recurseViperCommands(RootCmd, nil)
	setupLogger(&LoggingOptions{
		Verbosity: viper.GetInt("global-verbose"),
		LogFormat: viper.GetString("global-log-format"),
	})
}

func recurseViperCommands(root *cobra.Command, segments []string) {
	// Stolen from: github.com/abourget/viperbind
	var segmentPrefix string
	if len(segments) > 0 {
		segmentPrefix = strings.Join(segments, "-") + "-"
	}

	root.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		newVar := segmentPrefix + "global-" + f.Name
		viper.BindPFlag(newVar, f)
	})
	root.Flags().VisitAll(func(f *pflag.Flag) {
		newVar := segmentPrefix + "cmd-" + f.Name
		viper.BindPFlag(newVar, f)
	})

	for _, cmd := range root.Commands() {
		recurseViperCommands(cmd, append(segments, cmd.Name()))
	}
}
