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
var _ components.ComponentInterface = (*TrustyAI)(nil)

// TrustyAI struct holds the configuration for the TrustyAI component.
// +kubebuilder:object:generate=true
type TrustyAI struct {
	components.Component `json:""`
}

func (t *TrustyAI) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete TrustyAI Component CR
	trustyaiCR := &dsccomponentv1alpha1.TrustyAI{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrustyAI",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default-trustyai",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.TrustyAIComponentSpec{
			ComponentSpec: dsccomponentv1alpha1.ComponentSpec{
				Platform:              dsci.Status.Release.Name,
				ComponentName:         ComponentName,
				ApplicationsNamespace: dsci.Spec.ApplicationsNamespace,
				Monitoring:            dsci.Spec.Monitoring,
			},
		},
	}
	if enabled {
		cli.Create(ctx, trustyaiCR)
	} else {
		cli.Delete(ctx, trustyaiCR)
	}
	return nil
}
func (t *TrustyAI) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(t.DevFlags.Manifests) != 0 {
		manifestConfig := t.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentPathName, manifestConfig); err != nil {
			return err
		}
		// If overlay is defined, update paths
		defaultKustomizePath := "base"
		if manifestConfig.SourcePath != "" {
			defaultKustomizePath = manifestConfig.SourcePath
		}
		OverridePath = filepath.Join(deploy.DefaultManifestPath, ComponentPathName, defaultKustomizePath)
	}
	return nil
}

func (t *TrustyAI) GetComponentName() string {
	return ComponentName
}

func (t *TrustyAI) ReconcileComponent(ctx context.Context, cli client.Client, l logr.Logger,
	owner metav1.Object, componentSpec *dsccomponentv1alpha1.ComponentSpec, _ bool) error {
	var imageParamMap = map[string]string{
		"trustyaiServiceImage":  "RELATED_IMAGE_ODH_TRUSTYAI_SERVICE_IMAGE",
		"trustyaiOperatorImage": "RELATED_IMAGE_ODH_TRUSTYAI_SERVICE_OPERATOR_IMAGE",
	}
	entryPath := map[cluster.Platform]string{
		cluster.SelfManagedRhods: PathDownstream,
		cluster.ManagedRhods:     PathDownstream,
		cluster.OpenDataHub:      PathUpstream,
		cluster.Unknown:          PathUpstream,
	}[componentSpec.Platform]

	enabled := t.GetManagementState() == operatorv1.Managed
	monitoringEnabled := componentSpec.Monitoring.ManagementState == operatorv1.Managed

	if enabled {
		if t.DevFlags != nil {
			// Download manifests and update paths
			if err := t.OverrideManifests(ctx, componentSpec.Platform); err != nil {
				return err
			}
			if OverridePath != "" {
				entryPath = OverridePath
			}
		}
		if t.DevFlags == nil || len(t.DevFlags.Manifests) == 0 {
			if err := deploy.ApplyParams(entryPath, imageParamMap); err != nil {
				return fmt.Errorf("failed to update image %s: %w", entryPath, err)
			}
		}
	}
	// Deploy TrustyAI Operator
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, entryPath, componentSpec.ApplicationsNamespace, t.GetComponentName(), enabled); err != nil {
		return err
	}
	l.Info("apply manifests done")

	// Wait for deployment available
	if enabled {
		if err := cluster.WaitForDeploymentAvailable(ctx, cli, ComponentName, componentSpec.ApplicationsNamespace, 10, 2); err != nil {
			return fmt.Errorf("deployment for %s is not ready to server: %w", ComponentName, err)
		}
	}

	// CloudService Monitoring handling
	if componentSpec.Platform == cluster.ManagedRhods {
		if err := t.UpdatePrometheusConfig(cli, l, enabled && monitoringEnabled, ComponentName); err != nil {
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
