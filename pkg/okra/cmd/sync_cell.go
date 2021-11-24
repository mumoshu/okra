package cmd

import (
	"github.com/mumoshu/okra/pkg/cell"
	"github.com/spf13/cobra"
)

func syncCellCommand() *cobra.Command {
	var c cell.SyncInput
	cmd := &cobra.Command{
		Use: "cell",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cell.Sync(c)
			return err
		},
	}

	flag := cmd.Flags()

	flag.StringVar(&c.NS, "namespace", "", "Namespace of the target cell")
	flag.StringVar(&c.Name, "name", "", "Name of the target cell")

	return cmd
}
