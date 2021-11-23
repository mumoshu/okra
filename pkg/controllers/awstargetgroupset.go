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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/awstargetgroupset"
)

// AWSTargetGroupSetReconciler reconciles a AWSTargetGroupSet object
type AWSTargetGroupSetReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=okra.mumo.co,resources=awstargetgroupsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=okra.mumo.co,resources=awstargetgroupsets/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=okra.mumo.co,resources=awstargetgroupsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=awstargetgroup,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=awstargetgroup/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *AWSTargetGroupSetReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("awsTargetGroupSet", req.NamespacedName)

	var awsTargetGroupSet v1alpha1.AWSTargetGroupSet
	if err := r.Get(ctx, req.NamespacedName, &awsTargetGroupSet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if awsTargetGroupSet.ObjectMeta.DeletionTimestamp.IsZero() {
		finalizers, added := addFinalizer(awsTargetGroupSet.ObjectMeta.Finalizers)

		if added {
			newSet := awsTargetGroupSet.DeepCopy()
			newSet.ObjectMeta.Finalizers = finalizers

			if err := r.Update(ctx, newSet); err != nil {
				log.Error(err, "Failed to update AWSTargetGroupSet")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	} else {
		finalizers, removed := removeFinalizer(awsTargetGroupSet.ObjectMeta.Finalizers)

		if removed {
			// TODO do someo finalization if necessary

			newSet := awsTargetGroupSet.DeepCopy()
			newSet.ObjectMeta.Finalizers = finalizers

			if err := r.Update(ctx, newSet); err != nil {
				log.Error(err, "Failed to update AWSTargetGroupSet")
				return ctrl.Result{}, err
			}

			log.Info("Removed AWSTargetGroupSet")
		}

		return ctrl.Result{}, nil
	}

	awseks := awsTargetGroupSet.Spec.Generators[0].AWSEKS

	clusterSelector := labels.SelectorFromSet(awseks.ClusterSelector.MatchLabels)
	bindingSelector := labels.SelectorFromSet(awseks.BindingSelector.MatchLabels)

	config := awstargetgroupset.SyncInput{
		NS:              req.Namespace,
		ClusterSelector: clusterSelector.String(),
		BindingSelector: bindingSelector.String(),
		Labels:          awsTargetGroupSet.Spec.Template.Metadata.Labels,
	}

	results, err := awstargetgroupset.Sync(config)
	if err != nil {
		log.Error(err, "Syncing AWSTargetGroupSets")

		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	for _, r := range results {
		log.Info("%s AWSTargetGroup %s for cluster %s", r.Action, r.Name, r.Cluster)
	}

	r.Recorder.Event(&awsTargetGroupSet, corev1.EventTypeNormal, "SyncFinished", fmt.Sprintf("Sync finished on '%s'", awsTargetGroupSet.Name))

	return ctrl.Result{}, nil
}

func (r *AWSTargetGroupSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("awstargetgroupset-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AWSTargetGroupSet{}).
		Owns(&v1alpha1.AWSTargetGroup{}).
		Complete(r)
}
