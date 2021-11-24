package cmd

import "github.com/spf13/cobra"

func UpsertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "upsert",
	}

	cmd.AddCommand(upsertAWSTargetGroupSetCommand())
	cmd.AddCommand(upsertCellCommand())
	cmd.AddCommand(upsertTargetGroupBindingCommand())

	return cmd
}
