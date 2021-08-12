package cli

import (
	"context"
	"fmt"

	"github.com/streamingfast/dstore"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var parallelCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check steps state ",
	Args:  cobra.NoArgs,
	RunE:  runParallelCheck,
}

func init() {
	parallelCheckCmd.Flags().String("store-path", "", "Step store path to check")
	parallelCmd.AddCommand(parallelCheckCmd)
}

func runParallelCheck(_ *cobra.Command, _ []string) error {
	storePath := viper.GetString("parallel-check-cmd-store-path")
	if storePath == "" {
		return fmt.Errorf("ERROR - `--store-path` is required to perform a contiguity check")
	}
	store, err := dstore.NewStore(storePath, "", "", true)
	if err != nil {
		return fmt.Errorf("unable to create step store: %w", err)
	}

	seenBlock := false
	iterBlockNum := uint64(0)
	err = store.Walk(context.Background(), "", "", func(filename string) (err error) {
		startBlockNum, stopBlockNum, err := getBlockRange(filename)
		if err != nil {
			return err
		}
		if !seenBlock {
			iterBlockNum = stopBlockNum
			fmt.Printf("✅ Range %s\n", BlockRange{Start: startBlockNum, Stop: stopBlockNum})
			seenBlock = true
			return nil
		}
		if iterBlockNum != startBlockNum {
			fmt.Printf("❌ Range %s! (Broken range, start block is not equal to last stop block)\n", BlockRange{startBlockNum, stopBlockNum})
		} else {
			fmt.Printf("✅ Range %s\n", BlockRange{Start: startBlockNum, Stop: stopBlockNum})

		}
		iterBlockNum = stopBlockNum

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to walk input store: %w", err)
	}
	return nil
}
