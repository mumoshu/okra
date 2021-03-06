package cell

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/sync"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type cellComponentReconciler struct {
	cell          okrav1alpha1.Cell
	runtimeClient client.Client
	scheme        *runtime.Scheme
	cellStateHash string
}

type componentReconcilationResult int

const (
	ComponentInProgress componentReconcilationResult = iota
	ComponentPassed
	ComponentFailed
)

func (s cellComponentReconciler) componentSelectorLabels(componentID string) map[string]string {
	return map[string]string{
		LabelKeyStepIndex:     componentID,
		LabelKeyCell:          s.cell.Name,
		LabelKeyCellStateHash: s.cellStateHash,
	}
}

func (s cellComponentReconciler) outdatedComponentSelectorLabels() (labels.Selector, error) {
	return labels.Parse(LabelKeyCell + "=" + s.cell.Name + "," + LabelKeyCellStateHash + "!=" + s.cellStateHash)
}

func (s cellComponentReconciler) componentLabels(componentID, templateHash string) map[string]string {
	r := s.componentSelectorLabels(componentID)
	r[LabelKeyTemplateHash] = templateHash
	return r
}

func (s cellComponentReconciler) reconcileAnalysisRun(ctx context.Context, componentID string, analysis *rolloutsv1alpha1.RolloutAnalysis, validateT func(rolloutsv1alpha1.AnalysisTemplate) error) (componentReconcilationResult, error) {
	cell := s.cell
	runtimeClient := s.runtimeClient
	scheme := s.scheme

	//
	// Ensure that the previous analysis run has been successful, if any
	//

	var analysisRunList rolloutsv1alpha1.AnalysisRunList

	labelSelector, err := labels.Parse(LabelKeyStepIndex + "=" + componentID)
	if err != nil {
		return ComponentInProgress, err
	}

	if err := runtimeClient.List(ctx, &analysisRunList, &client.ListOptions{
		LabelSelector: labelSelector,
	}); err != nil {
		return ComponentInProgress, err
	}

	switch len(analysisRunList.Items) {
	case 0:
		tmpl := analysis.Templates[0]

		var args []rolloutsv1alpha1.Argument
		argsMap := make(map[string]rolloutsv1alpha1.Argument)

		var at rolloutsv1alpha1.AnalysisTemplate
		nsName := types.NamespacedName{Namespace: cell.Namespace, Name: tmpl.TemplateName}
		if err := runtimeClient.Get(ctx, nsName, &at); err != nil {
			log.Printf("Failed getting analysistemplate %s: %v", nsName, err)
			return ComponentInProgress, err
		}

		if validateT != nil {
			// Validation
			if err := validateT(at); err != nil {
				return ComponentFailed, err
			}
		}

		for _, a := range at.Spec.Args {
			argsMap[a.Name] = *a.DeepCopy()
		}

		for _, a := range analysis.Args {
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
				} else if a.ValueFrom != nil && arg.ValueFrom.FieldRef != nil {
					v, err := extractValueFromCell(&cell, a.ValueFrom.FieldRef.FieldPath)
					if err != nil {
						return ComponentFailed, err
					}
					arg.Value = &v
				}

				argsMap[a.Name] = arg
			}
		}

		for _, a := range argsMap {
			args = append(args, a)
		}

		spec := rolloutsv1alpha1.AnalysisRunSpec{
			Args:    args,
			Metrics: at.Spec.Metrics,
		}

		templateHash := sync.ComputeHash(spec)

		ar := rolloutsv1alpha1.AnalysisRun{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cell.Namespace,
				Name:      fmt.Sprintf("%s-%s-%s", cell.Name, componentID, tmpl.TemplateName),
				Labels:    s.componentLabels(componentID, templateHash),
			},
			Spec: spec,
		}
		if err := ctrl.SetControllerReference(&cell, &ar, scheme); err != nil {
			log.Printf("Failed setting controller reference on %s/%s: %v", ar.Namespace, ar.Name, err)
		}

		if err := runtimeClient.Create(ctx, &ar); err != nil {
			return ComponentInProgress, err
		}

		log.Printf("Created analysisrun %s", ar.Name)

		return ComponentInProgress, nil
	case 1:
		ar := analysisRunList.Items[0]

		switch ar.Status.Phase {
		case rolloutsv1alpha1.AnalysisPhaseError, rolloutsv1alpha1.AnalysisPhaseFailed:
			log.Printf("AnalysisRun %s failed with error: %v", ar.Name, ar.Status.Message)

			return ComponentFailed, nil
		case rolloutsv1alpha1.AnalysisPhaseSuccessful:
		default:
			log.Printf("Waiting for analysisrun %s of %s to become %s", ar.Name, ar.Status.Phase, rolloutsv1alpha1.AnalysisPhaseSuccessful)

			// We need to wait for this analysis run to succeed
			return ComponentInProgress, nil
		}
	default:
		return ComponentInProgress, errors.New("too many analysis runs")
	}

	return ComponentPassed, nil
}

