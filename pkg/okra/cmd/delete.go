package cmd

import "github.com/spf13/cobra"

func DeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "delete",
	}

	cmd.AddCommand(deleteClusterCommand())

	return cmd
}
