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

package datasciencepipelines

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManagedPipelineOptions) DeepCopyInto(out *ManagedPipelineOptions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManagedPipelineOptions.
func (in *ManagedPipelineOptions) DeepCopy() *ManagedPipelineOptions {
	if in == nil {
		return nil
	}
	out := new(ManagedPipelineOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManagedPipelinesSpec) DeepCopyInto(out *ManagedPipelinesSpec) {
	*out = *in
	out.InstructLab = in.InstructLab
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManagedPipelinesSpec.
func (in *ManagedPipelinesSpec) DeepCopy() *ManagedPipelinesSpec {
	if in == nil {
		return nil
	}
	out := new(ManagedPipelinesSpec)
	in.DeepCopyInto(out)
	return out
}
