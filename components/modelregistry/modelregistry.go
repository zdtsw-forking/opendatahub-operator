// Package modelregistry provides utility functions to config ModelRegistry, an ML Model metadata repository service
// +groupName=datasciencecluster.opendatahub.io
package modelregistry

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-logr/logr"
	operatorv1 "github.com/openshift/api/operator/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	infrav1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/infrastructure/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/components"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/conversion"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"

	_ "embed"
)

func (m *ModelRegistry) createDependencies(ctx context.Context, cli client.Client, componentSpec *dsccomponentv1alpha1.ComponentSpec) error {
	// create DefaultModelRegistryCert
	if err := cluster.PropagateDefaultIngressCertificate(ctx, cli, DefaultModelRegistryCert, componentSpec.ServiceMesh.ControlPlane.Namespace); err != nil {
		return err
	}
	return nil
}

func (m *ModelRegistry) removeDependencies(ctx context.Context, cli client.Client, componentSpec *dsccomponentv1alpha1.ComponentSpec) error {
	// delete DefaultModelRegistryCert
	certSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultModelRegistryCert,
			Namespace: componentSpec.ServiceMesh.ControlPlane.Namespace,
		},
	}
	// ignore error if the secret has already been removed
	if err := cli.Delete(ctx, &certSecret); client.IgnoreNotFound(err) != nil {
		return err
	}
	return nil
}

//go:embed resources/servicemesh-member.tmpl.yaml
var smmTemplate string

func enrollToServiceMesh(ctx context.Context, cli client.Client, componentSpec *dsccomponentv1alpha1.ComponentSpec, namespace *corev1.Namespace) error {
	tmpl, err := template.New("servicemeshmember").Parse(smmTemplate)
	if err != nil {
		return fmt.Errorf("error parsing servicemeshmember template: %w", err)
	}
	builder := strings.Builder{}
	controlPlaneData := struct {
		Namespace    string
		ControlPlane *infrav1.ControlPlaneSpec
	}{Namespace: namespace.Name, ControlPlane: &componentSpec.ServiceMesh.ControlPlane}

	if err = tmpl.Execute(&builder, controlPlaneData); err != nil {
		return fmt.Errorf("error executing servicemeshmember template: %w", err)
	}

	unstrObj, err := conversion.StrToUnstructured(builder.String())
	if err != nil || len(unstrObj) != 1 {
		return fmt.Errorf("error converting servicemeshmember template: %w", err)
	}

	return client.IgnoreAlreadyExists(cli.Create(ctx, unstrObj[0]))
}
