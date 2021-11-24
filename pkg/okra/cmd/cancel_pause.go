package cmd

import (
	"github.com/mumoshu/okra/pkg/pause"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func cancelPauseCommand() *cobra.Command {
	var input func() *pause.CancelInput
	cmd := &cobra.Command{
		Use: "pause",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := pause.Cancel(*input())
			return err
		},
	}
	input = initCancelPauseFlags(cmd.Flags(), &pause.CancelInput{})
	return cmd
}

func initCancelPauseFlags(flag *pflag.FlagSet, c *pause.CancelInput) func() *pause.CancelInput {
	flag.StringVar(&c.Pause.ObjectMeta.Namespace, "namespace", "", "Namespace of the pause")
	flag.StringVar(&c.Pause.ObjectMeta.Name, "name", "", "Name of the pause")

	return func() *pause.CancelInput {
		input := c
		input.Pause = *c.Pause.DeepCopy()

		return input
	}
}
