package cmd

import (
	"errors"
	"strings"

	"github.com/mumoshu/okra/pkg/clusterset"
	"github.com/mumoshu/okra/pkg/okraerror"
	"github.com/spf13/cobra"
)

func createClusterCommand() *cobra.Command {
	var c clusterset.CreateClusterInput

	var labelKVs []string

	cmd := &cobra.Command{
		Use:           "cluster [--namespace ns] [--name name] [--labels k1=v1,k2=v2] [--endpoint https://...] [--ca-data cadata]",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			labels := map[string]string{}
			for _, kv := range labelKVs {
				split := strings.Split(kv, "=")
				labels[split[0]] = split[1]
			}

			c.Labels = labels

			if err := clusterset.CreateCluster(c); err != nil {
				if !errors.Is(err, okraerror.Error{}) {
					cmd.SilenceUsage = true
				}

				return err
			}

			return nil
		},
	}

	flag := cmd.Flags()

	flag.BoolVar(&c.DryRun, "dry-run", c.DryRun, "")
	flag.StringVar(&c.NS, "namespace", c.NS, "")
	flag.StringVar(&c.Name, "name", c.Name, "")
	flag.StringVar(&c.Endpoint, "endpoint", c.Endpoint, "")
	flag.StringVar(&c.CAData, "ca-data", "", "")
	flag.StringSliceVar(&labelKVs, "labels", nil, "Comma-separated KEY=VALUE pairs of cluster secret labels")

	return cmd
}
