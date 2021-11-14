package cell

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/awstargetgroupset"
	"github.com/mumoshu/okra/pkg/clclient"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Provider struct {
}

type CreateInput struct {
	ListenerARN string
}

func (p *Provider) CreateConfigFromAWS(input CreateInput) error {
	return nil
}

type SyncInput struct {
	NS   string
	Name string

	Spec okrav1alpha1.CellSpec

	Client client.Client
}

func Sync(config SyncInput) error {
	ctx := context.TODO()

	managementClient := config.Client

	if managementClient == nil {
		var err error

		managementClient, err = clclient.New()
		if err != nil {
			return err
		}
	}

	albListenerARN := config.Spec.Ingress.AWSApplicationLoadBalancer.ListenerARN
	tgSelectorMatchLabels := config.Spec.Ingress.AWSApplicationLoadBalancer.TargetGroupSelector
	tgSelector := labels.SelectorFromSet(tgSelectorMatchLabels.MatchLabels)

	var albConfig okrav1alpha1.AWSApplicationLoadBalancerConfig

	if err := managementClient.Get(ctx, types.NamespacedName{Namespace: config.NS, Name: config.Name}, &albConfig); err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		}

		albConfig.Namespace = config.NS
		albConfig.Name = config.Name
		albConfig.Spec.ListenerARN = albListenerARN
	}

	labelKeys := config.Spec.Ingress.AWSApplicationLoadBalancer.TargetGroupSelector.VersionLabels
	if len(labelKeys) == 0 {
		labelKeys = []string{okrav1alpha1.DefaultVersionLabelKey}
	}

	latestTGs, err := awstargetgroupset.ListLatestAWSTargetGroups(awstargetgroupset.ListLatestAWSTargetGroupsInput{
		ListAWSTargetGroupsInput: awstargetgroupset.ListAWSTargetGroupsInput{
			NS:       config.NS,
			Selector: tgSelector.String(),
		},
		SemverLabelKeys: labelKeys,
	})
	if err != nil {
		return err
	}

	desiredTGs := map[string]okrav1alpha1.ForwardTargetGroup{}

	numLatestTGs := len(latestTGs)

	// Ensure there enough cluster replicas to start a canary release
	threshold := 1
	if config.Spec.Replicas != nil {
		threshold = int(*config.Spec.Replicas)
	}

	if numLatestTGs != threshold {
		return nil
	}

	// Do distribute weights evently so that the total becomes 100
	for i, tg := range latestTGs {
		weight := 100 / numLatestTGs

		if i == numLatestTGs-1 && numLatestTGs > 1 {
			weight = 100 - (weight * (numLatestTGs - 1))
		}

		desiredTGs[tg.Name] = okrav1alpha1.ForwardTargetGroup{
			Name:   tg.Name,
			ARN:    tg.Spec.ARN,
			Weight: weight,
		}
	}

	if len(albConfig.Spec.Listener.Rule.Forward.TargetGroups) == 0 {
		// ALB isn't initialized yet so we are creating the ALBConfig resource for the first time
		for _, tg := range desiredTGs {
			albConfig.Spec.Listener.Rule.Forward.TargetGroups = append(albConfig.Spec.Listener.Rule.Forward.TargetGroups, tg)
		}

		if err := managementClient.Create(ctx, &albConfig); err != nil {
			return err
		}
	} else if len(desiredTGs) != len(albConfig.Spec.Listener.Rule.Forward.TargetGroups) {
		// Do update immediately without analysis or step update when
		// it seems to have been triggered by an additional cluster that might have been
		// added to deal with more load.
		for _, tg := range desiredTGs {
			albConfig.Spec.Listener.Rule.Forward.TargetGroups = append(albConfig.Spec.Listener.Rule.Forward.TargetGroups, tg)
		}

		if err := managementClient.Update(ctx, &albConfig); err != nil {
			return err
		}
	} else {
		// This is a standard cell update for releasing a new app/cluster version.
		// Do a canary release.

		// Ensure that the previous analysis run has been successful, if any

		var stableTGsWeight, canaryTGsWeight int

		var stableTGs []okrav1alpha1.ForwardTargetGroup
		for _, tg := range albConfig.Spec.Listener.Rule.Forward.TargetGroups {
			stableTGsWeight += tg.Weight

			tg := tg

			if _, ok := desiredTGs[tg.Name]; ok {
				continue
			}

			stableTGs = append(stableTGs, tg)
		}

		var updatedTGs []okrav1alpha1.ForwardTargetGroup

		{
			canarySteps := config.Spec.UpdateStrategy.Canary.Steps

			var passedAllCanarySteps bool

			if len(canarySteps) > 0 {
				var analysisRunList rolloutsv1alpha1.AnalysisRunList

				var maxSuccessfulAnalysisRunStepIndex int
				for _, ar := range analysisRunList.Items {
					if ar.Status.Phase.Completed() {
						stepIndexStr := ar.Annotations["okra.mumo.co/step-index"]
						stepIndex, err := strconv.Atoi(stepIndexStr)
						if err != nil {
							return err
						}

						if stepIndex > maxSuccessfulAnalysisRunStepIndex {
							maxSuccessfulAnalysisRunStepIndex = stepIndex
						}
					}
				}

				stableTGsWeight = 100

				const stepIndexLabel = "okra.mumo.co/step-index"

			STEPS:
				for stepIndex, step := range canarySteps {
					if step.Analysis != nil {
						//
						// Ensure that the previous analysis run has been successful, if any
						//

						var analysisRunList rolloutsv1alpha1.AnalysisRunList

						stepIndexStr := strconv.Itoa(stepIndex)

						labelSelector, err := labels.Parse(stepIndexLabel + "=" + stepIndexStr)
						if err != nil {
							return err
						}

						if err := managementClient.List(ctx, &analysisRunList, &client.ListOptions{
							LabelSelector: labelSelector,
						}); err != nil {
							return err
						}

						if len(analysisRunList.Items) > 1 {
							return errors.New("too many analysis runs")
						}

						if len(analysisRunList.Items) == 0 {
							tmpl := step.Analysis.Templates[0]

							var args []rolloutsv1alpha1.Argument
							var argsMap map[string]rolloutsv1alpha1.Argument

							var at rolloutsv1alpha1.AnalysisTemplate
							if err := managementClient.Get(ctx, types.NamespacedName{Namespace: config.NS, Name: tmpl.TemplateName}, &at); err != nil {
								return err
							}

							for _, a := range at.Spec.Args {
								argsMap[a.Name] = *a.DeepCopy()
							}

							for _, a := range step.Analysis.Args {
								fromTemplate, ok := argsMap[a.Name]
								if ok {
									if a.Value != "" {
										fromTemplate.Value = &a.Value
									}
									argsMap[a.Name] = fromTemplate
								} else {
									arg := rolloutsv1alpha1.Argument{
										Name: a.Name,
									}

									if a.Value != "" {
										arg.Value = &a.Value
									}

									argsMap[a.Name] = arg
								}
							}

							for _, a := range argsMap {
								args = append(args, a)
							}

							ar := rolloutsv1alpha1.AnalysisRun{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: config.NS,
									Name:      fmt.Sprintf("%s-%s-%d", config.Name, tmpl.TemplateName, stepIndex),
									Labels: map[string]string{
										stepIndexLabel: stepIndexStr,
									},
								},
								Spec: rolloutsv1alpha1.AnalysisRunSpec{
									Args:    args,
									Metrics: at.Spec.Metrics,
								},
							}

							if err := managementClient.Create(ctx, &ar); err != nil {
								return err
							}

							return nil
						}

						for _, ar := range analysisRunList.Items {
							if ar.Status.Phase != rolloutsv1alpha1.AnalysisPhaseSuccessful {
								// We need to wait for this analysis run to succeed
								break STEPS
							}
						}
					} else if step.SetWeight != nil {
						stableTGsWeight -= int(*step.SetWeight)
					} else if step.Pause != nil {
						// TODO List Pause resource and break if it isn't expired yet
					} else {
						return fmt.Errorf("steps[%d]: only setWeight, analysis, and pause step are supported. got %v", stepIndex, step)
					}

					if stepIndex+1 == len(canarySteps) {
						passedAllCanarySteps = true
					}
				}
			}

			if passedAllCanarySteps || len(canarySteps) == 0 {
				stableTGsWeight = 0
			}

			if stableTGsWeight < 0 {
				return fmt.Errorf("stable tgs weight cannot be less than 0: %v", stableTGsWeight)
			}

			// Do update by step weight

			if stableTGsWeight > 0 {
				numStableTGs := len(stableTGs)

				updatedStableTGs := map[string]okrav1alpha1.ForwardTargetGroup{}

				for i, tg := range stableTGs {
					tg := tg

					weight := stableTGsWeight / numStableTGs

					if i == numStableTGs-1 && numStableTGs > 1 {
						weight = stableTGsWeight - (weight * (numStableTGs - 1))
					}

					updatedStableTGs[tg.Name] = okrav1alpha1.ForwardTargetGroup{
						Name:   tg.Name,
						ARN:    tg.ARN,
						Weight: weight,
					}
				}

				for _, tg := range updatedStableTGs {
					updatedTGs = append(updatedTGs, tg)
				}
			}

			canaryTGsWeight = 100 - stableTGsWeight

			if canaryTGsWeight > 0 {
				updatedCanatyTGs := map[string]okrav1alpha1.ForwardTargetGroup{}

				for i, tg := range latestTGs {
					weight := canaryTGsWeight / numLatestTGs

					if i == numLatestTGs-1 && numLatestTGs > 1 {
						weight = canaryTGsWeight - (weight * (numLatestTGs - 1))
					}

					updatedCanatyTGs[tg.Name] = okrav1alpha1.ForwardTargetGroup{
						Name:   tg.Name,
						ARN:    tg.Spec.ARN,
						Weight: weight,
					}
				}

				for _, tg := range updatedCanatyTGs {
					updatedTGs = append(updatedTGs, tg)
				}
			}
		}

		albConfig.Spec.Listener.Rule.Forward.TargetGroups = updatedTGs

		if err := managementClient.Update(ctx, &albConfig); err != nil {
			return err
		}
	}

	return nil
}
