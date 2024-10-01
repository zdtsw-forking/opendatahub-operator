// Package kserve provides utility functions to config Kserve as the Controller for serving ML models on arbitrary frameworks
// +groupName=datasciencecluster.opendatahub.io
package kserve

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	operatorv1 "github.com/openshift/api/operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	infrav1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/infrastructure/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/components"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
)

func (k *Kserve) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete Kserve Component CR
	kserveCR := &dsccomponentv1alpha1.Kserve{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Kserve",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default-kserve",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.KserveComponentSpec{
			ComponentSpec: dsccomponentv1alpha1.ComponentSpec{
				Platform:              dsci.Status.Release.Name,
				ComponentName:         ComponentName,
				DSCInitializationSpec: dsci.Spec,
				ComponentDevFlags: dsccomponentv1alpha1.DevFlags{
					LoggerMode: dsci.Spec.DevFlags.LogMode,
				},
			},
		},
	}
	if enabled {
		cli.Create(ctx, kserveCR)
	} else {
		cli.Delete(ctx, kserveCR)
	}
	return nil
}

var (
	ComponentName          = "kserve"
	Path                   = deploy.DefaultManifestPath + "/" + ComponentName + "/overlays/odh"
	DependentComponentName = "odh-model-controller"
	DependentPath          = deploy.DefaultManifestPath + "/" + DependentComponentName + "/base"
	ServiceMeshOperator    = "servicemeshoperator"
	ServerlessOperator     = "serverless-operator"
)

// Verifies that Kserve implements ComponentInterface.
var _ components.ComponentInterface = (*Kserve)(nil)

// +kubebuilder:validation:Pattern=`^(Serverless|RawDeployment)$`
type DefaultDeploymentMode string

var (
	// Serverless will be used as the default deployment mode for Kserve. This requires Serverless and ServiceMesh operators configured as dependencies.
	Serverless DefaultDeploymentMode = "Serverless"
	// RawDeployment will be used as the default deployment mode for Kserve.
	RawDeployment DefaultDeploymentMode = "RawDeployment"
)

// Kserve struct holds the configuration for the Kserve component.
// +kubebuilder:object:generate=true
type Kserve struct {
	components.Component `json:""`
	// Serving configures the KNative-Serving stack used for model serving. A Service
	// Mesh (Istio) is prerequisite, since it is used as networking layer.
	Serving infrav1.ServingSpec `json:"serving,omitempty"`
	// Configures the default deployment mode for Kserve. This can be set to 'Serverless' or 'RawDeployment'.
	// The value specified in this field will be used to set the default deployment mode in the 'inferenceservice-config' configmap for Kserve.
	// This field is optional. If no default deployment mode is specified, Kserve will use Serverless mode.
	// +kubebuilder:validation:Enum=Serverless;RawDeployment
	DefaultDeploymentMode DefaultDeploymentMode `json:"defaultDeploymentMode,omitempty"`
}


func (k *Kserve) Cleanup(ctx context.Context, cli client.Client, owner metav1.Object, instance *dsccomponentv1alpha1.ComponentSpec) error {
	if removeServerlessErr := k.removeServerlessFeatures(ctx, cli, owner, instance); removeServerlessErr != nil {
		return removeServerlessErr
	}

	return k.removeServiceMeshConfigurations(ctx, cli, owner, instance)
}
