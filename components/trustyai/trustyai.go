// Package trustyai provides utility functions to config TrustyAI, a bias/fairness and explainability toolkit
// +groupName=datasciencecluster.opendatahub.io
package trustyai

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	operatorv1 "github.com/openshift/api/operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/components"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
)

var (
	ComponentName     = "trustyai"
	ComponentPathName = "trustyai-service-operator"
	PathUpstream      = deploy.DefaultManifestPath + "/" + ComponentPathName + "/overlays/odh"
	PathDownstream    = deploy.DefaultManifestPath + "/" + ComponentPathName + "/overlays/rhoai"
	OverridePath      = ""
)

// Verifies that TrustyAI implements ComponentInterface.
// var _ components.ComponentInterface = (*TrustyAI)(nil)

// // TrustyAI struct holds the configuration for the TrustyAI component.
// // +kubebuilder:object:generate=true
// type TrustyAI struct {
// 	components.Component `json:""`
// }

