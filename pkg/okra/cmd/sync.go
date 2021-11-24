package cmd

import (
	"github.com/spf13/cobra"
)

func syncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "sync",
	}
	cmd.AddCommand(syncClusterSetCommand())
	cmd.AddCommand(syncAWSTargetGroupSetCommand())
	cmd.AddCommand(syncAWSApplicationLoadBalancerConfigCommand())
	cmd.AddCommand(syncCellCommand())
	cmd.AddCommand(syncPauseCommand())
	return cmd
}
