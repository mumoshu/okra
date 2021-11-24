package cmd

import (
	_ "github.com/aws/aws-sdk-go/service/eks"
	_ "k8s.io/client-go/plugin/pkg/client/auth/exec"

	"github.com/spf13/cobra"
)

const (
	ApplicationName = "okra"
)

func Run() error {
	cmd := &cobra.Command{
		Use: ApplicationName,
	}

	cmd.AddCommand(CancelCommand())
	cmd.AddCommand(CreateCommand())
	cmd.AddCommand(DeleteCommand())
	cmd.AddCommand(GetCommand())
	cmd.AddCommand(syncCommand())
	cmd.AddCommand(UpdateCommand())
	cmd.AddCommand(UpsertCommand())

	err := cmd.Execute()

	return err
}
