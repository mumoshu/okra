//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSApplicationLoadBalancerConfig) DeepCopyInto(out *AWSApplicationLoadBalancerConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSApplicationLoadBalancerConfig.
func (in *AWSApplicationLoadBalancerConfig) DeepCopy() *AWSApplicationLoadBalancerConfig {
	if in == nil {
		return nil
	}
	out := new(AWSApplicationLoadBalancerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSApplicationLoadBalancerConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSApplicationLoadBalancerConfigList) DeepCopyInto(out *AWSApplicationLoadBalancerConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSApplicationLoadBalancerConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSApplicationLoadBalancerConfigList.
func (in *AWSApplicationLoadBalancerConfigList) DeepCopy() *AWSApplicationLoadBalancerConfigList {
	if in == nil {
		return nil
	}
	out := new(AWSApplicationLoadBalancerConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSApplicationLoadBalancerConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSApplicationLoadBalancerConfigSpec) DeepCopyInto(out *AWSApplicationLoadBalancerConfigSpec) {
	*out = *in
	in.Listener.DeepCopyInto(&out.Listener)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSApplicationLoadBalancerConfigSpec.
func (in *AWSApplicationLoadBalancerConfigSpec) DeepCopy() *AWSApplicationLoadBalancerConfigSpec {
	if in == nil {
		return nil
	}
	out := new(AWSApplicationLoadBalancerConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSApplicationLoadBalancerConfigStatus) DeepCopyInto(out *AWSApplicationLoadBalancerConfigStatus) {
	*out = *in
	in.LastSyncTime.DeepCopyInto(&out.LastSyncTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSApplicationLoadBalancerConfigStatus.
func (in *AWSApplicationLoadBalancerConfigStatus) DeepCopy() *AWSApplicationLoadBalancerConfigStatus {
	if in == nil {
		return nil
	}
	out := new(AWSApplicationLoadBalancerConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSEKSClusterGenerator) DeepCopyInto(out *AWSEKSClusterGenerator) {
	*out = *in
	in.Selector.DeepCopyInto(&out.Selector)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSEKSClusterGenerator.
func (in *AWSEKSClusterGenerator) DeepCopy() *AWSEKSClusterGenerator {
	if in == nil {
		return nil
	}
	out := new(AWSEKSClusterGenerator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSEKSClusterSelector) DeepCopyInto(out *AWSEKSClusterSelector) {
	*out = *in
	if in.MatchTags != nil {
		in, out := &in.MatchTags, &out.MatchTags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSEKSClusterSelector.
func (in *AWSEKSClusterSelector) DeepCopy() *AWSEKSClusterSelector {
	if in == nil {
		return nil
	}
	out := new(AWSEKSClusterSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroup) DeepCopyInto(out *AWSTargetGroup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroup.
func (in *AWSTargetGroup) DeepCopy() *AWSTargetGroup {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSTargetGroup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupGenerator) DeepCopyInto(out *AWSTargetGroupGenerator) {
	*out = *in
	in.AWSEKS.DeepCopyInto(&out.AWSEKS)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupGenerator.
func (in *AWSTargetGroupGenerator) DeepCopy() *AWSTargetGroupGenerator {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupGenerator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupGeneratorAWSEKS) DeepCopyInto(out *AWSTargetGroupGeneratorAWSEKS) {
	*out = *in
	in.ClusterSelector.DeepCopyInto(&out.ClusterSelector)
	in.BindingSelector.DeepCopyInto(&out.BindingSelector)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupGeneratorAWSEKS.
func (in *AWSTargetGroupGeneratorAWSEKS) DeepCopy() *AWSTargetGroupGeneratorAWSEKS {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupGeneratorAWSEKS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupList) DeepCopyInto(out *AWSTargetGroupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSTargetGroup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupList.
func (in *AWSTargetGroupList) DeepCopy() *AWSTargetGroupList {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSTargetGroupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupSet) DeepCopyInto(out *AWSTargetGroupSet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupSet.
func (in *AWSTargetGroupSet) DeepCopy() *AWSTargetGroupSet {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSTargetGroupSet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupSetList) DeepCopyInto(out *AWSTargetGroupSetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSTargetGroupSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupSetList.
func (in *AWSTargetGroupSetList) DeepCopy() *AWSTargetGroupSetList {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupSetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSTargetGroupSetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupSetSpec) DeepCopyInto(out *AWSTargetGroupSetSpec) {
	*out = *in
	if in.Generators != nil {
		in, out := &in.Generators, &out.Generators
		*out = make([]AWSTargetGroupGenerator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Template.DeepCopyInto(&out.Template)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupSetSpec.
func (in *AWSTargetGroupSetSpec) DeepCopy() *AWSTargetGroupSetSpec {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupSetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupSetStatus) DeepCopyInto(out *AWSTargetGroupSetStatus) {
	*out = *in
	in.LastSyncTime.DeepCopyInto(&out.LastSyncTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupSetStatus.
func (in *AWSTargetGroupSetStatus) DeepCopy() *AWSTargetGroupSetStatus {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupSetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupSpec) DeepCopyInto(out *AWSTargetGroupSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupSpec.
func (in *AWSTargetGroupSpec) DeepCopy() *AWSTargetGroupSpec {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupStatus) DeepCopyInto(out *AWSTargetGroupStatus) {
	*out = *in
	in.Clusters.DeepCopyInto(&out.Clusters)
	in.LastSyncTime.DeepCopyInto(&out.LastSyncTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupStatus.
func (in *AWSTargetGroupStatus) DeepCopy() *AWSTargetGroupStatus {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupTemplate) DeepCopyInto(out *AWSTargetGroupTemplate) {
	*out = *in
	in.Metadata.DeepCopyInto(&out.Metadata)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupTemplate.
func (in *AWSTargetGroupTemplate) DeepCopy() *AWSTargetGroupTemplate {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSTargetGroupTemplateMetadata) DeepCopyInto(out *AWSTargetGroupTemplateMetadata) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSTargetGroupTemplateMetadata.
func (in *AWSTargetGroupTemplateMetadata) DeepCopy() *AWSTargetGroupTemplateMetadata {
	if in == nil {
		return nil
	}
	out := new(AWSTargetGroupTemplateMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cell) DeepCopyInto(out *Cell) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cell.
func (in *Cell) DeepCopy() *Cell {
	if in == nil {
		return nil
	}
	out := new(Cell)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Cell) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CellIngress) DeepCopyInto(out *CellIngress) {
	*out = *in
	if in.AWSApplicationLoadBalancer != nil {
		in, out := &in.AWSApplicationLoadBalancer, &out.AWSApplicationLoadBalancer
		*out = new(CellIngressAWSApplicationLoadBalancer)
		(*in).DeepCopyInto(*out)
	}
	if in.AWSNetworkLoadBalancer != nil {
		in, out := &in.AWSNetworkLoadBalancer, &out.AWSNetworkLoadBalancer
		*out = new(CellIngressAWSNetworkLoadBalancer)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CellIngress.
func (in *CellIngress) DeepCopy() *CellIngress {
	if in == nil {
		return nil
	}
	out := new(CellIngress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CellIngressAWSApplicationLoadBalancer) DeepCopyInto(out *CellIngressAWSApplicationLoadBalancer) {
	*out = *in
	in.Listener.DeepCopyInto(&out.Listener)
	in.TargetGroupSelector.DeepCopyInto(&out.TargetGroupSelector)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CellIngressAWSApplicationLoadBalancer.
func (in *CellIngressAWSApplicationLoadBalancer) DeepCopy() *CellIngressAWSApplicationLoadBalancer {
	if in == nil {
		return nil
	}
	out := new(CellIngressAWSApplicationLoadBalancer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CellIngressAWSNetworkLoadBalancer) DeepCopyInto(out *CellIngressAWSNetworkLoadBalancer) {
	*out = *in
	in.TargetGroupSelector.DeepCopyInto(&out.TargetGroupSelector)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CellIngressAWSNetworkLoadBalancer.
func (in *CellIngressAWSNetworkLoadBalancer) DeepCopy() *CellIngressAWSNetworkLoadBalancer {
	if in == nil {
		return nil
	}
	out := new(CellIngressAWSNetworkLoadBalancer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CellList) DeepCopyInto(out *CellList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Cell, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CellList.
func (in *CellList) DeepCopy() *CellList {
	if in == nil {
		return nil
	}
	out := new(CellList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CellList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CellSpec) DeepCopyInto(out *CellSpec) {
	*out = *in
	in.Ingress.DeepCopyInto(&out.Ingress)
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	in.UpdateStrategy.DeepCopyInto(&out.UpdateStrategy)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CellSpec.
func (in *CellSpec) DeepCopy() *CellSpec {
	if in == nil {
		return nil
	}
	out := new(CellSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CellStatus) DeepCopyInto(out *CellStatus) {
	*out = *in
	in.Clusters.DeepCopyInto(&out.Clusters)
	in.LastSyncTime.DeepCopyInto(&out.LastSyncTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CellStatus.
func (in *CellStatus) DeepCopy() *CellStatus {
	if in == nil {
		return nil
	}
	out := new(CellStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CellUpdateStrategy) DeepCopyInto(out *CellUpdateStrategy) {
	*out = *in
	if in.Canary != nil {
		in, out := &in.Canary, &out.Canary
		*out = new(CellUpdateStrategyCanary)
		(*in).DeepCopyInto(*out)
	}
	if in.BlueGreen != nil {
		in, out := &in.BlueGreen, &out.BlueGreen
		*out = new(CellUpdateStrategyBlueGreen)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CellUpdateStrategy.
func (in *CellUpdateStrategy) DeepCopy() *CellUpdateStrategy {
	if in == nil {
		return nil
	}
	out := new(CellUpdateStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CellUpdateStrategyBlueGreen) DeepCopyInto(out *CellUpdateStrategyBlueGreen) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CellUpdateStrategyBlueGreen.
func (in *CellUpdateStrategyBlueGreen) DeepCopy() *CellUpdateStrategyBlueGreen {
	if in == nil {
		return nil
	}
	out := new(CellUpdateStrategyBlueGreen)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CellUpdateStrategyCanary) DeepCopyInto(out *CellUpdateStrategyCanary) {
	*out = *in
	if in.Steps != nil {
		in, out := &in.Steps, &out.Steps
		*out = make([]rolloutsv1alpha1.CanaryStep, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Analysis != nil {
		in, out := &in.Analysis, &out.Analysis
		*out = new(rolloutsv1alpha1.RolloutAnalysisBackground)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CellUpdateStrategyCanary.
func (in *CellUpdateStrategyCanary) DeepCopy() *CellUpdateStrategyCanary {
	if in == nil {
		return nil
	}
	out := new(CellUpdateStrategyCanary)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterGenerator) DeepCopyInto(out *ClusterGenerator) {
	*out = *in
	in.AWSEKS.DeepCopyInto(&out.AWSEKS)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterGenerator.
func (in *ClusterGenerator) DeepCopy() *ClusterGenerator {
	if in == nil {
		return nil
	}
	out := new(ClusterGenerator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSecretTemplate) DeepCopyInto(out *ClusterSecretTemplate) {
	*out = *in
	in.Metadata.DeepCopyInto(&out.Metadata)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSecretTemplate.
func (in *ClusterSecretTemplate) DeepCopy() *ClusterSecretTemplate {
	if in == nil {
		return nil
	}
	out := new(ClusterSecretTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSecretTemplateMetadata) DeepCopyInto(out *ClusterSecretTemplateMetadata) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSecretTemplateMetadata.
func (in *ClusterSecretTemplateMetadata) DeepCopy() *ClusterSecretTemplateMetadata {
	if in == nil {
		return nil
	}
	out := new(ClusterSecretTemplateMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSet) DeepCopyInto(out *ClusterSet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSet.
func (in *ClusterSet) DeepCopy() *ClusterSet {
	if in == nil {
		return nil
	}
	out := new(ClusterSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterSet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSetList) DeepCopyInto(out *ClusterSetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSetList.
func (in *ClusterSetList) DeepCopy() *ClusterSetList {
	if in == nil {
		return nil
	}
	out := new(ClusterSetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterSetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSetSpec) DeepCopyInto(out *ClusterSetSpec) {
	*out = *in
	if in.Generators != nil {
		in, out := &in.Generators, &out.Generators
		*out = make([]ClusterGenerator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Template.DeepCopyInto(&out.Template)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSetSpec.
func (in *ClusterSetSpec) DeepCopy() *ClusterSetSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterSetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSetStatus) DeepCopyInto(out *ClusterSetStatus) {
	*out = *in
	in.Clusters.DeepCopyInto(&out.Clusters)
	in.LastSyncTime.DeepCopyInto(&out.LastSyncTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSetStatus.
func (in *ClusterSetStatus) DeepCopy() *ClusterSetStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterSetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSetStatusClusters) DeepCopyInto(out *ClusterSetStatusClusters) {
	*out = *in
	if in.Names != nil {
		in, out := &in.Names, &out.Names
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSetStatusClusters.
func (in *ClusterSetStatusClusters) DeepCopy() *ClusterSetStatusClusters {
	if in == nil {
		return nil
	}
	out := new(ClusterSetStatusClusters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Forward) DeepCopyInto(out *Forward) {
	*out = *in
	if in.TargetGroups != nil {
		in, out := &in.TargetGroups, &out.TargetGroups
		*out = make([]ForwardTargetGroup, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Forward.
func (in *Forward) DeepCopy() *Forward {
	if in == nil {
		return nil
	}
	out := new(Forward)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ForwardTargetGroup) DeepCopyInto(out *ForwardTargetGroup) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ForwardTargetGroup.
func (in *ForwardTargetGroup) DeepCopy() *ForwardTargetGroup {
	if in == nil {
		return nil
	}
	out := new(ForwardTargetGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Listener) DeepCopyInto(out *Listener) {
	*out = *in
	in.Rule.DeepCopyInto(&out.Rule)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Listener.
func (in *Listener) DeepCopy() *Listener {
	if in == nil {
		return nil
	}
	out := new(Listener)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ListenerRule) DeepCopyInto(out *ListenerRule) {
	*out = *in
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PathPatterns != nil {
		in, out := &in.PathPatterns, &out.PathPatterns
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Methods != nil {
		in, out := &in.Methods, &out.Methods
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SourceIPs != nil {
		in, out := &in.SourceIPs, &out.SourceIPs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make(map[string][]string, len(*in))
		for key, val := range *in {
			var outVal []string
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]string, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
	if in.QueryStrings != nil {
		in, out := &in.QueryStrings, &out.QueryStrings
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Forward.DeepCopyInto(&out.Forward)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ListenerRule.
func (in *ListenerRule) DeepCopy() *ListenerRule {
	if in == nil {
		return nil
	}
	out := new(ListenerRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pause) DeepCopyInto(out *Pause) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pause.
func (in *Pause) DeepCopy() *Pause {
	if in == nil {
		return nil
	}
	out := new(Pause)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Pause) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PauseList) DeepCopyInto(out *PauseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Pause, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PauseList.
func (in *PauseList) DeepCopy() *PauseList {
	if in == nil {
		return nil
	}
	out := new(PauseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PauseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PauseSpec) DeepCopyInto(out *PauseSpec) {
	*out = *in
	in.ExpireTime.DeepCopyInto(&out.ExpireTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PauseSpec.
func (in *PauseSpec) DeepCopy() *PauseSpec {
	if in == nil {
		return nil
	}
	out := new(PauseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PauseStatus) DeepCopyInto(out *PauseStatus) {
	*out = *in
	in.LastSyncTime.DeepCopyInto(&out.LastSyncTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PauseStatus.
func (in *PauseStatus) DeepCopy() *PauseStatus {
	if in == nil {
		return nil
	}
	out := new(PauseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TargetGroupBindingSelector) DeepCopyInto(out *TargetGroupBindingSelector) {
	*out = *in
	if in.MatchLabels != nil {
		in, out := &in.MatchLabels, &out.MatchLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TargetGroupBindingSelector.
func (in *TargetGroupBindingSelector) DeepCopy() *TargetGroupBindingSelector {
	if in == nil {
		return nil
	}
	out := new(TargetGroupBindingSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TargetGroupClusterSelector) DeepCopyInto(out *TargetGroupClusterSelector) {
	*out = *in
	if in.MatchLabels != nil {
		in, out := &in.MatchLabels, &out.MatchLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TargetGroupClusterSelector.
func (in *TargetGroupClusterSelector) DeepCopy() *TargetGroupClusterSelector {
	if in == nil {
		return nil
	}
	out := new(TargetGroupClusterSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TargetGroupSelector) DeepCopyInto(out *TargetGroupSelector) {
	*out = *in
	if in.MatchLabels != nil {
		in, out := &in.MatchLabels, &out.MatchLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.VersionLabels != nil {
		in, out := &in.VersionLabels, &out.VersionLabels
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TargetGroupSelector.
func (in *TargetGroupSelector) DeepCopy() *TargetGroupSelector {
	if in == nil {
		return nil
	}
	out := new(TargetGroupSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionBlocklist) DeepCopyInto(out *VersionBlocklist) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionBlocklist.
func (in *VersionBlocklist) DeepCopy() *VersionBlocklist {
	if in == nil {
		return nil
	}
	out := new(VersionBlocklist)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VersionBlocklist) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionBlocklistItem) DeepCopyInto(out *VersionBlocklistItem) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionBlocklistItem.
func (in *VersionBlocklistItem) DeepCopy() *VersionBlocklistItem {
	if in == nil {
		return nil
	}
	out := new(VersionBlocklistItem)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionBlocklistList) DeepCopyInto(out *VersionBlocklistList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VersionBlocklist, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionBlocklistList.
func (in *VersionBlocklistList) DeepCopy() *VersionBlocklistList {
	if in == nil {
		return nil
	}
	out := new(VersionBlocklistList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VersionBlocklistList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionBlocklistSpec) DeepCopyInto(out *VersionBlocklistSpec) {
	*out = *in
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VersionBlocklistItem, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionBlocklistSpec.
func (in *VersionBlocklistSpec) DeepCopy() *VersionBlocklistSpec {
	if in == nil {
		return nil
	}
	out := new(VersionBlocklistSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionBlocklistStatus) DeepCopyInto(out *VersionBlocklistStatus) {
	*out = *in
	in.LastSyncTime.DeepCopyInto(&out.LastSyncTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionBlocklistStatus.
func (in *VersionBlocklistStatus) DeepCopy() *VersionBlocklistStatus {
	if in == nil {
		return nil
	}
	out := new(VersionBlocklistStatus)
	in.DeepCopyInto(out)
	return out
}
