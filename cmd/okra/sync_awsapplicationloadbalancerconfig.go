package okra

import (
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/awsapplicationloadbalancer"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func syncAWSApplicationLoadBalancerConfigCommand() *cobra.Command {
	var syncInput func() *awsapplicationloadbalancer.SyncInput
	cmd := &cobra.Command{
		Use: "sync-awsapplicationloadbalancerconfig",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := awsapplicationloadbalancer.Sync(*syncInput())
			return err
		},
	}
	syncInput = initSyncAWSApplicationLoadBalancerConfigFlags(cmd.Flags(), &awsapplicationloadbalancer.SyncInput{})
	return cmd
}

func initSyncAWSApplicationLoadBalancerConfigFlags(flag *pflag.FlagSet, c *awsapplicationloadbalancer.SyncInput) func() *awsapplicationloadbalancer.SyncInput {
	var (
		tg1, tg2 okrav1alpha1.ForwardTargetGroup
	)

	flag.StringVar(&c.Region, "region", "", "AWS region where the target ALB is in")
	flag.StringVar(&c.Profile, "profile", "", "AWS profile that is used to access the target ALB")
	flag.StringVar(&c.Address, "address", "", "Custom address of AWS API endpoint that is used when testing")
	flag.StringVar(&c.Spec.ListenerARN, "listener-arn", "", "ARN of the AWS ALB Listener on which the Listener Rule for traffic management is created")
	flag.IntVar(&c.Spec.Listener.Rule.Priority, "listener-rule-priority", 0, "Priority of the ALB Listener Rule that is used as a unique ID of it")
	flag.StringVar(&tg1.ARN, "target-group-1-arn", "", "ARN of the first target group")
	flag.IntVar(&tg1.Weight, "target-group-1-weight", 50, "Weight of the first target group")
	flag.StringVar(&tg2.ARN, "target-group-2-arn", "", "ARN of the second target group")
	flag.IntVar(&tg2.Weight, "target-group-2-weight", 50, "Weight of the second target group")

	return func() *awsapplicationloadbalancer.SyncInput {
		tg1.Name = "first"
		tg2.Name = "second"

		spec := c.Spec.DeepCopy()
		spec.Listener.Rule.Forward.TargetGroups = append(spec.Listener.Rule.Forward.TargetGroups, tg1, tg2)

		input := c
		input.Spec = *spec

		return input
	}
}
