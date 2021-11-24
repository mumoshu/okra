package cmd

import (
	"fmt"
	"os"

	"github.com/mumoshu/okra/pkg/targetgroupbinding"
	"github.com/spf13/cobra"
)

func getTargetGroupBindingsCommand() *cobra.Command {
	var c targetgroupbinding.ListInput

	cmd := &cobra.Command{
		Use: "targetgroupbindings",
		RunE: func(cmd *cobra.Command, args []string) error {
			bindings, err := targetgroupbinding.List(c)

			for _, b := range bindings {
				fmt.Fprintf(os.Stdout, "%+v\n", b)
			}

			return err
		},
	}

	flag := cmd.Flags()

	flag.StringVar(&c.ClusterName, "cluster-name", "", "")
	flag.StringVar(&c.NS, "namespace", "", "")

	return cmd
}
