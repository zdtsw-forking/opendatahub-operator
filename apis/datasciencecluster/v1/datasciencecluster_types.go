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

package v1

import (
	"errors"
	"reflect"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "github.com/opendatahub-io/opendatahub-operator/v2/components"
	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"

	// "github.com/opendatahub-io/opendatahub-operator/v2/components/codeflare"
	// "github.com/opendatahub-io/opendatahub-operator/v2/components/dashboard"
	// "github.com/opendatahub-io/opendatahub-operator/v2/components/datasciencepipelines"
	// "github.com/opendatahub-io/opendatahub-operator/v2/components/kserve"
	// "github.com/opendatahub-io/opendatahub-operator/v2/components/kueue"
	// "github.com/opendatahub-io/opendatahub-operator/v2/components/modelmeshserving"
	// "github.com/opendatahub-io/opendatahub-operator/v2/components/ray"
	// "github.com/opendatahub-io/opendatahub-operator/v2/components/trainingoperator"
	// "github.com/opendatahub-io/opendatahub-operator/v2/components/trustyai"
	// "github.com/opendatahub-io/opendatahub-operator/v2/components/workbenches"

	//"github.com/opendatahub-io/opendatahub-operator/v2/controllers/status"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"

	operatorv1 "github.com/openshift/api/operator/v1"
)

// DataScienceClusterSpec defines the desired state of the cluster.
type DataScienceClusterSpec struct {
	// Override and fine tune specific component configurations.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=1
	Components Components `json:"components,omitempty"`
}

type Dashboard struct {
	Component `json:""`
}
type Workbenches struct {
	Component `json:""`
}
type ModelMeshServing struct {
	Component `json:""`
}
type DataSciencePipelines struct {
	Component `json:""`
}
type Kserve struct {
	Component `json:""`
}
type Kueue struct {
	Component `json:""`
}
type CodeFlare struct {
	Component `json:""`
}
type Ray struct {
	Component `json:""`
}
type TrainingOperator struct {
	Component `json:""`
}

type TrustyAI struct {
	Component `json:""`
}
type Components struct {
	// Dashboard component configuration.
	Dashboard Dashboard `json:"dashboard,omitempty"`
	// Workbenches component configuration.
	Workbenches Workbenches `json:"workbenches,omitempty"`

	// ModelMeshServing component configuration.
	// Does not support enabled Kserve at the same time
	ModelMeshServing ModelMeshServing `json:"modelmeshserving,omitempty"`

	// DataServicePipeline component configuration.
	// Require OpenShift Pipelines Operator to be installed before enable component
	DataSciencePipelines DataSciencePipelines `json:"datasciencepipelines,omitempty"`

	// Kserve component configuration.
	// Require OpenShift Serverless and OpenShift Service Mesh Operators to be installed before enable component
	// Does not support enabled ModelMeshServing at the same time
	Kserve Kserve `json:"kserve,omitempty"`

	// Kueue component configuration.
	Kueue Kueue `json:"kueue,omitempty"`

	// CodeFlare component configuration.
	// If CodeFlare Operator has been installed in the cluster, it should be uninstalled first before enabled component.
	CodeFlare CodeFlare `json:"codeflare,omitempty"`

	// Ray component configuration.
	Ray Ray `json:"ray,omitempty"`

	// TrustyAI component configuration.
	TrustyAI TrustyAI `json:"trustyai,omitempty"`

	//TrainingOperator component configuration.
	TrainingOperator TrainingOperator `json:"trainingoperator,omitempty"`
}

// ComponentsStatus defines the custom status of DataScienceCluster components.
type ComponentsStatus struct {
	// ModelRegistry component status
	ModelRegistry dsccomponentv1alpha1.ModelRegistryStatus `json:"modelregistry,omitempty"`
}

// DataScienceClusterStatus defines the observed state of DataScienceCluster.
type DataScienceClusterStatus struct {
	// Phase describes the Phase of DataScienceCluster reconciliation state
	// This is used by OLM UI to provide status information to the user
	Phase string `json:"phase,omitempty"`

	// Conditions describes the state of the DataScienceCluster resource.
	// +optional
	Conditions []conditionsv1.Condition `json:"conditions,omitempty"`

	// RelatedObjects is a list of objects created and maintained by this operator.
	// Object references will be added to this list after they have been created AND found in the cluster.
	// +optional
	RelatedObjects []corev1.ObjectReference `json:"relatedObjects,omitempty"`
	ErrorMessage   string                   `json:"errorMessage,omitempty"`

	// List of components with status if installed or not
	InstalledComponents map[string]bool `json:"installedComponents,omitempty"`

	// Expose component's specific status
	// +optional
	Components ComponentsStatus `json:"components,omitempty"`

	// Version and release type
	Release cluster.Release `json:"release,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName=dsc
//+kubebuilder:storageversion

// DataScienceCluster is the Schema for the datascienceclusters API.
type DataScienceCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataScienceClusterSpec   `json:"spec,omitempty"`
	Status DataScienceClusterStatus `json:"status,omitempty"`
}

// +groupName=datasciencecluster.opendatahub.io
type Component struct {
	// Set to one of the following values:
	//
	// - "Managed" : the operator is actively managing the component and trying to keep it active.
	//               It will only upgrade the component if it is safe to do so
	//
	// - "Removed" : the operator is actively managing the component and will not install it,
	//               or if it is installed, the operator will try to remove it
	//
	// +kubebuilder:validation:Enum=Managed;Removed
	ManagementState operatorv1.ManagementState `json:"managementState,omitempty"`
	// Add any other common fields across components below

	// Add developer fields
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=2
	DevFlags *dsccomponentv1alpha1.DSCDevFlags `json:"devFlags,omitempty"`
}

//+kubebuilder:object:root=true

// DataScienceClusterList contains a list of DataScienceCluster.
type DataScienceClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataScienceCluster `json:"items"`
}

// type ComponentInterface interface {
// 	GetManagementState() operatorv1.ManagementState
// 	NewMethod() string
// }

func init() {
	SchemeBuilder.Register(
		&DataScienceCluster{},
		&DataScienceClusterList{},
	)
}

func (d DataScienceCluster) NewMethod() string {
	return "TODO!"
}
func (c *Component) GetManagementState() operatorv1.ManagementState {
	return c.ManagementState
}

func (d *DataScienceCluster) GetComponents() ([]ComponentInterface, error) {
	var allComponents []ComponentInterface

	c := &d.Spec.Components

	definedComponents := reflect.ValueOf(c).Elem()
	for i := 0; i < definedComponents.NumField(); i++ {
		c := definedComponents.Field(i)
		if c.CanAddr() {
			component, ok := c.Addr().Interface().(ComponentInterface)
			if !ok {
				return allComponents, errors.New("this is not a pointer to ComponentInterface")
			}

			allComponents = append(allComponents, component)
		}
	}

	return allComponents, nil
}
