package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mumoshu/okra/pkg/awstargetgroupset"
	"github.com/spf13/cobra"
)

func syncAWSTargetGroupSetCommand() *cobra.Command {
	var c awstargetgroupset.SyncInput

	var (
		bindingSelector string
		labelKVs        []string
		create          bool
		delete          bool
	)

	cmd := &cobra.Command{
		Use: "awstargetgroupset",
		RunE: func(cmd *cobra.Command, args []string) error {
			c.BindingSelector = bindingSelector

			labels := map[string]string{}
			for _, kv := range labelKVs {
				split := strings.Split(kv, "=")
				labels[split[0]] = split[1]
			}

			c.Labels = labels

			var (
				bindings []awstargetgroupset.SyncResult
				err      error
			)

			if create && delete {
				bindings, err = awstargetgroupset.Sync(c)
			} else if create {
				bindings, err = awstargetgroupset.CreateMissingAWSTargetGroups(c)
			} else if delete {
				bindings, err = awstargetgroupset.DeleteOutdatedAWSTargetGroups(c)
			}

			for _, b := range bindings {
				fmt.Fprintf(os.Stdout, "%+v\n", b)
			}

			return err
		},
	}

	flag := cmd.Flags()

	flag.BoolVar(&c.DryRun, "dry-run", false, "")
	flag.StringVar(&c.ClusterName, "cluster-name", "", "ArgoCD Cluster name on which we find TargetGroupBinding")
	flag.StringVar(&c.NS, "namespace", "", "Namespace of the ArgoCD Cluster and the generated AWSTargetGroup resources")
	flag.StringVar(&bindingSelector, "targetgroupbinding-selector", "", "Comma-separated KEY=VALUE pairs of TargetGroupBinding resource labels")
	flag.StringSliceVar(&labelKVs, "labels", nil, "Comma-separated KEY=VALUE pairs of AWSTargetGroup labels")
	flag.BoolVar(&create, "create", true, "Sync by creating missing AWSTargetGroup resources")
	flag.BoolVar(&delete, "delete", true, "Sync by deleting outdated AWSTargetGroup resources")

	return cmd
}
