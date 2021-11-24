package cmd

import (
	"github.com/mumoshu/okra/pkg/clusterset"
	"github.com/spf13/cobra"
)

func deleteClusterCommand() *cobra.Command {
	var c clusterset.DeleteClusterInput

	cmd := &cobra.Command{
		Use: "cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterset.DeleteCluster(c)
		},
	}

	flag := cmd.Flags()

	flag.BoolVar(&c.DryRun, "dry-run", c.DryRun, "")
	flag.StringVar(&c.NS, "namespace", c.NS, "")
	flag.StringVar(&c.Name, "name", c.Name, "")

	return cmd
}
