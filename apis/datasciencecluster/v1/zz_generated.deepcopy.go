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

package v1

import (
	"github.com/opendatahub-io/opendatahub-operator/v2/apis/components"
	componentsv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Components) DeepCopyInto(out *Components) {
	*out = *in
	in.Dashboard.DeepCopyInto(&out.Dashboard)
	in.Workbenches.DeepCopyInto(&out.Workbenches)
	in.ModelMeshServing.DeepCopyInto(&out.ModelMeshServing)
	in.DataSciencePipelines.DeepCopyInto(&out.DataSciencePipelines)
	in.Kserve.DeepCopyInto(&out.Kserve)
	in.Kueue.DeepCopyInto(&out.Kueue)
	in.CodeFlare.DeepCopyInto(&out.CodeFlare)
	in.Ray.DeepCopyInto(&out.Ray)
	in.TrustyAI.DeepCopyInto(&out.TrustyAI)
	in.ModelRegistry.DeepCopyInto(&out.ModelRegistry)
	in.TrainingOperator.DeepCopyInto(&out.TrainingOperator)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Components.
func (in *Components) DeepCopy() *Components {
	if in == nil {
		return nil
	}
	out := new(Components)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComponentsStatus) DeepCopyInto(out *ComponentsStatus) {
	*out = *in
	if in.ModelRegistry != nil {
		in, out := &in.ModelRegistry, &out.ModelRegistry
		*out = new(componentsv1.ModelRegistryStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.Dashboard != nil {
		in, out := &in.Dashboard, &out.Dashboard
		*out = new(componentsv1.DashboardStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.CodeFlare != nil {
		in, out := &in.CodeFlare, &out.CodeFlare
		*out = new(components.Status)
		(*in).DeepCopyInto(*out)
	}
	if in.Ray != nil {
		in, out := &in.Ray, &out.Ray
		*out = new(components.Status)
		(*in).DeepCopyInto(*out)
	}
	if in.TrustyAI != nil {
		in, out := &in.TrustyAI, &out.TrustyAI
		*out = new(components.Status)
		(*in).DeepCopyInto(*out)
	}
	if in.TrainingOperator != nil {
		in, out := &in.TrainingOperator, &out.TrainingOperator
		*out = new(components.Status)
		(*in).DeepCopyInto(*out)
	}
	if in.Kueue != nil {
		in, out := &in.Kueue, &out.Kueue
		*out = new(components.Status)
		(*in).DeepCopyInto(*out)
	}
	if in.DataSciencePipelines != nil {
		in, out := &in.DataSciencePipelines, &out.DataSciencePipelines
		*out = new(components.Status)
		(*in).DeepCopyInto(*out)
	}
	if in.Kserve != nil {
		in, out := &in.Kserve, &out.Kserve
		*out = new(components.Status)
		(*in).DeepCopyInto(*out)
	}
	if in.ModelMeshServing != nil {
		in, out := &in.ModelMeshServing, &out.ModelMeshServing
		*out = new(components.Status)
		(*in).DeepCopyInto(*out)
	}
	if in.Workbenches != nil {
		in, out := &in.Workbenches, &out.Workbenches
		*out = new(components.Status)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComponentsStatus.
func (in *ComponentsStatus) DeepCopy() *ComponentsStatus {
	if in == nil {
		return nil
	}
	out := new(ComponentsStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataScienceCluster) DeepCopyInto(out *DataScienceCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataScienceCluster.
func (in *DataScienceCluster) DeepCopy() *DataScienceCluster {
	if in == nil {
		return nil
	}
	out := new(DataScienceCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DataScienceCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataScienceClusterList) DeepCopyInto(out *DataScienceClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DataScienceCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataScienceClusterList.
func (in *DataScienceClusterList) DeepCopy() *DataScienceClusterList {
	if in == nil {
		return nil
	}
	out := new(DataScienceClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DataScienceClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataScienceClusterSpec) DeepCopyInto(out *DataScienceClusterSpec) {
	*out = *in
	in.Components.DeepCopyInto(&out.Components)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataScienceClusterSpec.
func (in *DataScienceClusterSpec) DeepCopy() *DataScienceClusterSpec {
	if in == nil {
		return nil
	}
	out := new(DataScienceClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataScienceClusterStatus) DeepCopyInto(out *DataScienceClusterStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]conditionsv1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.RelatedObjects != nil {
		in, out := &in.RelatedObjects, &out.RelatedObjects
		*out = make([]corev1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.InstalledComponents != nil {
		in, out := &in.InstalledComponents, &out.InstalledComponents
		*out = make(map[string]bool, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Components != nil {
		in, out := &in.Components, &out.Components
		*out = new(ComponentsStatus)
		(*in).DeepCopyInto(*out)
	}
	in.Release.DeepCopyInto(&out.Release)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataScienceClusterStatus.
func (in *DataScienceClusterStatus) DeepCopy() *DataScienceClusterStatus {
	if in == nil {
		return nil
	}
	out := new(DataScienceClusterStatus)
	in.DeepCopyInto(out)
	return out
}
