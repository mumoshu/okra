package okra

import (
	"github.com/mumoshu/okra/pkg/cell"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func syncCellCommand() *cobra.Command {
	var syncInput func() *cell.SyncInput
	cmd := &cobra.Command{
		Use: "sync-cell",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cell.Sync(*syncInput())
			return err
		},
	}
	syncInput = initSyncCellFlags(cmd.Flags(), &cell.SyncInput{})
	return cmd
}

func initSyncCellFlags(flag *pflag.FlagSet, c *cell.SyncInput) func() *cell.SyncInput {
	flag.StringVar(&c.NS, "namespace", "", "Namespace of the target cell")
	flag.StringVar(&c.Name, "name", "", "Name of the target cell")

	return func() *cell.SyncInput {
		return c
	}
}