func (s cellComponentReconciler) reconcileExperiment(ctx context.Context, componentID string, exTemplate *rolloutsv1alpha1.RolloutExperimentStep) (componentReconcilationResult, error) {
	//
	// Ensure that the previous experiments has been successful, if any
	//

	runtimeClient := s.runtimeClient
	scheme := s.scheme
	cell := s.cell

	var experimentList rolloutsv1alpha1.ExperimentList

	labelSelector, err := labels.Parse(LabelKeyStepIndex + "=" + componentID)
	if err != nil {
		return ComponentInProgress, err
	}

	if err := runtimeClient.List(ctx, &experimentList, &client.ListOptions{
		LabelSelector: labelSelector,
	}); err != nil {
		return ComponentInProgress, err
	}

	numExperiments := len(experimentList.Items)

	var ex rolloutsv1alpha1.Experiment
	{

		d := exTemplate.Duration

		var templates []rolloutsv1alpha1.TemplateSpec

		for _, t := range exTemplate.Templates {
			var rs appsv1.ReplicaSet
			nsName := types.NamespacedName{Namespace: cell.Namespace, Name: string(t.SpecRef)}
			if err := runtimeClient.Get(ctx, nsName, &rs); err != nil {
				log.Printf("Failed getting experiment template replicaset %s: %v", nsName, err)
				return ComponentInProgress, err
			}

			s := t.Selector
			if s == nil {
				s = rs.Spec.Selector
			}

			templates = append(templates, rolloutsv1alpha1.TemplateSpec{
				Name:     t.Name,
				Replicas: t.Replicas,
				Selector: s,
				Template: rs.Spec.Template,
			})
		}

		var analyses []rolloutsv1alpha1.ExperimentAnalysisTemplateRef
		for _, a := range exTemplate.Analyses {
			var args []rolloutsv1alpha1.Argument
			for _, arg := range a.Args {
				var (
					value *string
				)

				if arg.Value != "" {
					value = &arg.Value
				} else if arg.ValueFrom != nil && arg.ValueFrom.FieldRef != nil {
					v, err := extractValueFromCell(&cell, arg.ValueFrom.FieldRef.FieldPath)
					if err != nil {
						return ComponentFailed, err
					}
					value = &v
				}

				args = append(args, rolloutsv1alpha1.Argument{
					Name:  arg.Name,
					Value: value,
				})
			}
			analyses = append(analyses, rolloutsv1alpha1.ExperimentAnalysisTemplateRef{
				Name:                  a.Name,
				TemplateName:          a.TemplateName,
				Args:                  args,
				RequiredForCompletion: a.RequiredForCompletion,
			})
		}

		spec := rolloutsv1alpha1.ExperimentSpec{
			Duration:  d,
			Templates: templates,
			Analyses:  analyses,
		}

		templateHash := sync.ComputeHash(spec)

		ex = rolloutsv1alpha1.Experiment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cell.Namespace,
				Name:      fmt.Sprintf("%s-%s-%s", cell.Name, componentID, "experiment"),
				Labels:    s.componentLabels(componentID, templateHash),
			},
			Spec: spec,
		}
		if err := ctrl.SetControllerReference(&cell, &ex, scheme); err != nil {
			log.Printf("Failed setting controller reference on %s/%s: %v", ex.Namespace, ex.Name, err)
		}
	}

	if numExperiments == 0 {
		if err := runtimeClient.Create(ctx, &ex); err != nil {
			return ComponentInProgress, err
		}

		log.Printf("Created experiment %s", ex.Name)

		return ComponentInProgress, nil
	}

	if numExperiments > 1 {
		return ComponentInProgress, errors.New("too many experiments")
	}

	var currentTemplateHash string
	if annotations := experimentList.Items[0].GetAnnotations(); annotations != nil {
		if templateHash := annotations[LabelKeyTemplateHash]; templateHash != "" {
			currentTemplateHash = templateHash
		}
	}

	var desiredTemplateHash string
	if annotations := ex.GetAnnotations(); annotations != nil {
		if templateHash := annotations[LabelKeyTemplateHash]; templateHash != "" {
			desiredTemplateHash = templateHash
		}
	}

	if currentTemplateHash != desiredTemplateHash {
		var current rolloutsv1alpha1.Experiment

		if err := runtimeClient.Get(ctx, types.NamespacedName{Namespace: ex.Namespace, Name: ex.Name}, &current); err != nil {
			return ComponentInProgress, err
		}

		current.Spec = ex.Spec

		for k, v := range ex.Labels {
			current.Labels[k] = v
		}

		for k, v := range ex.Annotations {
			current.Annotations[k] = v
		}

		if err := runtimeClient.Update(ctx, &current); err != nil {
			return ComponentInProgress, err
		}

		log.Printf("Updated experiment %s", ex.Name)

		return ComponentInProgress, nil
	}

	unchangedEx := experimentList.Items[0]

	switch unchangedEx.Status.Phase {
	case rolloutsv1alpha1.AnalysisPhaseSuccessful:
	case rolloutsv1alpha1.AnalysisPhaseError, rolloutsv1alpha1.AnalysisPhaseFailed:
		log.Printf("Experiment %s failed with error: %v", ex.Name, ex.Status.Message)

		return ComponentFailed, nil
	default:
		log.Printf("Waiting for experiment %s of %s to become %s", ex.Name, ex.Status.Phase, rolloutsv1alpha1.AnalysisPhaseSuccessful)

		// We need to wait for this analysis run to succeed
		return ComponentInProgress, nil
	}

	return ComponentPassed, nil
}

