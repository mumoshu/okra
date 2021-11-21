package analysis

import (
	"context"

	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	"github.com/mumoshu/okra/pkg/clclient"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UpdateInput struct {
	NS, Name string
	Phase    rolloutsv1alpha1.AnalysisPhase

	Client client.Client
}

func Update(input UpdateInput) error {
	ns, name := input.NS, input.Name

	ctx := context.TODO()

	managementClient := input.Client

	if managementClient == nil {
		var err error

		managementClient, err = clclient.New()
		if err != nil {
			return err
		}
	}

	var run rolloutsv1alpha1.AnalysisRun

	if err := managementClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &run); err != nil {
		return err
	}

	patch := &rolloutsv1alpha1.AnalysisRun{
		TypeMeta: v1.TypeMeta{
			APIVersion: rolloutsv1alpha1.GroupVersion.String(),
			Kind:       "AnalysisRun",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: run.Spec,
		Status: rolloutsv1alpha1.AnalysisRunStatus{
			Phase: input.Phase,
		},
	}
	if err := managementClient.Patch(ctx, patch, client.Apply, client.ForceOwnership, client.FieldOwner("okra")); err != nil {
		return err
	}

	return nil
}
