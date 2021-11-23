package cmd

import "github.com/spf13/cobra"

const (
	ApplicationName = "okratest"
)

func Run() error {
	cmd := &cobra.Command{
		Use: ApplicationName,
	}

	err := cmd.Execute()

	return err
}