func (s cellComponentReconciler) reconcilePause(ctx context.Context, componentID string, pauseTemplate *rolloutsv1alpha1.RolloutPause) (componentReconcilationResult, error) {
	// TODO List Pause resource and break if it isn't expired yet
	var pauseList okrav1alpha1.PauseList

	cell := s.cell
	runtimeClient := s.runtimeClient
	scheme := s.scheme

	t := metav1.Time{
		Time: time.Now().Add(time.Duration(time.Second.Nanoseconds() * int64(pauseTemplate.DurationSeconds()))),
	}

	labels := s.componentSelectorLabels(componentID)

	ns := cell.Namespace

	if err := runtimeClient.List(ctx, &pauseList, client.InNamespace(ns), client.MatchingLabels(labels)); err != nil {
		return ComponentInProgress, err
	}

	switch c := len(pauseList.Items); c {
	case 0:
		spec := okrav1alpha1.PauseSpec{
			ExpireTime: t,
		}

		templateHash := sync.ComputeHash(spec)

		pause := okrav1alpha1.Pause{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      fmt.Sprintf("%s-%s-%s", cell.Name, componentID, "pause"),
				Labels:    s.componentLabels(componentID, templateHash),
			},
			Spec: spec,
		}
		ctrl.SetControllerReference(&cell, &pause, scheme)

		if err := runtimeClient.Create(ctx, &pause); err != nil {
			return ComponentInProgress, err
		}

		log.Printf("Initiated pause %s until %s", pause.Name, t)

		return ComponentInProgress, nil
	case 1:
		pause := pauseList.Items[0]

		switch phase := pause.Status.Phase; phase {
		case okrav1alpha1.PausePhaseCancelled:
			log.Printf("Observed that pause %s had been cancelled. Continuing to the next step", pause.Name)
		case okrav1alpha1.PausePhaseExpired:
			log.Printf("Observed that pause %s had expired. Continuing to the next step", pause.Name)
		case okrav1alpha1.PausePhaseStarted:
			log.Printf("Still waiting for pause %s to expire or get cancelled", pause.Name)
			return ComponentInProgress, nil
		case "":
			log.Printf("Still waiting for pause %s to start", pause.Name)
			return ComponentInProgress, nil
		default:
			return ComponentFailed, fmt.Errorf("unexpected pause phase: %s", phase)
		}
	default:
		return ComponentInProgress, fmt.Errorf("unexpected number of pauses found: %d", c)
	}

	return ComponentPassed, nil
}
