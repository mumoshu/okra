package cmd

import (
	"github.com/mumoshu/okra/pkg/pause"
	"github.com/spf13/cobra"
)

func syncPauseCommand() *cobra.Command {
	var c pause.SyncInput
	cmd := &cobra.Command{
		Use: "pause",
		RunE: func(cmd *cobra.Command, args []string) error {
			input := c
			input.Pause = *c.Pause.DeepCopy()

			err := pause.Sync(c)
			return err
		},
	}

	flag := cmd.Flags()

	flag.StringVar(&c.Pause.ObjectMeta.Namespace, "namespace", "", "Namespace of the pause")
	flag.StringVar(&c.Pause.ObjectMeta.Name, "name", "", "Name of the pause")

	return cmd
}
