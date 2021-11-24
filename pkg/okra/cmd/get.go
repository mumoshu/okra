package cmd

import "github.com/spf13/cobra"

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get",
	}

	cmd.AddCommand(getAWSTargetGroupsCommand())
	cmd.AddCommand(getClustersCommand())
	cmd.AddCommand(getLatestAWSTargetGroupsCommand())
	cmd.AddCommand(getTargetGroupBindingsCommand())

	return cmd
}
