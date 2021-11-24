package cmd

import (
	"fmt"
	"os"

	"github.com/mumoshu/okra/pkg/awstargetgroupset"
	"github.com/spf13/cobra"
)

func getAWSTargetGroupsCommand() *cobra.Command {
	var c awstargetgroupset.ListAWSTargetGroupsInput

	cmd := &cobra.Command{
		Use: "awstargetgroups",
		RunE: func(cmd *cobra.Command, args []string) error {
			bindings, err := awstargetgroupset.ListAWSTargetGroups(c)

			for _, b := range bindings {
				fmt.Fprintf(os.Stdout, "%+v\n", b)
			}

			return err
		},
	}

	flag := cmd.Flags()

	flag.StringVar(&c.NS, "namespace", "", "Namespace of AWSTargetGroup resources")
	flag.StringVar(&c.Selector, "selector", "", "Label selector for AWSTargetGroup resources")

	return cmd
}
