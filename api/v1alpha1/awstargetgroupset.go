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

// AWSTargetGroupSpec defines the desired state of AWSTargetGroupp
type AWSTargetGroupSetSpec struct {
	ARN        string                    `json:"arn,omitempty"`
	Generators []AWSTargetGroupGenerator `json:"generators,omitempty"`
	Template   AWSTargetGroupTemplate    `json:"template,omitempty"`
}

type AWSTargetGroupGenerator struct {
	AWSEKS AWSTargetGroupGeneratorAWSEKS `json:"awseks,omitempty"`
}

type AWSTargetGroupGeneratorAWSEKS struct {
	ClusterSelector TargetGroupClusterSelector `json:"clusterSelector,omitempty"`
	BindingSelector TargetGroupBindingSelector `json:"bindingSelector,omitempty"`
}

type TargetGroupClusterSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type TargetGroupBindingSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type AWSTargetGroupTemplate struct {
	Metadata AWSTargetGroupTemplateMetadata `json:"metadata,omitempty"`
}

type AWSTargetGroupTemplateMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

// AWSTargetGroupSetStatus defines the observed state of AWSTargetGroupSet
type AWSTargetGroupSetStatus struct {
	LastSyncTime metav1.Time `json:"lastSyncTime"`
	Phase        string      `json:"phase"`
	Reason       string      `json:"reason"`
	Message      string      `json:"message"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastSyncTime",name=Last Sync,type=date

// AWSTargetGroupSet is the Schema for the AWSTargetGroupSet API
type AWSTargetGroupSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSTargetGroupSetSpec   `json:"spec,omitempty"`
	Status AWSTargetGroupSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CellList contains a list of Cell
type AWSTargetGroupSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSTargetGroupSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSTargetGroupSet{}, &AWSTargetGroupSetList{})
}
