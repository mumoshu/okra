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

	"github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/awsapplicationloadbalancer"
)

// AWSApplicationLoadBalancerConfigReconciler reconciles a AWSApplicationLoadBalancerConfig object
type AWSApplicationLoadBalancerConfigReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=okra.mumo.co,resources=awsapplicationloadbalancereconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=okra.mumo.co,resources=awsapplicationloadbalancereconfigs/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=okra.mumo.co,resources=awsapplicationloadbalancereconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *AWSApplicationLoadBalancerConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("awsApplicationLoadBalancerConfig", req.NamespacedName)

	var awsALBConfig v1alpha1.AWSApplicationLoadBalancerConfig
	if err := r.Get(ctx, req.NamespacedName, &awsALBConfig); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if awsALBConfig.ObjectMeta.DeletionTimestamp.IsZero() {
		finalizers, added := addFinalizer(awsALBConfig.ObjectMeta.Finalizers)

		if added {
			updated := awsALBConfig.DeepCopy()
			updated.ObjectMeta.Finalizers = finalizers

			if err := r.Update(ctx, updated); err != nil {
				log.Error(err, "Failed to update AWSApplicationLoadBalancerConfig")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	} else {
		finalizers, removed := removeFinalizer(awsALBConfig.ObjectMeta.Finalizers)

		if removed {
			// TODO do some finalization if necessary

			updated := awsALBConfig.DeepCopy()
			updated.ObjectMeta.Finalizers = finalizers

			if err := r.Update(ctx, updated); err != nil {
				log.Error(err, "Failed to update AWSApplicationLoadBalancerConfig")
				return ctrl.Result{}, err
			}

			log.Info("Removed AWSApplicationLoadBalancerConfig")
		}

		return ctrl.Result{}, nil
	}

	config := awsapplicationloadbalancer.SyncInput{
		Spec: awsALBConfig.Spec,
	}

	err := awsapplicationloadbalancer.Sync(config)
	if err != nil {
		log.Error(err, "Syncing AWSApplicationLoadBalancerConfig")

		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	r.Recorder.Event(&awsALBConfig, corev1.EventTypeNormal, "SyncFinished", fmt.Sprintf("Sync finished on '%s'", awsALBConfig.Name))

	return ctrl.Result{}, nil
}

func (r *AWSApplicationLoadBalancerConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("awsapplicationloadbalancerconfig-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AWSApplicationLoadBalancerConfig{}).
		Complete(r)
}
