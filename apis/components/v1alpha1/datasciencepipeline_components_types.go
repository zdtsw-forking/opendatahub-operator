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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
		"github.com/opendatahub-io/opendatahub-operator/v2/components/datasciencepipelines"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=.status.phase,description="Status"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=.metadata.creationTimestamp,description="The age of the resource"
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="Ready"
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`,description="Reason"
// +kubebuilder:printcolumn:name="Message",type=string,priority=1,JSONPath=`.status.conditions[?(@.type=="Ready")].message`,description="Message"
type DataSciencePipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataSciencePipelineSpec  `json:"spec,omitempty"`
	Status DataSciencePipelinetatus `json:"status,omitempty"`
}

type DataSciencePipelineSpec struct {
	ComponentSpec `json:",inline"` // Embedded ComponentSpec
	DataSciencePipelines datasciencepipelines.DataSciencePipelines `json:"datasciencepipelines,omitempty"`
}

// DataSciencePipelinetatus defines the custom status.
type DataSciencePipelinetatus struct {
	ComponentStatus `json:",inline"` // Embedded ComponentStatus
}
