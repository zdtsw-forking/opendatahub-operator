//go:build !ignore_autogenerated

/*
Copyright 2023.

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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodeFlare) DeepCopyInto(out *CodeFlare) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodeFlare.
func (in *CodeFlare) DeepCopy() *CodeFlare {
	if in == nil {
		return nil
	}
	out := new(CodeFlare)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CodeFlare) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodeFlareList) DeepCopyInto(out *CodeFlareList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CodeFlare, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodeFlareList.
func (in *CodeFlareList) DeepCopy() *CodeFlareList {
	if in == nil {
		return nil
	}
	out := new(CodeFlareList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CodeFlareList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodeFlareSpec) DeepCopyInto(out *CodeFlareSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodeFlareSpec.
func (in *CodeFlareSpec) DeepCopy() *CodeFlareSpec {
	if in == nil {
		return nil
	}
	out := new(CodeFlareSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodeFlareStatus) DeepCopyInto(out *CodeFlareStatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodeFlareStatus.
func (in *CodeFlareStatus) DeepCopy() *CodeFlareStatus {
	if in == nil {
		return nil
	}
	out := new(CodeFlareStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComponentSpec) DeepCopyInto(out *ComponentSpec) {
	*out = *in
	in.DSCComponentSpec.DeepCopyInto(&out.DSCComponentSpec)
	in.DSCISpec.DeepCopyInto(&out.DSCISpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComponentSpec.
func (in *ComponentSpec) DeepCopy() *ComponentSpec {
	if in == nil {
		return nil
	}
	out := new(ComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComponentStatus) DeepCopyInto(out *ComponentStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComponentStatus.
func (in *ComponentStatus) DeepCopy() *ComponentStatus {
	if in == nil {
		return nil
	}
	out := new(ComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DSCComponentSpec) DeepCopyInto(out *DSCComponentSpec) {
	*out = *in
	in.DSCDevFlags.DeepCopyInto(&out.DSCDevFlags)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DSCComponentSpec.
func (in *DSCComponentSpec) DeepCopy() *DSCComponentSpec {
	if in == nil {
		return nil
	}
	out := new(DSCComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DSCDevFlags) DeepCopyInto(out *DSCDevFlags) {
	*out = *in
	if in.Manifests != nil {
		in, out := &in.Manifests, &out.Manifests
		*out = make([]ManifestsConfig, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DSCDevFlags.
func (in *DSCDevFlags) DeepCopy() *DSCDevFlags {
	if in == nil {
		return nil
	}
	out := new(DSCDevFlags)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Dashboard) DeepCopyInto(out *Dashboard) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Dashboard.
func (in *Dashboard) DeepCopy() *Dashboard {
	if in == nil {
		return nil
	}
	out := new(Dashboard)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Dashboard) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DashboardComponentSpec) DeepCopyInto(out *DashboardComponentSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DashboardComponentSpec.
func (in *DashboardComponentSpec) DeepCopy() *DashboardComponentSpec {
	if in == nil {
		return nil
	}
	out := new(DashboardComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DashboardComponentStatus) DeepCopyInto(out *DashboardComponentStatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DashboardComponentStatus.
func (in *DashboardComponentStatus) DeepCopy() *DashboardComponentStatus {
	if in == nil {
		return nil
	}
	out := new(DashboardComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DashboardList) DeepCopyInto(out *DashboardList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Dashboard, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DashboardList.
func (in *DashboardList) DeepCopy() *DashboardList {
	if in == nil {
		return nil
	}
	out := new(DashboardList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DashboardList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataSciencePipelineSpec) DeepCopyInto(out *DataSciencePipelineSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataSciencePipelineSpec.
func (in *DataSciencePipelineSpec) DeepCopy() *DataSciencePipelineSpec {
	if in == nil {
		return nil
	}
	out := new(DataSciencePipelineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataSciencePipelines) DeepCopyInto(out *DataSciencePipelines) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataSciencePipelines.
func (in *DataSciencePipelines) DeepCopy() *DataSciencePipelines {
	if in == nil {
		return nil
	}
	out := new(DataSciencePipelines)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DataSciencePipelines) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataSciencePipelinesList) DeepCopyInto(out *DataSciencePipelinesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DataSciencePipelines, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataSciencePipelinesList.
func (in *DataSciencePipelinesList) DeepCopy() *DataSciencePipelinesList {
	if in == nil {
		return nil
	}
	out := new(DataSciencePipelinesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataSciencePipelinetatus) DeepCopyInto(out *DataSciencePipelinetatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataSciencePipelinetatus.
func (in *DataSciencePipelinetatus) DeepCopy() *DataSciencePipelinetatus {
	if in == nil {
		return nil
	}
	out := new(DataSciencePipelinetatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KFTOComponentSpec) DeepCopyInto(out *KFTOComponentSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KFTOComponentSpec.
func (in *KFTOComponentSpec) DeepCopy() *KFTOComponentSpec {
	if in == nil {
		return nil
	}
	out := new(KFTOComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KFTOComponentStatus) DeepCopyInto(out *KFTOComponentStatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KFTOComponentStatus.
func (in *KFTOComponentStatus) DeepCopy() *KFTOComponentStatus {
	if in == nil {
		return nil
	}
	out := new(KFTOComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Kserve) DeepCopyInto(out *Kserve) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Kserve.
func (in *Kserve) DeepCopy() *Kserve {
	if in == nil {
		return nil
	}
	out := new(Kserve)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Kserve) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KserveComponentSpec) DeepCopyInto(out *KserveComponentSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
	out.Kserve = in.Kserve
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KserveComponentSpec.
func (in *KserveComponentSpec) DeepCopy() *KserveComponentSpec {
	if in == nil {
		return nil
	}
	out := new(KserveComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KserveComponentStatus) DeepCopyInto(out *KserveComponentStatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KserveComponentStatus.
func (in *KserveComponentStatus) DeepCopy() *KserveComponentStatus {
	if in == nil {
		return nil
	}
	out := new(KserveComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KserveCustomizedSpec) DeepCopyInto(out *KserveCustomizedSpec) {
	*out = *in
	out.Serving = in.Serving
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KserveCustomizedSpec.
func (in *KserveCustomizedSpec) DeepCopy() *KserveCustomizedSpec {
	if in == nil {
		return nil
	}
	out := new(KserveCustomizedSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KserveList) DeepCopyInto(out *KserveList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Kserve, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KserveList.
func (in *KserveList) DeepCopy() *KserveList {
	if in == nil {
		return nil
	}
	out := new(KserveList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KserveList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Kueue) DeepCopyInto(out *Kueue) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Kueue.
func (in *Kueue) DeepCopy() *Kueue {
	if in == nil {
		return nil
	}
	out := new(Kueue)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Kueue) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KueueComponentSpec) DeepCopyInto(out *KueueComponentSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KueueComponentSpec.
func (in *KueueComponentSpec) DeepCopy() *KueueComponentSpec {
	if in == nil {
		return nil
	}
	out := new(KueueComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KueueComponentStatus) DeepCopyInto(out *KueueComponentStatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KueueComponentStatus.
func (in *KueueComponentStatus) DeepCopy() *KueueComponentStatus {
	if in == nil {
		return nil
	}
	out := new(KueueComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KueueList) DeepCopyInto(out *KueueList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Kueue, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KueueList.
func (in *KueueList) DeepCopy() *KueueList {
	if in == nil {
		return nil
	}
	out := new(KueueList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KueueList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManifestsConfig) DeepCopyInto(out *ManifestsConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManifestsConfig.
func (in *ManifestsConfig) DeepCopy() *ManifestsConfig {
	if in == nil {
		return nil
	}
	out := new(ManifestsConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelMeshServing) DeepCopyInto(out *ModelMeshServing) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelMeshServing.
func (in *ModelMeshServing) DeepCopy() *ModelMeshServing {
	if in == nil {
		return nil
	}
	out := new(ModelMeshServing)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ModelMeshServing) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelMeshServingComponentSpec) DeepCopyInto(out *ModelMeshServingComponentSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelMeshServingComponentSpec.
func (in *ModelMeshServingComponentSpec) DeepCopy() *ModelMeshServingComponentSpec {
	if in == nil {
		return nil
	}
	out := new(ModelMeshServingComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelMeshServingComponentStatus) DeepCopyInto(out *ModelMeshServingComponentStatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelMeshServingComponentStatus.
func (in *ModelMeshServingComponentStatus) DeepCopy() *ModelMeshServingComponentStatus {
	if in == nil {
		return nil
	}
	out := new(ModelMeshServingComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelMeshServingList) DeepCopyInto(out *ModelMeshServingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ModelMeshServing, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelMeshServingList.
func (in *ModelMeshServingList) DeepCopy() *ModelMeshServingList {
	if in == nil {
		return nil
	}
	out := new(ModelMeshServingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ModelMeshServingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelReg) DeepCopyInto(out *ModelReg) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelReg.
func (in *ModelReg) DeepCopy() *ModelReg {
	if in == nil {
		return nil
	}
	out := new(ModelReg)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ModelReg) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelRegComponentSpec) DeepCopyInto(out *ModelRegComponentSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
	out.ModelReg = in.ModelReg
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelRegComponentSpec.
func (in *ModelRegComponentSpec) DeepCopy() *ModelRegComponentSpec {
	if in == nil {
		return nil
	}
	out := new(ModelRegComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelRegComponentStatus) DeepCopyInto(out *ModelRegComponentStatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelRegComponentStatus.
func (in *ModelRegComponentStatus) DeepCopy() *ModelRegComponentStatus {
	if in == nil {
		return nil
	}
	out := new(ModelRegComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelRegCustomizedSpec) DeepCopyInto(out *ModelRegCustomizedSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelRegCustomizedSpec.
func (in *ModelRegCustomizedSpec) DeepCopy() *ModelRegCustomizedSpec {
	if in == nil {
		return nil
	}
	out := new(ModelRegCustomizedSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelRegList) DeepCopyInto(out *ModelRegList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ModelReg, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelRegList.
func (in *ModelRegList) DeepCopy() *ModelRegList {
	if in == nil {
		return nil
	}
	out := new(ModelRegList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ModelRegList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelRegistryStatus) DeepCopyInto(out *ModelRegistryStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelRegistryStatus.
func (in *ModelRegistryStatus) DeepCopy() *ModelRegistryStatus {
	if in == nil {
		return nil
	}
	out := new(ModelRegistryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Ray) DeepCopyInto(out *Ray) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Ray.
func (in *Ray) DeepCopy() *Ray {
	if in == nil {
		return nil
	}
	out := new(Ray)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Ray) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RayComponentSpec) DeepCopyInto(out *RayComponentSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RayComponentSpec.
func (in *RayComponentSpec) DeepCopy() *RayComponentSpec {
	if in == nil {
		return nil
	}
	out := new(RayComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RayComponentStatus) DeepCopyInto(out *RayComponentStatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RayComponentStatus.
func (in *RayComponentStatus) DeepCopy() *RayComponentStatus {
	if in == nil {
		return nil
	}
	out := new(RayComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RayList) DeepCopyInto(out *RayList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Ray, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RayList.
func (in *RayList) DeepCopy() *RayList {
	if in == nil {
		return nil
	}
	out := new(RayList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RayList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrainingOperator) DeepCopyInto(out *TrainingOperator) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrainingOperator.
func (in *TrainingOperator) DeepCopy() *TrainingOperator {
	if in == nil {
		return nil
	}
	out := new(TrainingOperator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TrainingOperator) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrainingOperatorList) DeepCopyInto(out *TrainingOperatorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TrainingOperator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrainingOperatorList.
func (in *TrainingOperatorList) DeepCopy() *TrainingOperatorList {
	if in == nil {
		return nil
	}
	out := new(TrainingOperatorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TrainingOperatorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrustyAI) DeepCopyInto(out *TrustyAI) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrustyAI.
func (in *TrustyAI) DeepCopy() *TrustyAI {
	if in == nil {
		return nil
	}
	out := new(TrustyAI)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TrustyAI) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrustyAIComponentSpec) DeepCopyInto(out *TrustyAIComponentSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrustyAIComponentSpec.
func (in *TrustyAIComponentSpec) DeepCopy() *TrustyAIComponentSpec {
	if in == nil {
		return nil
	}
	out := new(TrustyAIComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrustyAIComponentStatus) DeepCopyInto(out *TrustyAIComponentStatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrustyAIComponentStatus.
func (in *TrustyAIComponentStatus) DeepCopy() *TrustyAIComponentStatus {
	if in == nil {
		return nil
	}
	out := new(TrustyAIComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrustyAIList) DeepCopyInto(out *TrustyAIList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TrustyAI, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrustyAIList.
func (in *TrustyAIList) DeepCopy() *TrustyAIList {
	if in == nil {
		return nil
	}
	out := new(TrustyAIList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TrustyAIList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WBComponentSpec) DeepCopyInto(out *WBComponentSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WBComponentSpec.
func (in *WBComponentSpec) DeepCopy() *WBComponentSpec {
	if in == nil {
		return nil
	}
	out := new(WBComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WBComponentStatus) DeepCopyInto(out *WBComponentStatus) {
	*out = *in
	in.ComponentStatus.DeepCopyInto(&out.ComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WBComponentStatus.
func (in *WBComponentStatus) DeepCopy() *WBComponentStatus {
	if in == nil {
		return nil
	}
	out := new(WBComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Workbenches) DeepCopyInto(out *Workbenches) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Workbenches.
func (in *Workbenches) DeepCopy() *Workbenches {
	if in == nil {
		return nil
	}
	out := new(Workbenches)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Workbenches) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkbenchesList) DeepCopyInto(out *WorkbenchesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Workbenches, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkbenchesList.
func (in *WorkbenchesList) DeepCopy() *WorkbenchesList {
	if in == nil {
		return nil
	}
	out := new(WorkbenchesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WorkbenchesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
