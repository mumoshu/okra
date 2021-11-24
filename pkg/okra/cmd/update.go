package cmd

import "github.com/spf13/cobra"

func UpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "update",
	}

	cmd.AddCommand(updateAnalysisRunCommand())

	return cmd
}
