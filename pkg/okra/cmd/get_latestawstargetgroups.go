package cmd

import (
	"fmt"
	"os"

	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/awstargetgroupset"
	"github.com/spf13/cobra"
)

func getLatestAWSTargetGroupsCommand() *cobra.Command {
	var c awstargetgroupset.ListLatestAWSTargetGroupsInput

	cmd := &cobra.Command{
		Use: "latestawstargetgroups",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, bindings, err := awstargetgroupset.ListLatestAWSTargetGroups(c)

			for _, b := range bindings {
				fmt.Fprintf(os.Stdout, "%+v\n", b)
			}

			return err
		},
	}

	flag := cmd.Flags()

	flag.StringVar(&c.NS, "namespace", "", "Namespace of AWSTargetGroup resources")
	flag.StringVar(&c.Selector, "selector", "", "Label selector for AWSTargetGroup resources")
	flag.StringVar(&c.Version, "version", "", "Version number without the v prefix of the AWSTargetGroup resources. If omitted, it will fetch all the AWSTargetGroup resources and use the latest version found")
	flag.StringSliceVar(&c.SemverLabelKeys, "semver-label-key", []string{okrav1alpha1.DefaultVersionLabelKey}, "The key of the label as a container of the version number of the group")

	return cmd
}
