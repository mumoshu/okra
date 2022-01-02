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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
)

// CellSpec defines the desired state of ClusterSet
type CellSpec struct {
	Ingress  CellIngress `json:"ingress,omitempty"`
	Replicas *int32      `json:"replicas,omitempty"`
	// Version is the desired version number of target groups to be rolled out.
	// If the desired version is less than the current version, okra tries to swap the target groups registered in the loadbalancer
	// ASAP, so that a manual rollback can be done immediately.
	Version        string             `json:"version,omitempty"`
	UpdateStrategy CellUpdateStrategy `json:"updateStrategy,omitempty"`
}

type CellIngress struct {
	Type                       CellIngressType                        `json:"type,omitempty"`
	AWSApplicationLoadBalancer *CellIngressAWSApplicationLoadBalancer `json:"awsApplicationLoadBalancer,omitempty"`
	AWSNetworkLoadBalancer     *CellIngressAWSNetworkLoadBalancer     `json:"awsNetworkLoadBalancer,omitempty"`
}

type CellIngressType string

var ErrInvalidCellIngressType = fmt.Errorf("invalid cell ingress type")

func (v CellIngressType) String() string {
	return string(v)
}

func (v CellIngressType) Valid() error {
	switch v {
	case CellIngressTypeAWSApplicationLoadBalancer:
		return nil
	default:
		return errors.Wrapf(ErrInvalidCellIngressType, "get %s", v)
	}
}

func (v *CellIngressType) UnmarshalJSON(b []byte) error {
	*v = CellIngressType(strings.Trim(string(b), `"`))
	return v.Valid()
}

const (
	CellIngressTypeAWSApplicationLoadBalancer CellIngressType = "AWSApplicationLoadBalancer"
	CellIngressTypeAWSNetworkLoadBalancer     CellIngressType = "AWSNetworkLoadBalancer"
)

type CellIngressAWSApplicationLoadBalancer struct {
	ListenerARN         string              `json:"listenerARN,omitempty"`
	Listener            Listener            `json:"listener,omitempty"`
	TargetGroupSelector TargetGroupSelector `json:"targetGroupSelector,omitempty"`
}

type CellIngressAWSNetworkLoadBalancer struct {
	ListenerARN string `json:"listenerARN,omitempty"`

	TargetGroupSelector TargetGroupSelector `json:"targetGroupSelector,omitempty"`
}

type TargetGroupSelector struct {
	MatchLabels   map[string]string `json:"matchLabels,omitempty"`
	VersionLabels []string          `json:"versionLabels,omitempty"`
}

type CellUpdateStrategy struct {
	Type      CellUpdateStrategyType       `json:"type,omitempty"`
	Canary    *CellUpdateStrategyCanary    `json:"canary,omitempty"`
	BlueGreen *CellUpdateStrategyBlueGreen `json:"blueGreen,omitempty"`
}

type CellUpdateStrategyType string

const (
	CellUpdateStrategyTypeCanary    CellUpdateStrategyType = "Canary"
	CellUpdateStrategyTypeBlueGreen CellUpdateStrategyType = "BlueGreen"
)

type CellUpdateStrategyCanary struct {
	// Steps define the order of phases to execute the canary deployment
	// +optional
	Steps []rolloutsv1alpha1.CanaryStep `json:"steps,omitempty" protobuf:"bytes,3,rep,name=steps"`
	// Analysis runs a separate analysisRun while all the steps execute. This is intended to be a continuous validation of the new set of clusters
	Analysis *rolloutsv1alpha1.RolloutAnalysisBackground `json:"analysis,omitempty" protobuf:"bytes,7,opt,name=analysis"`
}

type CellUpdateStrategyBlueGreen struct {
}

// CellStatus defines the observed state of ClusterSet
type CellStatus struct {
	Clusters     ClusterSetStatusClusters `json:"clusters"`
	LastSyncTime metav1.Time              `json:"lastSyncTime"`
	Phase        string                   `json:"phase"`
	Reason       string                   `json:"reason"`
	Message      string                   `json:"message"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastSyncTime",name=Last Sync,type=date

// ClusterSet is the Schema for the ClusterSet API
type Cell struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CellSpec   `json:"spec,omitempty"`
	Status CellStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CellList contains a list of Cell
type CellList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cell `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cell{}, &CellList{})
}
