package cmd

import (
	"github.com/spf13/cobra"
)

func CancelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "cancel",
	}
	cmd.AddCommand(cancelPauseCommand())
	return cmd
}
