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
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
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

var (
	ComponentName   = "data-science-pipelines-operator"
	Path            = deploy.DefaultManifestPath + "/" + ComponentName + "/base"
	OverlayPath     = deploy.DefaultManifestPath + "/" + ComponentName + "/overlays"
	ArgoWorkflowCRD = "workflows.argoproj.io"
)

// Verifies that Dashboard implements ComponentInterface.
var _ components.ComponentInterface = (*DataSciencePipelines)(nil)

// DataSciencePipelines struct holds the configuration for the DataSciencePipelines component.
// +kubebuilder:object:generate=true
type DataSciencePipelines struct {
	components.Component `json:""`
}

func (d *DataSciencePipelines) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete DataSciencePipelines Component CR
	dspCR := &dsccomponentv1alpha1.DataSciencePipeline{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DataSciencePipeline",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default-dsp",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.DataSciencePipelineSpec{
			ComponentSpec: dsccomponentv1alpha1.ComponentSpec{
				Platform:              dsci.Status.Release.Name,
				ComponentName:         ComponentName,
				ApplicationsNamespace: dsci.Spec.ApplicationsNamespace,
				Monitoring:            dsci.Spec.Monitoring,
			},
		},
	}
	if enabled {
		cli.Create(ctx, dspCR)
	} else {
		cli.Delete(ctx, dspCR)
	}
	return nil
}
func (d *DataSciencePipelines) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(d.DevFlags.Manifests) != 0 {
		manifestConfig := d.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentName, manifestConfig); err != nil {
			return err
		}
		// If overlay is defined, update paths
		defaultKustomizePath := "base"
		if manifestConfig.SourcePath != "" {
			defaultKustomizePath = manifestConfig.SourcePath
		}
		Path = filepath.Join(deploy.DefaultManifestPath, ComponentName, defaultKustomizePath)
	}

	return nil
}

func (d *DataSciencePipelines) GetComponentName() string {
	return ComponentName
}

func (d *DataSciencePipelines) ReconcileComponent(ctx context.Context,
	cli client.Client,
	l logr.Logger,
	owner metav1.Object,
	componentSpec *dsccomponentv1alpha1.ComponentSpec,
	currentComponentExist bool,
) error {
	var imageParamMap = map[string]string{
		// v1
		"IMAGES_APISERVER":         "RELATED_IMAGE_ODH_ML_PIPELINES_API_SERVER_IMAGE",
		"IMAGES_ARTIFACT":          "RELATED_IMAGE_ODH_ML_PIPELINES_ARTIFACT_MANAGER_IMAGE",
		"IMAGES_PERSISTENTAGENT":   "RELATED_IMAGE_ODH_ML_PIPELINES_PERSISTENCEAGENT_IMAGE",
		"IMAGES_SCHEDULEDWORKFLOW": "RELATED_IMAGE_ODH_ML_PIPELINES_SCHEDULEDWORKFLOW_IMAGE",
		"IMAGES_CACHE":             "RELATED_IMAGE_ODH_ML_PIPELINES_CACHE_IMAGE",
		"IMAGES_DSPO":              "RELATED_IMAGE_ODH_DATA_SCIENCE_PIPELINES_OPERATOR_CONTROLLER_IMAGE",
		// v2
		"IMAGESV2_ARGO_APISERVER":          "RELATED_IMAGE_ODH_ML_PIPELINES_API_SERVER_V2_IMAGE",
		"IMAGESV2_ARGO_PERSISTENCEAGENT":   "RELATED_IMAGE_ODH_ML_PIPELINES_PERSISTENCEAGENT_V2_IMAGE",
		"IMAGESV2_ARGO_SCHEDULEDWORKFLOW":  "RELATED_IMAGE_ODH_ML_PIPELINES_SCHEDULEDWORKFLOW_V2_IMAGE",
		"IMAGESV2_ARGO_ARGOEXEC":           "RELATED_IMAGE_ODH_DATA_SCIENCE_PIPELINES_ARGO_ARGOEXEC_IMAGE",
		"IMAGESV2_ARGO_WORKFLOWCONTROLLER": "RELATED_IMAGE_ODH_DATA_SCIENCE_PIPELINES_ARGO_WORKFLOWCONTROLLER_IMAGE",
		"V2_DRIVER_IMAGE":                  "RELATED_IMAGE_ODH_ML_PIPELINES_DRIVER_IMAGE",
		"V2_LAUNCHER_IMAGE":                "RELATED_IMAGE_ODH_ML_PIPELINES_LAUNCHER_IMAGE",
		"IMAGESV2_ARGO_MLMDGRPC":           "RELATED_IMAGE_ODH_MLMD_GRPC_SERVER_IMAGE",
	}

	enabled := d.GetManagementState() == operatorv1.Managed
	monitoringEnabled := componentSpec.Monitoring.ManagementState == operatorv1.Managed

	if enabled {
		if d.DevFlags != nil {
			// Download manifests and update paths
			if err := d.OverrideManifests(ctx, componentSpec.Platform); err != nil {
				return err
			}
		}
		// skip check if the dependent operator has beeninstalled, this is done in dashboard
		// Update image parameters only when we do not have customized manifests set
		if d.DevFlags == nil || len(d.DevFlags.Manifests) == 0 {
			if err := deploy.ApplyParams(Path, imageParamMap); err != nil {
				return fmt.Errorf("failed to update image from %s : %w", Path, err)
			}
		}
		// Check for existing Argo Workflows
		if err := UnmanagedArgoWorkFlowExists(ctx, cli); err != nil {
			return err
		}
	}

	// new overlay
	manifestsPath := filepath.Join(OverlayPath, "rhoai")
	if componentSpec.Platform == cluster.OpenDataHub || componentSpec.Platform == "" {
		manifestsPath = filepath.Join(OverlayPath, "odh")
	}
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, manifestsPath, componentSpec.ApplicationsNamespace, ComponentName, enabled); err != nil {
		return err
	}
	l.Info("apply manifests done")

	// Wait for deployment available
	if enabled {
		if err := cluster.WaitForDeploymentAvailable(ctx, cli, ComponentName, componentSpec.ApplicationsNamespace, 20, 2); err != nil {
			return fmt.Errorf("deployment for %s is not ready to server: %w", ComponentName, err)
		}
	}

	// CloudService Monitoring handling
	if componentSpec.Platform == cluster.ManagedRhods {
		if err := d.UpdatePrometheusConfig(cli, l, enabled && monitoringEnabled, ComponentName); err != nil {
			return err
		}
		if err := deploy.DeployManifestsFromPath(ctx, cli, owner,
			filepath.Join(deploy.DefaultManifestPath, "monitoring", "prometheus", "apps"),
			componentSpec.Monitoring.Namespace,
			"prometheus", true); err != nil {
			return err
		}
		l.Info("updating SRE monitoring done")
	}

	return nil
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
