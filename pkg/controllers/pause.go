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

	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/pause"
)

// PauseReconciler reconciles a Pause object
type PauseReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=okra.mumo.co,resources=pauses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=okra.mumo.co,resources=pauses/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=okra.mumo.co,resources=pauses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *PauseReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pause", req.NamespacedName)

	var pauseResource okrav1alpha1.Pause
	if err := r.Get(ctx, req.NamespacedName, &pauseResource); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if pauseResource.ObjectMeta.DeletionTimestamp.IsZero() {
		finalizers, added := addFinalizer(pauseResource.ObjectMeta.Finalizers)

		if added {
			updated := pauseResource.DeepCopy()
			updated.ObjectMeta.Finalizers = finalizers

			if err := r.Update(ctx, updated); err != nil {
				log.Error(err, "Failed to update Pause")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	} else {
		finalizers, removed := removeFinalizer(pauseResource.ObjectMeta.Finalizers)

		if removed {
			// TODO do some finalization if necessary

			updated := pauseResource.DeepCopy()
			updated.ObjectMeta.Finalizers = finalizers

			if err := r.Update(ctx, updated); err != nil {
				log.Error(err, "Failed to update Pause")
				return ctrl.Result{}, err
			}

			log.Info("Removed Pause")
		}

		return ctrl.Result{}, nil
	}

	t := time.Now()
	config := pause.SyncInput{
		Pause:  pauseResource,
		Now:    &t,
		Client: r.Client,
	}

	err := pause.Sync(config)
	if err != nil {
		log.Error(err, "Syncing Pause")

		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	r.Recorder.Event(&pauseResource, corev1.EventTypeNormal, "SyncFinished", fmt.Sprintf("Sync finished on '%s'", pauseResource.Name))

	return ctrl.Result{}, nil
}

func (r *PauseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("pause-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&okrav1alpha1.Pause{}).
		Complete(r)
}
