package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mumoshu/okra/pkg/clusterset"
	"github.com/mumoshu/okra/pkg/okraerror"
	"github.com/spf13/cobra"
)

func getClustersCommand() *cobra.Command {
	var c clusterset.ListClustersInput

	cmd := &cobra.Command{
		Use:           "cluster",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			clusters, err := clusterset.ListClusters(c)
			if err != nil {
				if !errors.Is(err, okraerror.Error{}) {
					cmd.SilenceUsage = true
				}

				return err
			}

			for _, c := range clusters {
				var kvs []string
				for k, v := range c.Labels {
					kvs = append(kvs, k+"="+v)
				}
				labels := strings.Join(kvs, ",")
				fmt.Fprintf(os.Stdout, "%v\t%s\n", c.Name, labels)
			}

			return nil
		},
	}

	flag := cmd.Flags()

	flag.StringVar(&c.NS, "namespace", c.NS, "")
	flag.StringVar(&c.Selector, "selectoro", "", "")

	return cmd
}
