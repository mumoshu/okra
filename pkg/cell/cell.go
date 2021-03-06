package cell

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"

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
	LabelKeyStepIndex     = "okra.mumo.co/step-index"
	LabelKeyTemplateHash  = "okra.mumo.co/template-hash"
	LabelKeyCellStateHash = "okra.mumo.co/cell-state-hash"
	LabelKeyCell          = "cell"
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

	sort.Slice(desiredTGs, func(i, j int) bool {
		return desiredTGs[i].Name < desiredTGs[j].Name
	})

	// We use this to clean up outdated analysisruns, experiments, and pauses
	cellStateHash := sync.ComputeHash(desiredTGs)

	// Do distribute weights evently so that the total becomes 100
	desiredTGsByName := distributeWeights(100, desiredTGs)

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

	var currentStableTGsWeight, currentCanaryTGsWeight, desiredCanaryTGsWeight int

	var (
		currentCanaryTGs  []okrav1alpha1.ForwardTargetGroup
		currentStableTGs  []okrav1alpha1.ForwardTargetGroup
		rollbackRequested bool

		currentStableTGsMaxVer *semver.Version
		currentStableTGsByVer  = map[string][]okrav1alpha1.ForwardTargetGroup{}
	)

	for _, tg := range albConfig.Spec.Listener.Rule.Forward.TargetGroups {
		// Divide target groups already registered to our ALB config
		// between canary and stable versions, which are necessary for a gradual update.

		tg := tg

		if _, ok := desiredTGsByName[tg.Name]; ok {
			currentCanaryTGsWeight += tg.Weight
			currentCanaryTGs = append(currentCanaryTGs, tg)
			continue
		}

		currentStableTGs = append(currentStableTGs, tg)

		currentStableTGsWeight += tg.Weight

		// Check if rollback is requested
		ver := allKnownTGsNameToVer[tg.Name]

		currentVer, err := semver.Parse(ver)
		if err != nil {
			log.Printf("Skipped incorrect label value %s: %v", ver, err)
			continue
		}

		if desiredVer.LT(currentVer) {
			rollbackRequested = true
		}

		// Make sure there's only one stable version by
		// stripping out all the older stable versions
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

	if desiredVerIsBlocked {
		log.Printf("Version %s is blocked. Please specify another version that is not blocked to start a rollout.", desiredVer)
		return nil
	}

	// Now, we need to update cell.status
	// so that values in it can be used from within field paths
	// contained in experiment and analysis step args.
	cell.Status.DesiredVersion = desiredVer.String()

	canary := cell.Spec.UpdateStrategy.Canary
	canarySteps := canary.Steps

	passedAllCanarySteps = currentCanaryTGsWeight == 100

	desiredStableTGsWeight := 100

	if len(canarySteps) > 0 && !passedAllCanarySteps {
		var analysisRunList rolloutsv1alpha1.AnalysisRunList

		if err := runtimeClient.List(ctx, &analysisRunList, &client.ListOptions{
			LabelSelector: everythingOwnedByThisCell,
		}); err != nil {
			return err
		}

		ccr := cellComponentReconciler{
			cell:          cell,
			runtimeClient: runtimeClient,
			scheme:        scheme,
			cellStateHash: cellStateHash,
		}

		objects := []runtime.Object{
			&rolloutsv1alpha1.AnalysisRun{},
			&rolloutsv1alpha1.Experiment{},
			&okrav1alpha1.Pause{},
		}

		outdatedComponents, err := ccr.outdatedComponentSelectorLabels()
		if err != nil {
			return err
		}

		for _, o := range objects {
			// Seems like we need to explicitly specify the namespace with client.InNamespace.
			// Otherwise it results in `Error: the server could not find the requested resource (delete analysisruns.argoproj.io)`
			if err := runtimeClient.DeleteAllOf(ctx, o, client.InNamespace(cell.Namespace), &client.DeleteAllOfOptions{
				ListOptions: client.ListOptions{
					LabelSelector: outdatedComponents,
				},
			}); err != nil {
				log.Printf("Failed deleting %Ts: %v", o, err)
				return err
			}

			log.Printf("Deleted all %Ts with %s, if any", o, outdatedComponents)
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
					r, err := ccr.reconcileAnalysisRun(ctx, "bg", &a.RolloutAnalysis, nil)
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

			var (
				r   componentReconcilationResult
				err error
			)

			if step.Analysis != nil {
				r, err = ccr.reconcileAnalysisRun(ctx, stepIndexStr, step.Analysis, func(at rolloutsv1alpha1.AnalysisTemplate) error {
					for _, m := range at.Spec.Metrics {
						if d, _ := m.Interval.Duration(); d > 0 && m.Count == nil {
							return fmt.Errorf("analysistemplate %s: metric %s: step analysis should have non-zero count", at.Name, m.Name)
						}
					}
					return nil
				})
			} else if step.Experiment != nil {
				r, err = ccr.reconcileExperiment(ctx, stepIndexStr, step.Experiment)
			} else if step.SetWeight != nil {
				desiredStableTGsWeight -= int(*step.SetWeight)

				r = ComponentPassed
			} else if step.Pause != nil {
				r, err = ccr.reconcilePause(ctx, stepIndexStr, step.Pause)
			} else {
				return fmt.Errorf("steps[%d]: only setWeight, analysis, and pause step are supported. got %v", stepIndex, step)
			}

			if err != nil {
				return err
			} else if r == ComponentInProgress {
				break STEPS
			} else if r == ComponentFailed {
				anyStepFailed = true
				break STEPS
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

	// Do update by step weight
	var updatedTGs []okrav1alpha1.ForwardTargetGroup

	updatedStableTGs := redistributeWeights(desiredStableTGsWeight, currentStableTGs)

	for _, tg := range updatedStableTGs {
		updatedTGs = append(updatedTGs, tg)
	}

	desiredCanaryTGsWeight = 100 - desiredStableTGsWeight

	updatedCanaryTGsByName := distributeWeights(desiredCanaryTGsWeight, desiredTGs)

	for _, tg := range updatedCanaryTGsByName {
		updatedTGs = append(updatedTGs, tg)
	}

	sort.Slice(updatedTGs, func(i, j int) bool {
		return updatedTGs[i].Name < updatedTGs[j].Name
	})

	albConfig.Spec.Listener.Rule.Forward.TargetGroups = updatedTGs

	currentHash := albConfig.Annotations[LabelKeyTemplateHash]
	desiredHash := sync.ComputeHash(albConfig.Spec)

	if currentHash != desiredHash {
		if currentStableTGsWeight != desiredStableTGsWeight {
			log.Printf("Changing stable weight(%v): %d -> %d\n", currentStableTGsMaxVer, currentStableTGsWeight, desiredStableTGsWeight)
		}
		if currentCanaryTGsWeight != desiredCanaryTGsWeight {
			log.Printf("Changing canary(%s) weight: %d -> %d\n", desiredVer, currentCanaryTGsWeight, desiredCanaryTGsWeight)
		}

		metav1.SetMetaDataAnnotation(&albConfig.ObjectMeta, LabelKeyTemplateHash, desiredHash)

		if err := runtimeClient.Update(ctx, &albConfig); err != nil {
			return err
		}

		updated := make(map[string]int)
		for _, tg := range updatedTGs {
			updated[tg.Name] = tg.Weight
		}

		log.Printf("Updated target groups and weights to: %v\n", updated)
	} else {
		log.Printf("No change detected on AWSApplicationLoadBalancerConfig and target group weights. Skipped updating.")
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

	return nil
}
