package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mumoshu/okra/pkg/targetgroupbinding"
	"github.com/spf13/cobra"
)

func upsertTargetGroupBindingCommand() *cobra.Command {
	var c targetgroupbinding.ApplyInput

	var labelKVs []string

	cmd := &cobra.Command{
		Use: "targetgroupbinding",
		RunE: func(cmd *cobra.Command, args []string) error {
			labels := map[string]string{}
			for _, kv := range labelKVs {
				split := strings.Split(kv, "=")
				labels[split[0]] = split[1]
			}

			c.Labels = labels

			binding, err := targetgroupbinding.Apply(c)

			if binding != nil {
				fmt.Fprintf(os.Stdout, "%+v\n", binding)
			}

			return err
		},
	}

	flag := cmd.Flags()

	flag.BoolVar(&c.DryRun, "dry-run", c.DryRun, "")
	flag.StringVar(&c.ClusterNamespace, "cluster-namespace", c.ClusterNamespace, "")
	flag.StringVar(&c.ClusterName, "cluster-name", c.ClusterName, "")
	flag.StringVar(&c.TargetGroupARN, "target-group-arn", c.TargetGroupARN, "")
	flag.StringVar(&c.Name, "name", c.Name, "")
	flag.StringVar(&c.Namespace, "namespace", c.Namespace, "")
	flag.StringSliceVar(&labelKVs, "labels", nil, "Comma-separated KEY=VALUE pairs of cluster secret labels")

	return cmd
}
