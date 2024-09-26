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

var (
	ComponentName     = "codeflare"
	CodeflarePath     = deploy.DefaultManifestPath + "/" + ComponentName + "/default"
	CodeflareOperator = "codeflare-operator"
	ParamsPath        = deploy.DefaultManifestPath + "/" + ComponentName + "/manager"
)

// Verifies that CodeFlare implements ComponentInterface.
var _ components.ComponentInterface = (*CodeFlare)(nil)

// CodeFlare struct holds the configuration for the CodeFlare component.
// +kubebuilder:object:generate=true
type CodeFlare struct {
	components.Component `json:""`
}

func (d *CodeFlare) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {

	// create/delete CodeFlare Component CR
	cfoCR := &dsccomponentv1alpha1.CodeFlare{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CodeFlare",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default-cfo",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.CodeFlareSpec{
			ComponentSpec: dsccomponentv1alpha1.ComponentSpec{
				Platform:              dsci.Status.Release.Name,
				ComponentName:         ComponentName,
				ApplicationsNamespace: dsci.Spec.ApplicationsNamespace,
				Monitoring:            dsci.Spec.Monitoring,
			},
		},
	}
	if enabled {
		cli.Create(ctx, cfoCR)
	} else {
		cli.Delete(ctx, cfoCR)
	}
	return nil
}
func (c *CodeFlare) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(c.DevFlags.Manifests) != 0 {
		manifestConfig := c.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentName, manifestConfig); err != nil {
			return err
		}
		// If overlay is defined, update paths
		defaultKustomizePath := "default"
		if manifestConfig.SourcePath != "" {
			defaultKustomizePath = manifestConfig.SourcePath
		}
		CodeflarePath = filepath.Join(deploy.DefaultManifestPath, ComponentName, defaultKustomizePath)
	}

	return nil
}

func (c *CodeFlare) GetComponentName() string {
	return ComponentName
}

func (c *CodeFlare) ReconcileComponent(ctx context.Context,
	cli client.Client,
	l logr.Logger,
	owner metav1.Object,
	componentSpec *dsccomponentv1alpha1.ComponentSpec,
	_ bool) error {
	var imageParamMap = map[string]string{
		"codeflare-operator-controller-image": "RELATED_IMAGE_ODH_CODEFLARE_OPERATOR_IMAGE", // no need mcad, embedded in cfo
	}

	enabled := c.GetManagementState() == operatorv1.Managed
	monitoringEnabled := componentSpec.Monitoring.ManagementState == operatorv1.Managed

	if enabled {
		if c.DevFlags != nil {
			// Download manifests and update paths
			if err := c.OverrideManifests(ctx, componentSpec.Platform); err != nil {
				return err
			}
		}
		// check if the CodeFlare operator is installed: it should not be installed
		// Both ODH and RHOAI should have the same operator name
		dependentOperator := CodeflareOperator

		if found, err := cluster.OperatorExists(ctx, cli, dependentOperator); err != nil {
			return fmt.Errorf("operator exists throws error %w", err)
		} else if found {
			return fmt.Errorf("operator %s is found. Please uninstall the operator before enabling %s component",
				dependentOperator, ComponentName)
		}

		// Update image parameters only when we do not have customized manifests set
		if c.DevFlags == nil || len(c.DevFlags.Manifests) == 0 {
			if err := deploy.ApplyParams(ParamsPath, imageParamMap, map[string]string{"namespace": componentSpec.ApplicationsNamespace}); err != nil {
				return fmt.Errorf("failed update image from %s : %w", CodeflarePath+"/bases", err)
			}
		}
	}

	// Deploy Codeflare
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, //nolint:revive,nolintlint
		CodeflarePath,
		componentSpec.ApplicationsNamespace,
		ComponentName, enabled); err != nil {
		return err
	}
	l.Info("apply manifests done")

	if enabled {
		if err := cluster.WaitForDeploymentAvailable(ctx, cli, ComponentName, componentSpec.ApplicationsNamespace, 20, 2); err != nil {
			return fmt.Errorf("deployment for %s is not ready to server: %w", ComponentName, err)
		}
	}

	// CloudServiceMonitoring handling
	if componentSpec.Platform == cluster.ManagedRhods {
		// inject prometheus codeflare*.rules in to /opt/manifests/monitoring/prometheus/prometheus-configs.yaml
		if err := c.UpdatePrometheusConfig(cli, l, enabled && monitoringEnabled, ComponentName); err != nil {
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
