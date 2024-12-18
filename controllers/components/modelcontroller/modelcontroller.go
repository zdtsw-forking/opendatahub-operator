package modelcontroller

import (
	"fmt"

	operatorv1 "github.com/openshift/api/operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	componentApi "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	dscv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/datasciencecluster/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/componentsregistry"
	odhdeploy "github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
)

const (
	ComponentName = componentApi.ModelControllerComponentName
)

var DefaultPath = odhdeploy.DefaultManifestPath + "/" + ComponentName + "/base"

type componentHandler struct{}

func init() { //nolint:gochecknoinits
	componentsregistry.Add(&componentHandler{})
}

func (s *componentHandler) GetName() string {
	return componentApi.ModelControllerComponentName
}

func (s *componentHandler) GetManagementState(dsc *dscv1.DataScienceCluster) operatorv1.ManagementState {
	if dsc.Spec.Components.ModelMeshServing.ManagementState == operatorv1.Managed || dsc.Spec.Components.Kserve.ManagementState == operatorv1.Managed {
		return operatorv1.Managed
	}
	return operatorv1.Removed
}

func (s *componentHandler) NewCRObject(dsc *dscv1.DataScienceCluster) client.Object {
	mcAnnotations := make(map[string]string)
	mcAnnotations[annotations.ManagementStateAnnotation] = string(s.GetManagementState(dsc))

	// extra logic to set the management .spec.component.managementState, to not leave blank {}
	kState := operatorv1.Removed
	if dsc.Spec.Components.Kserve.ManagementState == operatorv1.Managed {
		kState = operatorv1.Managed
	}

	mState := operatorv1.Removed
	if dsc.Spec.Components.ModelMeshServing.ManagementState == operatorv1.Managed {
		mState = operatorv1.Managed
	}

	return client.Object(&componentApi.ModelController{
		TypeMeta: metav1.TypeMeta{
			Kind:       componentApi.ModelControllerKind,
			APIVersion: componentApi.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        componentApi.ModelControllerInstanceName,
			Annotations: mcAnnotations,
		},
		Spec: componentApi.ModelControllerSpec{
			ModelMeshServing: &componentApi.ModelControllerMMSpec{
				ManagementState: mState,
				DevFlagsSpec:    dsc.Spec.Components.ModelMeshServing.DevFlagsSpec,
			},
			Kserve: &componentApi.ModelControllerKerveSpec{
				ManagementState: kState,
				DevFlagsSpec:    dsc.Spec.Components.Kserve.DevFlagsSpec,
				NIM:             dsc.Spec.Components.Kserve.NIM,
			},
		},
	})
}

// Init for set images.
func (s *componentHandler) Init(platform cluster.Platform) error {
	var imageParamMap = map[string]string{
		"odh-model-controller": "RELATED_IMAGE_ODH_MODEL_CONTROLLER_IMAGE",
	}
	// Update image parameters
	if err := odhdeploy.ApplyParams(DefaultPath, imageParamMap); err != nil {
		return fmt.Errorf("failed to update images on path %s: %w", DefaultPath, err)
	}

	return nil
}
