/*
Copyright 2020 The Okra authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	//"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/cell"
)

// CellReconciler reconciles a Cell object
type CellReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=okra.mumo.co,resources=cells,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=okra.mumo.co,resources=cells/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=okra.mumo.co,resources=cells/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=okra.mumo.co,resources=versionblocklists,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *CellReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cell", req.NamespacedName)

	var cellResource okrav1alpha1.Cell
	if err := r.Get(ctx, req.NamespacedName, &cellResource); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if cellResource.ObjectMeta.DeletionTimestamp.IsZero() {
		finalizers, added := addFinalizer(cellResource.ObjectMeta.Finalizers)

		if added {
			updated := cellResource.DeepCopy()
			updated.ObjectMeta.Finalizers = finalizers

			if err := r.Update(ctx, updated); err != nil {
				log.Error(err, "Failed to update Cell")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	} else {
		finalizers, removed := removeFinalizer(cellResource.ObjectMeta.Finalizers)

		if removed {
			// TODO do some finalization if necessary

			updated := cellResource.DeepCopy()
			updated.ObjectMeta.Finalizers = finalizers

			if err := r.Update(ctx, updated); err != nil {
				log.Error(err, "Failed to update Cell")
				return ctrl.Result{}, err
			}

			log.Info("Removed Cell")
		}

		return ctrl.Result{}, nil
	}

	err := cell.Sync(cell.SyncInput{
		Cell:   &cellResource,
		Client: r.Client,
		Scheme: r.Scheme,
	})
	if err != nil {
		log.Error(err, "Syncing Cell")

		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	r.Recorder.Event(&cellResource, corev1.EventTypeNormal, "SyncFinished", fmt.Sprintf("Sync finished on '%s'", cellResource.Name))

	return ctrl.Result{}, nil
}

func (r *CellReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("cell-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&okrav1alpha1.Cell{}).
		Owns(&okrav1alpha1.AWSApplicationLoadBalancerConfig{}).
		Owns(&okrav1alpha1.Pause{}).
		Owns(&rolloutsv1alpha1.AnalysisRun{}).
		Owns(&rolloutsv1alpha1.Experiment{}).
		Complete(r)
}
