package cell

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/blang/semver"
	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/awstargetgroupset"
	"github.com/mumoshu/okra/pkg/clclient"
	"github.com/mumoshu/okra/pkg/sync"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LabelKeyStepIndex    = "okra.mumo.co/step-index"
	LabelKeyTemplateHash = "okra.mumo.co/template-hash"
	LabelKeyCell         = "cell"
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

	key := types.NamespacedName{Namespace: cell.Namespace, Name: cell.Name}

	albListenerARN := cell.Spec.Ingress.AWSApplicationLoadBalancer.ListenerARN
	tgSelectorMatchLabels := cell.Spec.Ingress.AWSApplicationLoadBalancer.TargetGroupSelector
	tgSelector := labels.SelectorFromSet(tgSelectorMatchLabels.MatchLabels)

	var albConfig okrav1alpha1.AWSApplicationLoadBalancerConfig
	var albConfigExists bool
	var desiredALBConfigSpec okrav1alpha1.AWSApplicationLoadBalancerConfigSpec

	if err := runtimeClient.Get(ctx, key, &albConfig); err != nil {
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

	const LabelKeyALBConfigHash = "alb-config-hash"

	desiredALBConfigSpec.Listener = cell.Spec.Ingress.AWSApplicationLoadBalancer.Listener
	desiredALBConfigSpec.ListenerARN = albListenerARN
	desiredALBConfigSpecHash := sync.ComputeHash(desiredALBConfigSpec)
	currentALBConfigSpecHash := albConfig.Annotations[LabelKeyALBConfigHash]

	labelKeys := cell.Spec.Ingress.AWSApplicationLoadBalancer.TargetGroupSelector.VersionLabels
	if len(labelKeys) == 0 {
		labelKeys = []string{okrav1alpha1.DefaultVersionLabelKey}
	}

	v := cell.Spec.Version

	desiredVer, desiredTGs, err := awstargetgroupset.ListLatestAWSTargetGroups(awstargetgroupset.ListLatestAWSTargetGroupsInput{
		ListAWSTargetGroupsInput: awstargetgroupset.ListAWSTargetGroupsInput{
			NS:       cell.Namespace,
			Selector: tgSelector.String(),
		},
		SemverLabelKeys: labelKeys,
		Version:         v,
	})
	if err != nil {
		return err
	}

	if v != "" {
		log.Printf("Using cell.Spec.Version(%s) instead of latest version", v)
	}

	allKnownTGs, err := awstargetgroupset.ListAWSTargetGroups(awstargetgroupset.ListAWSTargetGroupsInput{
		NS:       cell.Namespace,
		Selector: tgSelector.String(),
	})
	if err != nil {
		return err
	}

	allKnownTGsNameToVer := make(map[string]string)
	for _, tg := range allKnownTGs {
		var ver string
		for _, l := range labelKeys {
			v, ok := tg.Labels[l]
			if ok {
				ver = v
				break
			}
		}

		if ver != "" {
			allKnownTGsNameToVer[tg.Name] = ver
		}
	}

	desiredTGsByName := map[string]okrav1alpha1.ForwardTargetGroup{}

	numLatestTGs := len(desiredTGs)

	// Ensure there enough cluster replicas to start a canary release
	threshold := 1
	if cell.Spec.Replicas != nil {
		threshold = int(*cell.Spec.Replicas)
	}

	log.Printf("key=%s, albConfigExists=%v, tgSelector=%s, len(latestTGs)=%d\n", key, albConfigExists, tgSelector.String(), len(desiredTGs))

	if numLatestTGs != threshold {
		return nil
	}

	// Do distribute weights evently so that the total becomes 100
	for i, tg := range desiredTGs {
		weight := 100 / numLatestTGs

		if i == numLatestTGs-1 && numLatestTGs > 1 {
			weight = 100 - (weight * (numLatestTGs - 1))
		}

		desiredTGsByName[tg.Name] = okrav1alpha1.ForwardTargetGroup{
			Name:   tg.Name,
			ARN:    tg.Spec.ARN,
			Weight: weight,
		}
	}

	if !albConfigExists {
		// ALB isn't initialized yet so we are creating the ALBConfig resource for the first time
		for _, tg := range desiredTGsByName {
			albConfig.Spec.Listener.Rule.Forward.TargetGroups = append(albConfig.Spec.Listener.Rule.Forward.TargetGroups, tg)
		}

		metav1.SetMetaDataAnnotation(&albConfig.ObjectMeta, LabelKeyALBConfigHash, desiredALBConfigSpecHash)

		if err := runtimeClient.Create(ctx, &albConfig); err != nil {
			return fmt.Errorf("creating albconfig: %w", err)
		}

		updated := make(map[string]int)
		for _, tg := range desiredTGsByName {
			updated[tg.Name] = tg.Weight
		}

		log.Printf("Created target groups and weights to: %v", updated)

		return nil
	}

	if currentALBConfigSpecHash != desiredALBConfigSpecHash {
		metav1.SetMetaDataAnnotation(&albConfig.ObjectMeta, LabelKeyALBConfigHash, desiredALBConfigSpecHash)

		albConfig.Spec = desiredALBConfigSpec

		if err := runtimeClient.Update(ctx, &albConfig); err != nil {
			return fmt.Errorf("updating albconfig: %w", err)
		}

		return nil
	}

	// This is a standard cell update for releasing a new app/cluster version.
	// Do a canary release.

	// Ensure that the previous analysis run has been successful, if any

	var currentStableTGsWeight, currentCanaryTGsWeight, canaryTGsWeight int

	var (
		currentStableTGs []okrav1alpha1.ForwardTargetGroup
		currentCanaryTGs []okrav1alpha1.ForwardTargetGroup
	)

	for _, tg := range albConfig.Spec.Listener.Rule.Forward.TargetGroups {
		tg := tg

		if _, ok := desiredTGsByName[tg.Name]; ok {
			currentCanaryTGsWeight += tg.Weight
			currentCanaryTGs = append(currentCanaryTGs, tg)
			continue
		}

		currentStableTGs = append(currentStableTGs, tg)

		currentStableTGsWeight += tg.Weight
	}

	var (
		rollbackRequested bool
	)

	for _, tg := range currentStableTGs {
		ver := allKnownTGsNameToVer[tg.Name]

		currentVer, err := semver.Parse(ver)
		if err != nil {
			log.Printf("Skipped incorrect label value %s: %v", ver, err)
			continue
		}

		if desiredVer.LT(currentVer) {
			rollbackRequested = true
		}
	}

	var currentStableTGsMaxVer *semver.Version
	currentStableTGsByVer := map[string][]okrav1alpha1.ForwardTargetGroup{}
	for _, tg := range currentStableTGs {
		ver := allKnownTGsNameToVer[tg.Name]

		currentVer, err := semver.Parse(ver)
		if err != nil {
			log.Printf("Skipped incorrect label value %s: %v", ver, err)
			continue
		}

		if currentStableTGsMaxVer == nil || currentStableTGsMaxVer.LT(currentVer) {
			currentStableTGsMaxVer = &currentVer
		}

		currentStableTGsByVer[currentVer.String()] = append(currentStableTGsByVer[currentVer.String()], tg)
	}

	if currentStableTGsMaxVer != nil {
		currentStableTGs = currentStableTGsByVer[currentStableTGsMaxVer.String()]
	}

	// Do update immediately without analysis or step update when
	// it seems to have been triggered by an additional cluster that might have been
	// added to deal with more load.
	scaleRequested := len(currentCanaryTGs) > 0 && len(desiredTGsByName) != len(currentCanaryTGs) && currentCanaryTGsWeight == 100

	if rollbackRequested || scaleRequested {
		// Immediately update LB config as quickly as possible when
		// either a rollback or a scale in/out is requested.

		albConfig.Spec.Listener.Rule.Forward.TargetGroups = nil
		for _, tg := range desiredTGsByName {
			albConfig.Spec.Listener.Rule.Forward.TargetGroups = append(albConfig.Spec.Listener.Rule.Forward.TargetGroups, tg)
		}
		for _, tg := range currentStableTGs {
			tg.Weight = 0
			albConfig.Spec.Listener.Rule.Forward.TargetGroups = append(albConfig.Spec.Listener.Rule.Forward.TargetGroups, tg)
		}

		if err := runtimeClient.Update(ctx, &albConfig); err != nil {
			return fmt.Errorf("updating albconfig: %w", err)
		}

		updated := make(map[string]int)
		for _, tg := range desiredTGsByName {
			updated[tg.Name] = tg.Weight
		}

		log.Printf("Updated target groups and weights to: %v", updated)

		if rollbackRequested {
			log.Printf("Finished rollback")
		} else {
			log.Printf("Finished scaling")
		}

		return nil
	}

	var (
		passedAllCanarySteps bool
		anyStepFailed        bool
		desiredVerIsBlocked  bool
	)

	// TODO Use client.MatchingLabels?
	everythingOwnedByThisCell, err := labels.Parse(LabelKeyCell + "=" + cell.Name)
	if err != nil {
		return err
	}

	desiredStableTGsWeight := 100

	var bl okrav1alpha1.VersionBlocklist

	if err := runtimeClient.Get(ctx, key, &bl); err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		}
	}

	for _, item := range bl.Spec.Items {
		if item.Version == desiredVer.String() {
			desiredVerIsBlocked = true
			break
		}
	}

	if !desiredVerIsBlocked {
		canary := cell.Spec.UpdateStrategy.Canary
		canarySteps := canary.Steps

		passedAllCanarySteps = currentCanaryTGsWeight == 100

		if len(canarySteps) > 0 && !passedAllCanarySteps {
			var analysisRunList rolloutsv1alpha1.AnalysisRunList

			if err := runtimeClient.List(ctx, &analysisRunList, &client.ListOptions{
				LabelSelector: everythingOwnedByThisCell,
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

			ccr := cellComponentReconciler{
				cell:          cell,
				runtimeClient: runtimeClient,
				scheme:        scheme,
			}

		STEPS:
			for stepIndex, step := range canarySteps {
				stepIndexStr := strconv.Itoa(stepIndex)

				if a := canary.Analysis; a != nil {
					// A background analysis works very much like
					// Argo Rollouts Background Analysis as documented at
					// https://argoproj.github.io/argo-rollouts/features/analysis/#background-analysis
					// except that okra's works against clusters(backing e.g. AWSTargetGroups) instead of replicasets.

					start := int32(0)
					if a.StartingStep != nil {
						start = *a.StartingStep
					}

					if int32(stepIndex) >= start {
						r, err := ccr.reconcileAnalysisRun(ctx, "bg", &a.RolloutAnalysis)
						if err != nil {
							return err
						} else if r == ComponentFailed {
							anyStepFailed = true
							break STEPS
						}

						// We accept both StepInProgress and StepPassed
						// as a background analysis makes the cell degraded
						// only if it failed.
					}
				}

				if step.Analysis != nil {
					r, err := ccr.reconcileAnalysisRun(ctx, stepIndexStr, step.Analysis)
					if err != nil {
						return err
					} else if r == ComponentInProgress {
						break STEPS
					} else if r == ComponentFailed {
						anyStepFailed = true
						break STEPS
					}
				} else if step.Experiment != nil {
					r, err := ccr.reconcileExperiment(ctx, stepIndexStr, step.Experiment)
					if err != nil {
						return err
					} else if r == ComponentInProgress {
						break STEPS
					} else if r == ComponentFailed {
						anyStepFailed = true
						break STEPS
					}
				} else if step.SetWeight != nil {
					desiredStableTGsWeight -= int(*step.SetWeight)

					if desiredStableTGsWeight < currentStableTGsWeight {
						break STEPS
					}
				} else if step.Pause != nil {
					// TODO List Pause resource and break if it isn't expired yet
					var pauseList okrav1alpha1.PauseList

					ns := cell.Namespace

					labels := map[string]string{
						LabelKeyStepIndex: stepIndexStr,
						LabelKeyCell:      cell.Name,
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
								Name:      fmt.Sprintf("%s-%d-%s", cell.Name, stepIndex, "pause"),
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

		if anyStepFailed {
			desiredStableTGsWeight = 100
		}

		if desiredStableTGsWeight < 0 {
			return fmt.Errorf("stable tgs weight cannot be less than 0: %v", desiredStableTGsWeight)
		}

		log.Printf("stable weight(%v): %d -> %d\n", currentStableTGsMaxVer, currentStableTGsWeight, desiredStableTGsWeight)

		// Do update by step weight
		var updatedTGs []okrav1alpha1.ForwardTargetGroup

		numStableTGs := len(currentStableTGs)

		updatedStableTGs := map[string]okrav1alpha1.ForwardTargetGroup{}

		for i, tg := range currentStableTGs {
			tg := tg

			var weight int

			if desiredStableTGsWeight > 0 {
				weight = desiredStableTGsWeight / numStableTGs

				if i == numStableTGs-1 && numStableTGs > 1 {
					weight = desiredStableTGsWeight - (weight * (numStableTGs - 1))
				}
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

		canaryTGsWeight = 100 - desiredStableTGsWeight

		var canaryVersion string
		for _, tg := range desiredTGs {
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

		for i, tg := range desiredTGs {
			var weight int

			if canaryTGsWeight > 0 {
				weight = canaryTGsWeight / numLatestTGs

				if i == numLatestTGs-1 && numLatestTGs > 1 {
					weight = canaryTGsWeight - (weight * (numLatestTGs - 1))
				}
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

		sort.Slice(updatedTGs, func(i, j int) bool {
			return updatedTGs[i].Name < updatedTGs[j].Name
		})

		updated := make(map[string]int)
		for _, tg := range updatedTGs {
			updated[tg.Name] = tg.Weight
		}

		albConfig.Spec.Listener.Rule.Forward.TargetGroups = updatedTGs

		currentHash := albConfig.Annotations[LabelKeyTemplateHash]
		desiredHash := sync.ComputeHash(albConfig.Spec)

		if currentHash != desiredHash {
			metav1.SetMetaDataAnnotation(&albConfig.ObjectMeta, LabelKeyTemplateHash, desiredHash)

			if err := runtimeClient.Update(ctx, &albConfig); err != nil {
				return err
			}

			log.Printf("Updated target groups and weights to: %v\n", updated)
		} else {
			log.Printf("Skipped updating target groups")
		}
	}

	if anyStepFailed {
		var bl okrav1alpha1.VersionBlocklist

		item := okrav1alpha1.VersionBlocklistItem{
			Version: desiredVer.String(),
			Cause:   "AnalysisRun failed",
		}

		if err := runtimeClient.Get(ctx, types.NamespacedName{Namespace: cell.Namespace, Name: cell.Name}, &bl); err != nil {
			if !kerrors.IsNotFound(err) {
				return err
			}

			bl = okrav1alpha1.VersionBlocklist{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cell.Namespace,
					Name:      cell.Name,
				},
				Spec: okrav1alpha1.VersionBlocklistSpec{
					Items: []okrav1alpha1.VersionBlocklistItem{
						item,
					},
				},
			}
			if err := runtimeClient.Create(ctx, &bl); err != nil {
				return err
			}
		} else {
			bl.Spec.Items = append(bl.Spec.Items, item)

			if err := runtimeClient.Update(ctx, &bl); err != nil {
				return err
			}
		}
	}

	log.Printf("Finishing reconcilation. desiredTargetTGsWeight=%v, passedAllCanarySteps=%v, anyStepFailed=%v, desiredVerIsBlocked=%v", desiredStableTGsWeight, passedAllCanarySteps, anyStepFailed, desiredVerIsBlocked)

	if desiredStableTGsWeight == 0 && passedAllCanarySteps || anyStepFailed || desiredVerIsBlocked {
		objects := []runtime.Object{
			&rolloutsv1alpha1.AnalysisRun{},
			&rolloutsv1alpha1.Experiment{},
			&okrav1alpha1.Pause{},
		}

		for _, o := range objects {
			// Seems like we need to explicitly specify the namespace with client.InNamespace.
			// Otherwise it results in `Error: the server could not find the requested resource (delete analysisruns.argoproj.io)`
			if err := runtimeClient.DeleteAllOf(ctx, o, client.InNamespace(cell.Namespace), &client.DeleteAllOfOptions{
				ListOptions: client.ListOptions{
					LabelSelector: everythingOwnedByThisCell,
				},
			}); err != nil {
				log.Printf("Failed deleting %Ts: %v", o, err)
				return err
			}

			log.Printf("Deleted all %Ts with %s, if any", o, everythingOwnedByThisCell)
		}
	}

	return nil
}
