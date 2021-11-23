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

// VersionBlocklistSpec defines the desired state of VersionBlocklist
type VersionBlocklistSpec struct {
	Items []VersionBlocklistItem `json:"items"`
}

type VersionBlocklistItem struct {
	Version string `json:"version"`
	Cause   string `json:"cause"`
}

// VersionBlocklistStatus defines the observed state of VersionBlocklist
type VersionBlocklistStatus struct {
	LastSyncTime metav1.Time `json:"lastSyncTime"`
	Phase        PausePhase  `json:"phase"`
	Reason       string      `json:"reason"`
	Message      string      `json:"message"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".stats.phase",name=Phase,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastSyncTime",name=Last Sync,type=date

// Pause represents a pause in a cell canary update
type VersionBlocklist struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VersionBlocklistSpec   `json:"spec,omitempty"`
	Status VersionBlocklistStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PauseList contains a list of Pause
type VersionBlocklistList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VersionBlocklist `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VersionBlocklist{}, &VersionBlocklistList{})
}
