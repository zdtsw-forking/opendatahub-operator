// Package datasciencepipelines provides utility functions to config Data Science Pipelines:
// Pipeline solution for end to end MLOps workflows that support the Kubeflow Pipelines SDK, Tekton and Argo Workflows.
// +groupName=datasciencecluster.opendatahub.io
package datasciencepipelines

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	operatorv1 "github.com/openshift/api/operator/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	conditionsv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/components"
	"github.com/opendatahub-io/opendatahub-operator/v2/controllers/status"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
)


// Verifies that Dashboard implements ComponentInterface.
var _ components.ComponentInterface = (*DataSciencePipelines)(nil)

// DataSciencePipelines struct holds the configuration for the DataSciencePipelines component.
// +kubebuilder:object:generate=true
type DataSciencePipelines struct {
	components.Component `json:""`
}


func UnmanagedArgoWorkFlowExists(ctx context.Context,
	cli client.Client) error {
	workflowCRD := &apiextensionsv1.CustomResourceDefinition{}
	if err := cli.Get(ctx, client.ObjectKey{Name: ArgoWorkflowCRD}, workflowCRD); err != nil {
		if k8serr.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get existing Workflow CRD : %w", err)
	}
	// Verify if existing workflow is deployed by ODH with label
	odhLabelValue, odhLabelExists := workflowCRD.Labels[labels.ODH.Component(ComponentName)]
	if odhLabelExists && odhLabelValue == "true" {
		return nil
	}
	return fmt.Errorf("%s CRD already exists but not deployed by this operator. "+
		"Remove existing Argo workflows or set `spec.components.datasciencepipelines.managementState` to Removed to proceed ", ArgoWorkflowCRD)
}

func SetExistingArgoCondition(conditions *[]conditionsv1.Condition, reason, message string) {
	status.SetCondition(conditions, string(status.CapabilityDSPv2Argo), reason, message, corev1.ConditionFalse)
	status.SetComponentCondition(conditions, ComponentName, status.ReconcileFailed, message, corev1.ConditionFalse)
}
