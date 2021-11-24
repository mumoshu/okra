package cmd

import (
	"strings"

	"github.com/mumoshu/okra/pkg/clusterset"
	"github.com/spf13/cobra"
)

func syncClusterSetCommand() *cobra.Command {
	var c clusterset.SyncInput

	var (
		eksTags  []string
		labelKVs []string
		create   bool
		delete   bool
	)

	cmd := &cobra.Command{
		Use: "clusterset",
		RunE: func(cmd *cobra.Command, args []string) error {
			tags := map[string]string{}
			for _, kv := range eksTags {
				split := strings.Split(kv, "=")
				tags[split[0]] = split[1]
			}

			c.EKSTags = tags

			labels := map[string]string{}
			for _, kv := range labelKVs {
				split := strings.Split(kv, "=")
				labels[split[0]] = split[1]
			}

			c.Labels = labels

			if create && delete {
				return clusterset.Sync(c)
			} else if create {
				return clusterset.CreateMissingClusters(c)
			} else if delete {
				return clusterset.DeleteOutdatedClusters(c)
			}

			return nil
		},
	}

	flag := cmd.Flags()

	flag.BoolVar(&c.DryRun, "dry-run", false, "")
	flag.StringVar(&c.NS, "namespace", "", "")
	flag.StringSliceVar(&eksTags, "eks-tags", nil, "Comma-separated KEY=VALUE pairs of EKS control-plane tags")
	flag.StringSliceVar(&labelKVs, "labels", nil, "Comma-separated KEY=VALUE pairs of cluster secret labels")
	flag.BoolVar(&create, "create", true, "Sync by creating missing clusters")
	flag.BoolVar(&delete, "delete", true, "Sync by deleting outdated clusters")

	return cmd
}
