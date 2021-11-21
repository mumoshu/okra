package okra

import (
	"github.com/mumoshu/okra/pkg/pause"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func syncPauseCommand() *cobra.Command {
	var input func() *pause.SyncInput
	cmd := &cobra.Command{
		Use: "sync-pause",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := pause.Sync(*input())
			return err
		},
	}
	input = initSyncPauseFlags(cmd.Flags(), &pause.SyncInput{})
	return cmd
}

func initSyncPauseFlags(flag *pflag.FlagSet, c *pause.SyncInput) func() *pause.SyncInput {
	flag.StringVar(&c.Pause.ObjectMeta.Namespace, "namespace", "", "Namespace of the pause")
	flag.StringVar(&c.Pause.ObjectMeta.Name, "name", "", "Name of the pause")

	return func() *pause.SyncInput {
		input := c
		input.Pause = *c.Pause.DeepCopy()

		return input
	}
}
