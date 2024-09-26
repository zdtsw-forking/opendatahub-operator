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
	infrav1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/infrastructure/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=dsckserve
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=.status.phase,description="Status"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=.metadata.creationTimestamp,description="The age of the resource"
type DSCKserve struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KserveComponentSpec   `json:"spec,omitempty"`
	Status KserveComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:validation:Pattern=`^(Serverless|RawDeployment)$`
type DefaultDeploymentMode string

type KserveComponentSpec struct {
	DSCComponentSpec `json:",inline"` // Embedded DSCComponentSpec
	// Serving configures the KNative-Serving stack used for model serving. A Service
	// Mesh (Istio) is prerequisite, since it is used as networking layer.
	Serving infrav1.ServingSpec `json:"serving,omitempty"`
	// Configures the default deployment mode for Kserve. This can be set to 'Serverless' or 'RawDeployment'.
	// The value specified in this field will be used to set the default deployment mode in the 'inferenceservice-config' configmap for Kserve.
	// This field is optional. If no default deployment mode is specified, Kserve will use Serverless mode.
	// +kubebuilder:validation:Enum=Serverless;RawDeployment
	DefaultDeploymentMode DefaultDeploymentMode `json:"defaultDeploymentMode,omitempty"`
}

// KserveComponentStatus defines the custom status of DSCCodeFlare.
type KserveComponentStatus struct {
	DSCComponentStatus `json:",inline"` // Embedded DSCComponentStatus
}
