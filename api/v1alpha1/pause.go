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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PausePhase are a set of phases of pause
type PausePhase string

const (
	// RolloutPhaseHealthy indicates a pause is started
	PausePhaseStarted PausePhase = "Started"
	// PausePhaseExpired indicates a pause is expired
	PausePhaseExpired PausePhase = "Expired"
	// PausePhaseExpired indicates a pause is cancelled via a human operation or any other controller operation
	PausePhaseCancelled PausePhase = "Cancelled"
)

// PauseSpec defines the desired state of Pause
type PauseSpec struct {
	ExpireTime metav1.Time `json:"expireTime,omitempty"`
}

// PauseStatus defines the observed state of Pause
type PauseStatus struct {
	LastSyncTime metav1.Time `json:"lastSyncTime"`
	Phase        PausePhase  `json:"phase"`
	Reason       string      `json:"reason"`
	Message      string      `json:"message"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".stats.phase",name=Phase,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.expireTime",name=Expires At,type=date
// +kubebuilder:printcolumn:JSONPath=".status.lastSyncTime",name=Last Sync,type=date

// Pause represents a pause in a cell canary update
type Pause struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PauseSpec   `json:"spec,omitempty"`
	Status PauseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PauseList contains a list of Pause
type PauseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pause `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pause{}, &PauseList{})
}
