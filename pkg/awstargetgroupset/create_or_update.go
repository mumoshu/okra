package awstargetgroupset

import (
	"context"

	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/clclient"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ApplyInput struct {
	NS              string
	Name            string
	ClusterSelector map[string]string
	BindingSelector map[string]string
	Labels          map[string]string

	Client  client.Client
	Scheme  *runtime.Scheme
	Context context.Context
	Log     interface {
		Info(msg string, keysAndValues ...interface{})
	}
}

var noopLogger logger

type logger struct {
}

func (_ logger) Info(msg string, keysAndValues ...interface{}) {

}

func CreateOrUpdate(input ApplyInput) (*okrav1alpha1.AWSTargetGroupSet, error) {
	ctx := input.Context
	if ctx == nil {
		ctx = context.TODO()
	}

	logger := input.Log
	if logger == nil {
		logger = noopLogger
	}

	client, _, err := clclient.Init(input.Client, input.Scheme)
	if err != nil {
		return nil, err
	}

	set := okrav1alpha1.AWSTargetGroupSet{}
	set.SetNamespace(input.NS)
	set.SetName(input.Name)
	op, err := ctrl.CreateOrUpdate(ctx, client, &set, func() error {
		set.Spec.Generators = []okrav1alpha1.AWSTargetGroupGenerator{
			{
				AWSEKS: okrav1alpha1.AWSTargetGroupGeneratorAWSEKS{
					ClusterSelector: okrav1alpha1.TargetGroupClusterSelector{
						MatchLabels: input.ClusterSelector,
					},
					BindingSelector: okrav1alpha1.TargetGroupBindingSelector{
						MatchLabels: input.BindingSelector,
					},
				},
			},
		}
		set.Spec.Template = okrav1alpha1.AWSTargetGroupTemplate{
			Metadata: okrav1alpha1.AWSTargetGroupTemplateMetadata{
				Labels: input.Labels,
			},
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("Applied AWSTargetGroupSet", "op", op)
	}

	return &set, nil
}
