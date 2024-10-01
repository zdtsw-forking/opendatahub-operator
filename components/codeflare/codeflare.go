// Package codeflare provides utility functions to config CodeFlare as part of the stack
// which makes managing distributed compute infrastructure in the cloud easy and intuitive for Data Scientists
// +groupName=datasciencecluster.opendatahub.io
package codeflare

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



// Verifies that CodeFlare implements ComponentInterface.
var _ components.ComponentInterface = (*CodeFlare)(nil)

// CodeFlare struct holds the configuration for the CodeFlare component.
// +kubebuilder:object:generate=true
type CodeFlare struct {
	components.Component `json:""`
}

