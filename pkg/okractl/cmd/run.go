package cmd

import (
	"github.com/spf13/cobra"

	okracmd "github.com/mumoshu/okra/pkg/okra/cmd"
)

const (
	ApplicationName = "okractl"
)

func Run() error {
	cmd := &cobra.Command{
		Use: ApplicationName,
	}

	cmd.AddCommand(okracmd.CancelCommand())
	cmd.AddCommand(okracmd.CreateCommand())
	cmd.AddCommand(okracmd.DeleteCommand())
	cmd.AddCommand(okracmd.GetCommand())
	cmd.AddCommand(okracmd.UpdateCommand())
	cmd.AddCommand(okracmd.UpsertCommand())

	err := cmd.Execute()

	return err
}
