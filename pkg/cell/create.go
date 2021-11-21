package cell

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
	Cell okrav1alpha1.Cell

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

func CreateOrUpdate(input ApplyInput) error {
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
		return err
	}

	cell := okrav1alpha1.Cell{}
	cell.SetNamespace(input.Cell.Namespace)
	cell.SetName(input.Cell.Name)

	op, err := ctrl.CreateOrUpdate(ctx, client, &cell, func() error {
		cell.Spec = input.Cell.Spec

		return nil
	})
	if err != nil {
		return err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("Reconciled Cell", "op", op)
	}

	return nil
}
