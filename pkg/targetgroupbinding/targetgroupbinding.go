package targetgroupbinding

import (
	"context"
	"fmt"

	"github.com/mumoshu/okra/api/elbv2/v1beta1"
	"github.com/mumoshu/okra/pkg/clclient"
	"github.com/mumoshu/okra/pkg/okraerror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ListInput struct {
	ClusterName string
	NS          string
}

func List(input ListInput) ([]v1beta1.TargetGroupBinding, error) {
	clusterName := input.ClusterName
	ns := input.NS

	clientset, err := clclient.NewClientSet()
	if err != nil {
		return nil, okraerror.New(err)
	}

	ctx := context.Background()

	secret, err := clientset.CoreV1().Secrets(ns).Get(ctx, clusterName, metav1.GetOptions{})
	if err != nil {
		return nil, okraerror.New(err)
	}

	// for k, v := range secret.Data {
	// 	fmt.Fprintf(os.Stderr, "%s=%s\n", k, v)
	// }

	client, err := clclient.NewFromClusterSecret(*secret)
	if err != nil {
		return nil, err
	}

	var bindings v1beta1.TargetGroupBindingList

	optionalNS := ""

	if err := client.List(ctx, &bindings, &runtimeclient.ListOptions{Namespace: optionalNS}); err != nil {
		return nil, okraerror.New(err)
	}

	return bindings.Items, nil
}

type ApplyInput struct {
	ClusterName      string
	ClusterNamespace string
	Name             string
	Namespace        string
	TargetGroupARN   string
	Labels           map[string]string
	DryRun           bool
}

func Apply(input ApplyInput) (*v1beta1.TargetGroupBinding, error) {
	clientset, err := clclient.NewClientSet()
	if err != nil {
		return nil, okraerror.New(err)
	}

	ctx := context.Background()

	secret, err := clientset.CoreV1().Secrets(input.ClusterNamespace).Get(ctx, input.ClusterName, metav1.GetOptions{})
	if err != nil {
		return nil, okraerror.New(err)
	}

	client, err := clclient.NewFromClusterSecret(*secret)
	if err != nil {
		return nil, err
	}

	var binding v1beta1.TargetGroupBinding
	var bindingExists bool

	if err := client.Get(ctx, types.NamespacedName{Namespace: input.Namespace, Name: input.Name}, &binding); err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		binding.Name = input.Name
		binding.Namespace = input.Namespace
	} else {
		bindingExists = true
	}
	binding.Labels = input.Labels
	binding.Spec.TargetGroupARN = input.TargetGroupARN

	var dryRun []string

	if input.DryRun {
		dryRun = []string{metav1.DryRunAll}
	}
	if bindingExists {
		if err := client.Update(ctx, &binding); err != nil {
			return nil, okraerror.New(fmt.Errorf("updating binding: %w", err))
		}
	} else {
		if err := client.Create(ctx, &binding, &runtimeclient.CreateOptions{DryRun: dryRun}); err != nil {
			return nil, okraerror.New(fmt.Errorf("creating binding: %w", err))
		}
	}

	return &binding, nil
}
