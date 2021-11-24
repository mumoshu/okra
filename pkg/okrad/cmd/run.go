package cmd

import (
	"errors"

	"github.com/mumoshu/okra/pkg/manager"
	"github.com/mumoshu/okra/pkg/okraerror"
	"github.com/spf13/cobra"
)

const (
	ApplicationName = "okrad"
)

func Run() error {
	m := &manager.Manager{}

	cmd := &cobra.Command{
		Use: ApplicationName,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := m.Run()
			if !errors.Is(err, okraerror.Error{}) {
				cmd.SilenceUsage = true
			}
			return err
		},
	}

	m.AddPFlags(cmd.Flags())

	err := cmd.Execute()

	return err
}
