package okra

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/cell"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newSyncCellCommand() *cobra.Command {
	var syncInput func() *cell.SyncInput
	cmd := &cobra.Command{
		Use: "sync-cell",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cell.Sync(*syncInput())
			return err
		},
	}
	syncInput = initSyncCellFlags(cmd.Flags(), &cell.SyncInput{})
	return cmd
}

func initSyncCellFlags(flag *pflag.FlagSet, c *cell.SyncInput) func() *cell.SyncInput {
	var (
		replicas            int
		listenerARN         string
		targetGroupSelector okrav1alpha1.TargetGroupSelector
		canarySteps         []string
		matchLabels         []string
	)

	flag.StringVar(&c.NS, "namespace", "", "Namespace of the target cell")
	flag.StringVar(&c.Name, "name", "", "Name of the target cell")
	flag.StringVar(&listenerARN, "listener-arn", "", "ARN of the target AWS Application Load Balancer Listener that is used to receive all the traffic across cluster versions")
	flag.StringSliceVar(&matchLabels, "match-label", []string{}, "KVs of labels that is used as target group selector")
	flag.StringSliceVar(&targetGroupSelector.VersionLabels, "version-label", []string{okrav1alpha1.DefaultVersionLabelKey}, "Key of the label that is used to indicate the version number of the target group")
	flag.IntVar(&replicas, "", 0, "")
	flag.StringSliceVar(&canarySteps, "canary-steps", []string{}, "List of canary step definitions. Each step is delimited by a comma(,) and can be one of \"weight=INT\", \"pause=DURATION\", and \"analysis=TEMPLATE:arg1=val1:arg2=val2\"")

	return func() *cell.SyncInput {
		spec := c.Spec.DeepCopy()

		if replicas != 0 {
			r32 := int32(replicas)
			spec.Replicas = &r32
		}

		targetGroupSelector.MatchLabels = make(map[string]string)
		for _, l := range matchLabels {
			kv := strings.Split(l, "=")
			targetGroupSelector.MatchLabels[kv[0]] = kv[1]
		}

		var cs []rolloutsv1alpha1.CanaryStep
		for _, s := range canarySteps {
			var kind, arg string

			{
				splits := strings.SplitN(s, "=", 2)

				if len(splits) != 2 {
					panic(fmt.Errorf("pause: unexpected number of args. got %V, wanted only one arg", splits[1:]))
				}

				kind = splits[0]
				arg = splits[1]
			}

			var step rolloutsv1alpha1.CanaryStep

			switch kind {
			case "weight":
				w, err := strconv.Atoi(arg)
				if err != nil {
					panic(fmt.Errorf("parsing weight from %s: %w", arg, err))
				}

				w32 := int32(w)
				step.SetWeight = &w32
			case "pause":
				d, err := time.ParseDuration(arg)
				if err != nil {
					panic(fmt.Errorf("parsing duration from %s: %w", arg, err))
				}

				step.Pause = &rolloutsv1alpha1.RolloutPause{
					Duration: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: d.String(),
					},
				}
			case "analysis":
				tplAndArgs := strings.Split(arg, ":")
				tpl := tplAndArgs[0]

				var args []rolloutsv1alpha1.AnalysisRunArgument
				for _, a := range tplAndArgs[1:] {
					kv := strings.Split(a, "=")

					args = append(args, rolloutsv1alpha1.AnalysisRunArgument{
						Name:  kv[0],
						Value: kv[1],
					})
				}

				step.Analysis = &rolloutsv1alpha1.RolloutAnalysis{
					Templates: []rolloutsv1alpha1.RolloutAnalysisTemplate{
						{
							TemplateName: tpl,
						},
					},
					Args: args,
				}
			default:
				panic(fmt.Errorf("unsupported canary step kind: %s", kind))
			}

			cs = append(cs, step)
		}

		spec.UpdateStrategy = okrav1alpha1.CellUpdateStrategy{
			Type: okrav1alpha1.CellUpdateStrategyTypeCanary,
			Canary: &okrav1alpha1.CellUpdateStrategyCanary{
				Steps: cs,
			},
		}

		spec.Ingress.AWSApplicationLoadBalancer = &okrav1alpha1.CellIngressAWSApplicationLoadBalancer{
			ListenerARN:         listenerARN,
			TargetGroupSelector: targetGroupSelector,
		}

		input := c
		input.Spec = *spec

		return input
	}
}
