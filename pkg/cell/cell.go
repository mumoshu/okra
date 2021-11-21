package cell

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/awstargetgroupset"
	"github.com/mumoshu/okra/pkg/clclient"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LabelKeyStepIndex = "okra.mumo.co/step-index"
	LabelKeyCell      = "cell"
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

	Cell *okrav1alpha1.Cell

	Scheme *runtime.Scheme
	Client client.Client
}

func Sync(config SyncInput) error {
	ctx := context.TODO()

	runtimeClient, scheme, err := clclient.Init(config.Client, config.Scheme)
	if err != nil {
		return err
	}

	var cell okrav1alpha1.Cell

	if config.Cell != nil {
		cell = *config.Cell
	} else {
		if err := runtimeClient.Get(ctx, types.NamespacedName{Namespace: config.NS, Name: config.Name}, &cell); err != nil {
			return err
		}
	}

	albListenerARN := cell.Spec.Ingress.AWSApplicationLoadBalancer.ListenerARN
	tgSelectorMatchLabels := cell.Spec.Ingress.AWSApplicationLoadBalancer.TargetGroupSelector
	tgSelector := labels.SelectorFromSet(tgSelectorMatchLabels.MatchLabels)

	var albConfig okrav1alpha1.AWSApplicationLoadBalancerConfig
	var albConfigExists bool

	if err := runtimeClient.Get(ctx, types.NamespacedName{Namespace: cell.Namespace, Name: cell.Name}, &albConfig); err != nil {
		log.Printf("%v\n", err)
		if !kerrors.IsNotFound(err) {
			return err
		}

		albConfig.Namespace = cell.Namespace
		albConfig.Name = cell.Name
		albConfig.Spec.ListenerARN = albListenerARN
		albConfig.Spec.Listener = cell.Spec.Ingress.AWSApplicationLoadBalancer.Listener
		ctrl.SetControllerReference(&cell, &albConfig, scheme)
	} else {
		albConfigExists = true
	}

	labelKeys := cell.Spec.Ingress.AWSApplicationLoadBalancer.TargetGroupSelector.VersionLabels
	if len(labelKeys) == 0 {
		labelKeys = []string{okrav1alpha1.DefaultVersionLabelKey}
	}

	desiredVer, latestTGs, err := awstargetgroupset.ListLatestAWSTargetGroups(awstargetgroupset.ListLatestAWSTargetGroupsInput{
		ListAWSTargetGroupsInput: awstargetgroupset.ListAWSTargetGroupsInput{
			NS:       config.NS,
			Selector: tgSelector.String(),
		},
		SemverLabelKeys: labelKeys,
	})
	if err != nil {
		return err
	}

	currentTGs, err := awstargetgroupset.ListAWSTargetGroups(awstargetgroupset.ListAWSTargetGroupsInput{
		NS:       config.NS,
		Selector: tgSelector.String(),
	})
	if err != nil {
		return err
	}

	currentTGNameToVer := make(map[string]string)
	for _, tg := range currentTGs {
		var ver string
		for _, l := range labelKeys {
			v, ok := tg.Labels[l]
			if ok {
				ver = v
				break
			}
		}

		if ver != "" {
			currentTGNameToVer[tg.Name] = ver
		}
	}

	desiredTGs := map[string]okrav1alpha1.ForwardTargetGroup{}

	numLatestTGs := len(latestTGs)

	// Ensure there enough cluster replicas to start a canary release
	threshold := 1
	if cell.Spec.Replicas != nil {
		threshold = int(*cell.Spec.Replicas)
	}

	log.Printf("cell=%s/%s, albConfigExists=%v, tgSelector=%s, len(latestTGs)=%d, len(desiredTGs)=%d\n", cell.Namespace, cell.Name, albConfigExists, tgSelector.String(), len(latestTGs), len(desiredTGs))

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

	if !albConfigExists {
		// ALB isn't initialized yet so we are creating the ALBConfig resource for the first time
		for _, tg := range desiredTGs {
			albConfig.Spec.Listener.Rule.Forward.TargetGroups = append(albConfig.Spec.Listener.Rule.Forward.TargetGroups, tg)
		}

		if err := runtimeClient.Create(ctx, &albConfig); err != nil {
			return fmt.Errorf("creating albconfig: %w", err)
		}

		updated := make(map[string]int)
		for _, tg := range desiredTGs {
			updated[tg.Name] = tg.Weight
		}

		log.Printf("Created target groups and weights to: %v", updated)
	} else {
		// This is a standard cell update for releasing a new app/cluster version.
		// Do a canary release.

		// Ensure that the previous analysis run has been successful, if any

		var currentStableTGsWeight, currentCanaryTGsWeight, canaryTGsWeight int

		var (
			stableTGs []okrav1alpha1.ForwardTargetGroup
			canaryTGs []okrav1alpha1.ForwardTargetGroup
		)

		for _, tg := range albConfig.Spec.Listener.Rule.Forward.TargetGroups {
			tg := tg

			if _, ok := desiredTGs[tg.Name]; ok {
				currentCanaryTGsWeight += tg.Weight
				canaryTGs = append(canaryTGs, tg)
				continue
			}

			stableTGs = append(stableTGs, tg)

			currentStableTGsWeight += tg.Weight
		}

		var desiredAndCanaryAreSameVersion bool

		if len(canaryTGs) > 0 {
			for _, tg := range canaryTGs {
				ver := currentTGNameToVer[tg.Name]
				if ver == desiredVer.String() {
					desiredAndCanaryAreSameVersion = true
					break
				}
			}
		}

		// TODO check desired TGs version and canary TGs version and do immediate update only when the version matches?
		if desiredAndCanaryAreSameVersion && len(desiredTGs) != len(canaryTGs) {
			// Do update immediately without analysis or step update when
			// it seems to have been triggered by an additional cluster that might have been
			// added to deal with more load.
			albConfig.Spec.Listener.Rule.Forward.TargetGroups = nil
			for _, tg := range desiredTGs {
				albConfig.Spec.Listener.Rule.Forward.TargetGroups = append(albConfig.Spec.Listener.Rule.Forward.TargetGroups, tg)
			}

			if err := runtimeClient.Update(ctx, &albConfig); err != nil {
				return fmt.Errorf("updating albconfig: %w", err)
			}

			updated := make(map[string]int)
			for _, tg := range desiredTGs {
				updated[tg.Name] = tg.Weight
			}

			log.Printf("Updated target groups and weights to: %v", updated)

			return nil
		}

		var updatedTGs []okrav1alpha1.ForwardTargetGroup

		var passedAllCanarySteps bool

		// TODO Use client.MatchingLabels?
		ownedByCellLabelSelector, err := labels.Parse(LabelKeyCell + "=" + cell.Name)
		if err != nil {
			return err
		}

		desiredStableTGsWeight := 100

		{
			canarySteps := cell.Spec.UpdateStrategy.Canary.Steps

			passedAllCanarySteps = currentCanaryTGsWeight == 100

			if len(canarySteps) > 0 && !passedAllCanarySteps {
				var analysisRunList rolloutsv1alpha1.AnalysisRunList

				if err := runtimeClient.List(ctx, &analysisRunList, &client.ListOptions{
					LabelSelector: ownedByCellLabelSelector,
				}); err != nil {
					return err
				}

				var maxSuccessfulAnalysisRunStepIndex int
				for _, ar := range analysisRunList.Items {
					if ar.Status.Phase.Completed() {
						stepIndexStr, ok := ar.Labels[LabelKeyStepIndex]
						if !ok {
							log.Printf("AnalysisRun %q does not have as step-index label. Perhaps this is not the one managed by okra? Skipping.", ar.Name)
							continue
						}
						stepIndex, err := strconv.Atoi(stepIndexStr)
						if err != nil {
							return fmt.Errorf("parsing step index %q: %v", stepIndexStr, err)
						}

						if stepIndex > maxSuccessfulAnalysisRunStepIndex {
							maxSuccessfulAnalysisRunStepIndex = stepIndex
						}
					}
				}

			STEPS:
				for stepIndex, step := range canarySteps {
					stepIndexStr := strconv.Itoa(stepIndex)

					if step.Analysis != nil {
						//
						// Ensure that the previous analysis run has been successful, if any
						//

						var analysisRunList rolloutsv1alpha1.AnalysisRunList

						labelSelector, err := labels.Parse(LabelKeyStepIndex + "=" + stepIndexStr)
						if err != nil {
							return err
						}

						if err := runtimeClient.List(ctx, &analysisRunList, &client.ListOptions{
							LabelSelector: labelSelector,
						}); err != nil {
							return err
						}

						switch len(analysisRunList.Items) {
						case 0:
							tmpl := step.Analysis.Templates[0]

							var args []rolloutsv1alpha1.Argument
							argsMap := make(map[string]rolloutsv1alpha1.Argument)

							var at rolloutsv1alpha1.AnalysisTemplate
							if err := runtimeClient.Get(ctx, types.NamespacedName{Namespace: config.NS, Name: tmpl.TemplateName}, &at); err != nil {
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
										LabelKeyStepIndex: stepIndexStr,
										LabelKeyCell:      config.Name,
									},
								},
								Spec: rolloutsv1alpha1.AnalysisRunSpec{
									Args:    args,
									Metrics: at.Spec.Metrics,
								},
							}
							ctrl.SetControllerReference(&cell, &ar, scheme)

							if err := runtimeClient.Create(ctx, &ar); err != nil {
								return err
							}

							log.Printf("Created analysisrun %s", ar.Name)

							break STEPS
						case 1:
							for _, ar := range analysisRunList.Items {
								if ar.Status.Phase != rolloutsv1alpha1.AnalysisPhaseSuccessful {
									if ar.Status.Phase == rolloutsv1alpha1.AnalysisPhaseFailed {
										// TODO Suspend and mark it as permanent failure when analysis run timed out
										log.Printf("AnalysisRun %s failed", ar.Name)
										break STEPS
									}

									log.Printf("Waiting for analysisrun %s of %s to become %s", ar.Name, ar.Status.Phase, rolloutsv1alpha1.AnalysisPhaseSuccessful)

									// We need to wait for this analysis run to succeed
									break STEPS
								}
							}
						default:
							return errors.New("too many analysis runs")
						}
					} else if step.SetWeight != nil {
						desiredStableTGsWeight -= int(*step.SetWeight)

						if desiredStableTGsWeight < currentStableTGsWeight {
							break STEPS
						}
					} else if step.Pause != nil {
						// TODO List Pause resource and break if it isn't expired yet
						var pauseList okrav1alpha1.PauseList

						ns := config.NS

						labels := map[string]string{
							LabelKeyStepIndex: stepIndexStr,
							LabelKeyCell:      config.Name,
						}

						if err := runtimeClient.List(ctx, &pauseList, client.InNamespace(ns), client.MatchingLabels(labels)); err != nil {
							return err
						}

						switch c := len(pauseList.Items); c {
						case 0:
							t := metav1.Time{
								Time: time.Now().Add(time.Duration(time.Second.Nanoseconds() * int64(step.Pause.DurationSeconds()))),
							}

							pause := okrav1alpha1.Pause{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: ns,
									Name:      fmt.Sprintf("%s-%d-%s", config.Name, stepIndex, "pause"),
									Labels:    labels,
								},
								Spec: okrav1alpha1.PauseSpec{
									ExpireTime: t,
								},
							}
							ctrl.SetControllerReference(&cell, &pause, scheme)

							if err := runtimeClient.Create(ctx, &pause); err != nil {
								return err
							}

							log.Printf("Initiated pause %s until %s", pause.Name, t)

							break STEPS
						case 1:
							pause := pauseList.Items[0]

							switch phase := pause.Status.Phase; phase {
							case okrav1alpha1.PausePhaseCancelled:
								log.Printf("Observed that pause %s had been cancelled. Continuing to the next step", pause.Name)
							case okrav1alpha1.PausePhaseExpired:
								log.Printf("Observed that pause %s had expired. Continuing to the next step", pause.Name)
							case okrav1alpha1.PausePhaseStarted:
								log.Printf("Still waiting for pause %s to expire or get cancelled", pause.Name)
								break STEPS
							case "":
								log.Printf("Still waiting for pause %s to start", pause.Name)
								break STEPS
							default:
								return fmt.Errorf("unexpected pause phase: %s", phase)
							}
						default:
							return fmt.Errorf("unexpected number of pauses found: %d", c)
						}
					} else {
						return fmt.Errorf("steps[%d]: only setWeight, analysis, and pause step are supported. got %v", stepIndex, step)
					}

					if stepIndex+1 == len(canarySteps) {
						passedAllCanarySteps = true
					}
				}
			}

			if passedAllCanarySteps || len(canarySteps) == 0 {
				desiredStableTGsWeight = 0
			}

			if desiredStableTGsWeight < 0 {
				return fmt.Errorf("stable tgs weight cannot be less than 0: %v", desiredStableTGsWeight)
			}

			log.Printf("stable weight: %d -> %d\n", currentStableTGsWeight, desiredStableTGsWeight)

			// Do update by step weight

			if desiredStableTGsWeight > 0 {
				numStableTGs := len(stableTGs)

				updatedStableTGs := map[string]okrav1alpha1.ForwardTargetGroup{}

				for i, tg := range stableTGs {
					tg := tg

					weight := desiredStableTGsWeight / numStableTGs

					if i == numStableTGs-1 && numStableTGs > 1 {
						weight = desiredStableTGsWeight - (weight * (numStableTGs - 1))
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

			canaryTGsWeight = 100 - desiredStableTGsWeight

			if canaryTGsWeight > 0 {
				var canaryVersion string
				for _, tg := range latestTGs {
					for _, l := range labelKeys {
						v, ok := tg.Labels[l]
						if ok {
							canaryVersion = v
							break
						}
					}
				}
				log.Printf("canary(%s) weight: %d -> %d\n", canaryVersion, currentCanaryTGsWeight, canaryTGsWeight)

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

		updated := make(map[string]int)
		for _, tg := range updatedTGs {
			updated[tg.Name] = tg.Weight
		}

		log.Printf("updating target groups and weights to: %v\n", updated)

		albConfig.Spec.Listener.Rule.Forward.TargetGroups = updatedTGs

		if err := runtimeClient.Update(ctx, &albConfig); err != nil {
			return err
		}

		if desiredStableTGsWeight == 0 && passedAllCanarySteps {
			// Seems like we need to explicitly specify the namespace with client.InNamespace.
			// Otherwise it results in `Error: the server could not find the requested resource (delete analysisruns.argoproj.io)`
			if err := runtimeClient.DeleteAllOf(ctx, &rolloutsv1alpha1.AnalysisRun{}, client.InNamespace(cell.Namespace), &client.DeleteAllOfOptions{
				ListOptions: client.ListOptions{
					LabelSelector: ownedByCellLabelSelector,
				},
			}); err != nil {
				log.Printf("Failed deleting analysis runs: %v", err)
				return err
			}

			log.Printf("Deleted all analysis runs with %s, if any", ownedByCellLabelSelector)

			if err := runtimeClient.DeleteAllOf(ctx, &okrav1alpha1.Pause{}, client.InNamespace(cell.Namespace), &client.DeleteAllOfOptions{
				ListOptions: client.ListOptions{
					LabelSelector: ownedByCellLabelSelector,
				},
			}); err != nil {
				return err
			}

			log.Printf("Deleted all pauses with %s as completed, if any", ownedByCellLabelSelector)
		}
	}

	return nil
}
