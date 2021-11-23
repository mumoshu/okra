package okra

import (
	"fmt"
	"os"
	"strings"

	"github.com/mumoshu/okra/pkg/awstargetgroupset"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func createOrUpdateAWSTargetGroupCommand() *cobra.Command {
	var input func() *awstargetgroupset.ApplyInput
	cmd := &cobra.Command{
		Use: "create-or-update-awstargetgroupset",
		RunE: func(cmd *cobra.Command, args []string) error {
			created, err := awstargetgroupset.CreateOrUpdate(*input())

			if created != nil {
				fmt.Fprintf(os.Stdout, "%+v\n", created)
			}

			return err
		},
	}
	input = InitCreateOrUpdateAWSTargetGroupSetFlags(cmd.Flags(), &awstargetgroupset.ApplyInput{})

	return cmd
}

func InitCreateOrUpdateAWSTargetGroupSetFlags(flag *pflag.FlagSet, c *awstargetgroupset.ApplyInput) func() *awstargetgroupset.ApplyInput {
	var (
		bindingSelector string
		clusterSelector string
		labelKVs        []string
	)

	flag.StringVar(&c.Name, "name", "", "Name of AWSTargetGroupSet to be created or updated")
	flag.StringVar(&c.NS, "namespace", "", "Namespace of the ArgoCD Cluster and the generated AWSTargetGroup resources")
	flag.StringVar(&bindingSelector, "targetgroupbinding-selector", "", "Comma-separated KEY=VALUE pairs of TargetGroupBinding resource labels to be used as selector")
	flag.StringVar(&clusterSelector, "cluster-selector", "", "Comma-separated KEY=VALUE pairs of ArgoCD Cluster resource labels to be used as selector")
	flag.StringSliceVar(&labelKVs, "labels", nil, "Comma-separated KEY=VALUE pairs of AWSTargetGroup labels")

	return func() *awstargetgroupset.ApplyInput {
		bindings := make(map[string]string)
		bindingKVs := strings.Split(bindingSelector, ",")
		for _, kv := range bindingKVs {
			split := strings.Split(kv, "=")
			bindings[split[0]] = split[1]
		}

		clusters := make(map[string]string)
		clusterKVs := strings.Split(clusterSelector, ",")
		for _, kv := range clusterKVs {
			split := strings.Split(kv, "=")
			clusters[split[0]] = split[1]
		}

		c.BindingSelector = bindings
		c.ClusterSelector = clusters

		labels := map[string]string{}
		for _, kv := range labelKVs {
			split := strings.Split(kv, "=")
			labels[split[0]] = split[1]
		}

		c.Labels = labels

		return c
	}
}
