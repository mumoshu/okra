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

// AWSApplicationLoadBalancerConfigSpec defines the desired state of AWSApplicationLoadBalancerConfigp
type AWSApplicationLoadBalancerConfigSpec struct {
	ListenerARN string  `json:"listenerARN,omitempty"`
	Forward     Forward `json:"forward,omitempty"`
}

type Forward struct {
	TargetGroups []ForwardTargetGroup `json:"targetGroups,omitempty"`
}

type ForwardTargetGroup struct {
	Name   string `json:"name,omitempty"`
	ARN    string `json:"arn,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

// AWSApplicationLoadBalancerConfigStatus defines the observed state of AWSApplicationLoadBalancerConfig
type AWSApplicationLoadBalancerConfigStatus struct {
	LastSyncTime metav1.Time `json:"lastSyncTime"`
	Phase        string      `json:"phase"`
	Reason       string      `json:"reason"`
	Message      string      `json:"message"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastSyncTime",name=Last Sync,type=date

// AWSApplicationLoadBalancerConfig is the Schema for the AWSApplicationLoadBalancerConfig API
type AWSApplicationLoadBalancerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSApplicationLoadBalancerConfigSpec   `json:"spec,omitempty"`
	Status AWSApplicationLoadBalancerConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CellList contains a list of Cell
type AWSApplicationLoadBalancerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSApplicationLoadBalancerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSApplicationLoadBalancerConfig{}, &AWSApplicationLoadBalancerConfigList{})
}
