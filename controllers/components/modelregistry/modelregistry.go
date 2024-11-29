package modelregistry

import (
	"context"
	"fmt"

	operatorv1 "github.com/openshift/api/operator/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	componentsv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1"
	dscv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/datasciencecluster/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/controllers/status"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	cr "github.com/opendatahub-io/opendatahub-operator/v2/pkg/componentsregistry"
	odhdeploy "github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
)

type componentHandler struct{}

func init() { //nolint:gochecknoinits
	cr.Add(&componentHandler{})
}

func (s *componentHandler) GetName() string {
	return componentsv1.ModelRegistryComponentName
}

func (s *componentHandler) GetManagementState(dsc *dscv1.DataScienceCluster) operatorv1.ManagementState {
	if dsc.Spec.Components.ModelRegistry.ManagementState == operatorv1.Managed {
		return operatorv1.Managed
	}
	return operatorv1.Removed
}

func (s *componentHandler) Init(_ cluster.Platform) error {
	mi := baseManifestInfo(BaseManifestsSourcePath)

	params := make(map[string]string)
	for k, v := range imagesMap {
		params[k] = v
	}
	for k, v := range extraParamsMap {
		params[k] = v
	}

	if err := odhdeploy.ApplyParams(mi.String(), params); err != nil {
		return fmt.Errorf("failed to update images on path %s: %w", mi, err)
	}

	return nil
}

func (s *componentHandler) NewCRObject(dsc *dscv1.DataScienceCluster) client.Object {
	componentAnnotations := make(map[string]string)
	componentAnnotations[annotations.ManagementStateAnnotation] = string(s.GetManagementState(dsc))
	return client.Object(&componentsv1.ModelRegistry{
		TypeMeta: metav1.TypeMeta{
			Kind:       componentsv1.ModelRegistryKind,
			APIVersion: componentsv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        componentsv1.ModelRegistryInstanceName,
			Annotations: componentAnnotations,
		},
		Spec: componentsv1.ModelRegistrySpec{
			ModelRegistryCommonSpec: dsc.Spec.Components.ModelRegistry.ModelRegistryCommonSpec,
		},
	})
}
func (s *componentHandler) UpdateDSCStatus(ctx context.Context, dsc *dscv1.DataScienceCluster, c client.Object) string {
	mr, _ := c.(*componentsv1.ModelRegistry)

	if dsc.Status.Components.ModelRegistry == nil {
		dsc.Status.Components.ModelRegistry = &componentsv1.ModelRegistryStatus{}
	}
	dsc.Status.Components.ModelRegistry.Phase = mr.Status.Phase
	dsc.Status.Components.ModelRegistry.RegistriesNamespace = mr.Status.RegistriesNamespace

	var r, m string
	st := corev1.ConditionUnknown
	rc := meta.FindStatusCondition(mr.Status.Conditions, status.ConditionTypeReady)
	if rc != nil {
		r = rc.Reason
		m = rc.Message
		st = corev1.ConditionStatus(rc.Status)
	}
	status.SetComponentCondition(&dsc.Status.Conditions, componentsv1.ModelRegistryKind, r, m, st)
	return mr.Status.Phase
}
