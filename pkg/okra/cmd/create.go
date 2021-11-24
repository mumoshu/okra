package cmd

import "github.com/spf13/cobra"

func CreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create",
	}

	cmd.AddCommand(createAnalysisRunCommand())
	cmd.AddCommand(createClusterCommand())

	return cmd
}
