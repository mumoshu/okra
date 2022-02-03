package okratest

import (
	"context"
	"io"

	"github.com/mumoshu/okra/pkg/clclient"
	"github.com/mumoshu/okra/pkg/okraerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExampleApp struct {
	// Image is the "repo:tag" of the contaimer image
	Image string

	ClientSet kubernetes.Interface
	Client    client.Client

	ClusterName, Namespace, ServiceName string
}

func NewExampleApp(ctx context.Context, clusterName, ns, serviceName string) (*ExampleApp, error) {
	clientset, err := clclient.NewClientSet()
	if err != nil {
		return nil, okraerror.New(err)
	}

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

	image := "mumoshu/okra-exampleapp:latest"

	return &ExampleApp{
		Image:       image,
		ClientSet:   clientset,
		Client:      client,
		ClusterName: clusterName,
		Namespace:   ns,
		ServiceName: serviceName,
	}, nil
}

func (a *ExampleApp) Start(ctx context.Context) error {
	for _, o := range a.ResourceObjects() {
		if err := a.Client.Create(ctx, o, client.FieldOwner("okratest")); err != nil {
			return err
		}
	}
	return nil
}

func (a *ExampleApp) ResourceObjects() []runtime.Object {
	return nil
}

func (a *ExampleApp) WriteManifests(out io.Writer) error {
	return nil
}

func (a *ExampleApp) Get(path string) error {
	return nil
}

func (a *ExampleApp) Wait() error {
	return nil
}

func (a *ExampleApp) Stop(ctx context.Context) error {
	for _, o := range a.ResourceObjects() {
		if err := a.Client.Delete(ctx, o); err != nil {
			return err
		}
	}
	return nil
}
