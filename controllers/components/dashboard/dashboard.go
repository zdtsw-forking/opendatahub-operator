package dashboard

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
	return componentsv1.DashboardComponentName
}

func (s *componentHandler) GetManagementState(dsc *dscv1.DataScienceCluster) operatorv1.ManagementState {
	if dsc.Spec.Components.Dashboard.ManagementState == operatorv1.Managed {
		return operatorv1.Managed
	}
	return operatorv1.Removed
}

func (s *componentHandler) UpdateDSCStatus(ctx context.Context, dsc *dscv1.DataScienceCluster, c client.Object) string {
	d, _ := c.(*componentsv1.Dashboard)
	if dsc.Status.Components.Dashboard == nil {
		dsc.Status.Components.Dashboard = &componentsv1.DashboardStatus{}
	}
	dsc.Status.Components.Dashboard.Phase = d.Status.Phase
	dsc.Status.Components.Dashboard.URL = d.Status.URL

	var r, m string
	st := corev1.ConditionUnknown
	if rc := meta.FindStatusCondition(d.Status.Conditions, status.ConditionTypeReady); rc != nil {
		r = rc.Reason
		m = rc.Message
		st = corev1.ConditionStatus(rc.Status)
	}
	status.SetComponentCondition(&dsc.Status.Conditions, componentsv1.DashboardKind, r, m, st)
	return d.Status.Phase
}

func (s *componentHandler) Init(platform cluster.Platform) error {
	mi := defaultManifestInfo(platform)

	if err := odhdeploy.ApplyParams(mi.String(), imagesMap); err != nil {
		return fmt.Errorf("failed to update images on path %s: %w", manifestPaths[platform], err)
	}

	return nil
}

func (s *componentHandler) NewCRObject(dsc *dscv1.DataScienceCluster) client.Object {
	dashboardAnnotations := make(map[string]string)
	dashboardAnnotations[annotations.ManagementStateAnnotation] = string(s.GetManagementState(dsc))

	return client.Object(&componentsv1.Dashboard{
		TypeMeta: metav1.TypeMeta{
			Kind:       componentsv1.DashboardKind,
			APIVersion: componentsv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        componentsv1.DashboardInstanceName,
			Annotations: dashboardAnnotations,
		},
		Spec: componentsv1.DashboardSpec{
			DashboardCommonSpec: dsc.Spec.Components.Dashboard.DashboardCommonSpec,
		},
	})
}
