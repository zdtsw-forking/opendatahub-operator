package workbenches

import (
	"context"
	"fmt"

	operatorv1 "github.com/openshift/api/operator/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opendatahub-io/opendatahub-operator/v2/apis/components"
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
	return componentsv1.WorkbenchesComponentName
}

func (s *componentHandler) GetManagementState(dsc *dscv1.DataScienceCluster) operatorv1.ManagementState {
	if dsc.Spec.Components.Workbenches.ManagementState == operatorv1.Managed {
		return operatorv1.Managed
	}
	return operatorv1.Removed
}
func (s *componentHandler) UpdateDSCStatus(ctx context.Context, dsc *dscv1.DataScienceCluster, c client.Object) string {
	w, _ := c.(*componentsv1.Workbenches)
	if dsc.Status.Components.Workbenches == nil {
		dsc.Status.Components.Workbenches = &components.Status{}
	}
	dsc.Status.Components.Workbenches.Phase = w.Status.Phase

	var r, m string
	st := corev1.ConditionUnknown
	if rc := meta.FindStatusCondition(w.Status.Conditions, status.ConditionTypeReady); rc != nil {
		r = rc.Reason
		m = rc.Message
		st = corev1.ConditionStatus(rc.Status)
	}
	status.SetComponentCondition(&dsc.Status.Conditions, componentsv1.WorkbenchesKind, r, m, st)
	return w.Status.Phase
}

func (s *componentHandler) NewCRObject(dsc *dscv1.DataScienceCluster) client.Object {
	workbenchesAnnotations := make(map[string]string)
	workbenchesAnnotations[annotations.ManagementStateAnnotation] = string(s.GetManagementState(dsc))

	return client.Object(&componentsv1.Workbenches{
		TypeMeta: metav1.TypeMeta{
			Kind:       componentsv1.WorkbenchesKind,
			APIVersion: componentsv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        componentsv1.WorkbenchesInstanceName,
			Annotations: workbenchesAnnotations,
		},
		Spec: componentsv1.WorkbenchesSpec{
			WorkbenchesCommonSpec: dsc.Spec.Components.Workbenches.WorkbenchesCommonSpec,
		},
	})
}

func (s *componentHandler) Init(platform cluster.Platform) error {
	nbcManifestInfo := notebookControllerManifestInfo(notebookControllerManifestSourcePath)
	if err := odhdeploy.ApplyParams(nbcManifestInfo.String(), map[string]string{
		"odh-notebook-controller-image": "RELATED_IMAGE_ODH_NOTEBOOK_CONTROLLER_IMAGE",
	}); err != nil {
		return fmt.Errorf("failed to update params.env from %s : %w", nbcManifestInfo.String(), err)
	}

	kfNbcManifestInfo := kfNotebookControllerManifestInfo(kfNotebookControllerManifestSourcePath)
	if err := odhdeploy.ApplyParams(kfNbcManifestInfo.String(), map[string]string{
		"odh-kf-notebook-controller-image": "RELATED_IMAGE_ODH_KF_NOTEBOOK_CONTROLLER_IMAGE",
	}); err != nil {
		return fmt.Errorf("failed to update params.env from %s : %w", kfNbcManifestInfo.String(), err)
	}

	return nil
}
