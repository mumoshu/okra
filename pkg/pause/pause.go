package pause

import (
	"context"
	"fmt"
	"log"
	"time"

	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/clclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SyncInput struct {
	Pause okrav1alpha1.Pause

	Now *time.Time

	Client client.Client
}

func Sync(config SyncInput) error {
	now := config.Now
	if now == nil {
		t := time.Now()
		now = &t
	}

	ctx := context.TODO()

	managementClient := config.Client

	if managementClient == nil {
		var err error

		managementClient, err = clclient.New()
		if err != nil {
			return err
		}
	}

	pause := config.Pause
	ns := pause.Namespace
	name := pause.Name

	var current okrav1alpha1.Pause

	if err := managementClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &current); err != nil {
		return err
	}

	expireAt := current.Spec.ExpireTime

	log.Printf("expire at %s, current time is %s", expireAt, *now)

	if expireAt.Time.Before(*now) {
		status := current.Status.DeepCopy()
		status.Phase = okrav1alpha1.PausePhaseExpired
		status.LastSyncTime = metav1.Now()

		updated := &okrav1alpha1.Pause{
			TypeMeta: metav1.TypeMeta{
				APIVersion: okrav1alpha1.GroupVersion.String(),
				Kind:       "Pause",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
			Spec:   current.Spec,
			Status: *status,
		}

		if err := managementClient.Status().Patch(ctx, updated, client.Apply, client.ForceOwnership, client.FieldOwner("okra")); err != nil {
			return err
		}
		log.Printf("Updated pause %s to phase %s", name, updated.Status.Phase)
	} else if current.Status.Phase == "" {
		status := current.Status.DeepCopy()
		status.Phase = okrav1alpha1.PausePhaseStarted
		status.LastSyncTime = metav1.Now()

		updated := &okrav1alpha1.Pause{
			TypeMeta: metav1.TypeMeta{
				APIVersion: okrav1alpha1.GroupVersion.String(),
				Kind:       "Pause",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
			Spec:   current.Spec,
			Status: *status,
		}

		if err := managementClient.Status().Patch(ctx, updated, client.Apply, client.ForceOwnership, client.FieldOwner("okra")); err != nil {
			return err
		}

		log.Printf("Started pause %s", name)
	}

	return nil
}

type CancelInput struct {
	Pause okrav1alpha1.Pause

	Client client.Client
}

func Cancel(config CancelInput) error {
	ctx := context.TODO()

	managementClient := config.Client

	if managementClient == nil {
		var err error

		managementClient, err = clclient.New()
		if err != nil {
			return err
		}
	}

	pause := config.Pause
	phase := pause.Status.Phase
	ns, name := pause.Namespace, pause.Name

	var current okrav1alpha1.Pause

	if err := managementClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &current); err != nil {
		return err
	}

	switch current.Status.Phase {
	case okrav1alpha1.PausePhaseCancelled:
		return fmt.Errorf("cannot cancel already cancelled pause %s", pause.Name)
	case okrav1alpha1.PausePhaseExpired:
		return fmt.Errorf("cannot cancel already expired pause %s", pause.Name)
	case okrav1alpha1.PausePhaseStarted:
		status := current.Status.DeepCopy()
		status.Phase = okrav1alpha1.PausePhaseCancelled
		status.LastSyncTime = metav1.Now()

		updated := &okrav1alpha1.Pause{
			TypeMeta: metav1.TypeMeta{
				APIVersion: okrav1alpha1.GroupVersion.String(),
				Kind:       "Pause",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
			Spec:   current.Spec,
			Status: *status,
		}
		if err := managementClient.Status().Patch(ctx, updated, client.Apply, client.ForceOwnership, client.FieldOwner("okra")); err != nil {
			return err
		}

		log.Printf("canceled started pause %s", name)
	case "":
		status := current.Status.DeepCopy()
		status.Phase = okrav1alpha1.PausePhaseCancelled
		status.LastSyncTime = metav1.Now()

		updated := &okrav1alpha1.Pause{
			TypeMeta: metav1.TypeMeta{
				APIVersion: okrav1alpha1.GroupVersion.String(),
				Kind:       "Pause",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
			Spec:   current.Spec,
			Status: *status,
		}
		if err := managementClient.Status().Patch(ctx, updated, client.Apply, client.ForceOwnership, client.FieldOwner("okra")); err != nil {
			return err
		}

		log.Printf("canceled unstarted pause %s", name)
	default:
		return fmt.Errorf("unexpected pause phase: %s", phase)
	}

	return nil
}
