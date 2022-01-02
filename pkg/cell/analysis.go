package cell

import (
	"context"
	"errors"
	"fmt"
	"log"

	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
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
}

type stepReconcileResult int

const (
	StepInProgress stepReconcileResult = iota
	StepPassed
	StepFailed
)

func (s cellComponentReconciler) reconcileAnalysisRun(ctx context.Context, componentID string, analysis *rolloutsv1alpha1.RolloutAnalysis) (stepReconcileResult, error) {
	cell := s.cell
	runtimeClient := s.runtimeClient
	scheme := s.scheme

	//
	// Ensure that the previous analysis run has been successful, if any
	//

	var analysisRunList rolloutsv1alpha1.AnalysisRunList

	labelSelector, err := labels.Parse(LabelKeyStepIndex + "=" + componentID)
	if err != nil {
		return StepInProgress, err
	}

	if err := runtimeClient.List(ctx, &analysisRunList, &client.ListOptions{
		LabelSelector: labelSelector,
	}); err != nil {
		return StepInProgress, err
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
			return StepInProgress, err
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
				}

				argsMap[a.Name] = arg
			}
		}

		for _, a := range argsMap {
			args = append(args, a)
		}

		ar := rolloutsv1alpha1.AnalysisRun{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cell.Namespace,
				Name:      fmt.Sprintf("%s-%s-%s", cell.Name, componentID, tmpl.TemplateName),
				Labels: map[string]string{
					LabelKeyStepIndex: componentID,
					LabelKeyCell:      cell.Name,
				},
			},
			Spec: rolloutsv1alpha1.AnalysisRunSpec{
				Args:    args,
				Metrics: at.Spec.Metrics,
			},
		}
		if err := ctrl.SetControllerReference(&cell, &ar, scheme); err != nil {
			log.Printf("Failed setting controller reference on %s/%s: %v", ar.Namespace, ar.Name, err)
		}

		if err := runtimeClient.Create(ctx, &ar); err != nil {
			return StepInProgress, err
		}

		log.Printf("Created analysisrun %s", ar.Name)

		return StepInProgress, nil
	case 1:
		ar := analysisRunList.Items[0]

		switch ar.Status.Phase {
		case rolloutsv1alpha1.AnalysisPhaseError, rolloutsv1alpha1.AnalysisPhaseFailed:
			log.Printf("AnalysisRun %s failed with error: %v", ar.Name, ar.Status.Message)

			return StepFailed, nil
		case rolloutsv1alpha1.AnalysisPhaseSuccessful:
		default:
			log.Printf("Waiting for analysisrun %s of %s to become %s", ar.Name, ar.Status.Phase, rolloutsv1alpha1.AnalysisPhaseSuccessful)

			// We need to wait for this analysis run to succeed
			return StepInProgress, nil
		}
	default:
		return StepInProgress, errors.New("too many analysis runs")
	}

	return StepPassed, nil
}
