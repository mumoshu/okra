package targetgroupbinding

import (
	"context"

	"github.com/mumoshu/okra/api/elbv2/v1beta1"
	"github.com/mumoshu/okra/pkg/clclient"
	"github.com/mumoshu/okra/pkg/okraerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type CreateInput struct {
	ClusterName      string
	ClusterNamespace string
	Name             string
	Namespace        string
	TargetGroupARN   string
	Labels           map[string]string
	DryRun           bool
}

func Create(input CreateInput) (*v1beta1.TargetGroupBinding, error) {
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

	binding.Name = input.Name
	binding.Namespace = input.Namespace
	binding.Labels = input.Labels
	binding.Spec.TargetGroupARN = input.TargetGroupARN

	var dryRun []string

	if input.DryRun {
		dryRun = []string{metav1.DryRunAll}
	}

	if err := client.Create(ctx, &binding, &runtimeclient.CreateOptions{DryRun: dryRun}); err != nil {
		return nil, okraerror.New(err)
	}

	return &binding, nil
}
