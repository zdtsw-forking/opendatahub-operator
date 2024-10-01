// Package dashboard provides utility functions to config Open Data Hub Dashboard: A web dashboard that displays
// installed Open Data Hub components with easy access to component UIs and documentation
// +groupName=datasciencecluster.opendatahub.io
package dashboard

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	"github.com/opendatahub-io/opendatahub-operator/v2/components"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
)

// Verifies that Dashboard implements ComponentInterface.
var _ components.ComponentInterface = (*Dashboard)(nil)

// Dashboard struct holds the configuration for the Dashboard component.
// +kubebuilder:object:generate=true
type Dashboard struct {
	components.Component `json:""`
}

func updateKustomizeVariable(ctx context.Context, cli client.Client, platform cluster.Platform, componentSpec *dsccomponentv1alpha1.ComponentSpec) (map[string]string, error) {
	adminGroups := map[cluster.Platform]string{
		cluster.SelfManagedRhods: "rhods-admins",
		cluster.ManagedRhods:     "dedicated-admins",
		cluster.OpenDataHub:      "odh-admins",
		cluster.Unknown:          "odh-admins",
	}[platform]

	sectionTitle := map[cluster.Platform]string{
		cluster.SelfManagedRhods: "OpenShift Self Managed Services",
		cluster.ManagedRhods:     "OpenShift Managed Services",
		cluster.OpenDataHub:      "OpenShift Open Data Hub",
		cluster.Unknown:          "OpenShift Open Data Hub",
	}[platform]

	consoleLinkDomain, err := cluster.GetDomain(ctx, cli)
	if err != nil {
		return nil, fmt.Errorf("error getting console route URL %s : %w", consoleLinkDomain, err)
	}
	consoleURL := map[cluster.Platform]string{
		cluster.SelfManagedRhods: "https://rhods-dashboard-" + componentSpec.DSCISpec.ApplicationsNamespace + "." + consoleLinkDomain,
		cluster.ManagedRhods:     "https://rhods-dashboard-" + componentSpec.DSCISpec.ApplicationsNamespace + "." + consoleLinkDomain,
		cluster.OpenDataHub:      "https://odh-dashboard-" + componentSpec.DSCISpec.ApplicationsNamespace + "." + consoleLinkDomain,
		cluster.Unknown:          "https://odh-dashboard-" + componentSpec.DSCISpec.ApplicationsNamespace + "." + consoleLinkDomain,
	}[platform]

	return map[string]string{
		"admin_groups":  adminGroups,
		"dashboard-url": consoleURL,
		"section-title": sectionTitle,
	}, nil
}
