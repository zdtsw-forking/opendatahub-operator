package v1alpha1

import (
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	//"github.com/opendatahub-io/opendatahub-operator/v2/components"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	//"github.com/opendatahub-io/opendatahub-operator/v2/pkg/components"
	//infrav1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/infrastructure/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
)

type ComponentSpec struct {
	Platform              cluster.Platform             `json:"platform,omitempty"`
	ComponentName         string                       `json:"componentName,omitempty"`
	DSCComponentSpec		  DSCComponentSpec				 `json:"dscComponent,omitempty"`
	DSCISpec dsciv1.DSCInitializationSpec `json:"dscinitializationspec,omitempty"`
} 

// ComponentStatus defines the custom status of ComponentSpec.
type ComponentStatus struct {
	// +operator-sdk:csv:customresourcedefinitions:type=status
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Phase      string             `json:"phase,omitempty"`
	ManagementState operatorv1.ManagementState `json:"managementState,omitempty"`
}

type DSCComponentSpec struct {
	ComponentDevFlags     DevFlags                     `json:"componentdevflags,omitempty"`
	ComponentCustomization map[string]interface{}       `json:"componentcustomization,omitempty"`
}

type DevFlags struct {
	LoggerMode  string      `json:"LoggerMode,omitempty"` // TODO: enabled different logging on component based?
	DSCDevFlags DSCDevFlags `json:"dscdevflags,omitempty"`
}

// DevFlags defines list of fields that can be used by developers to test customizations. This is not recommended
// to be used in production environment.
// +kubebuilder:object:generate=true
type DSCDevFlags struct {
	// List of custom manifests for the given component
	// +optional
	Manifests []ManifestsConfig `json:"manifests,omitempty"`
}

type ManifestsConfig struct {
	// uri is the URI point to a git repo with tag/branch. e.g.  https://github.com/org/repo/tarball/<tag/branch>
	// +optional
	// +kubebuilder:default:=""
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=1
	URI string `json:"uri,omitempty"`

	// contextDir is the relative path to the folder containing manifests in a repository, default value "manifests"
	// +optional
	// +kubebuilder:default:="manifests"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=2
	ContextDir string `json:"contextDir,omitempty"`

	// sourcePath is the subpath within contextDir where kustomize builds start. Examples include any sub-folder or path: `base`, `overlays/dev`, `default`, `odh` etc.
	// +optional
	// +kubebuilder:default:=""
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=3
	SourcePath string `json:"sourcePath,omitempty"`
}
