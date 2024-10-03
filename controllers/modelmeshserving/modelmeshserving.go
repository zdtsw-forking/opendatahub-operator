package modelmeshserving

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	operatorv1 "github.com/openshift/api/operator/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
)

type ModelMeshServingReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

var (
	ComponentName          = "model-mesh"
	Path                   = deploy.DefaultManifestPath + "/" + ComponentName + "/overlays/odh"
	DependentComponentName = "odh-model-controller"
	DependentPath          = deploy.DefaultManifestPath + "/" + DependentComponentName + "/base"
)

func (m *ModelMeshServingReconciler) GetManagementState() operatorv1.ManagementState {
	return operatorv1.Managed
}

// SetupWithManager sets up the controller with the Manager.
func (m *ModelMeshServingReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.ModelMeshServing{}).
		Owns(&admissionregistrationv1.MutatingWebhookConfiguration{}).
		Owns(
			&admissionregistrationv1.ValidatingWebhookConfiguration{},
			builder.WithPredicates(modelMeshwebhookPredicates),
		).
		Owns(&corev1.Secret{}).
		Owns(
			&appsv1.Deployment{}, builder.WithPredicates(componentDeploymentPredicates)).
		Owns(
			&corev1.Service{},
			builder.WithPredicates(predicate.Or(predicate.GenerationChangedPredicate{}, modelMeshGeneralPredicates))).
		Owns(
			&corev1.ServiceAccount{}).
		Owns(
			&rbacv1.Role{},
			builder.WithPredicates(predicate.Or(predicate.GenerationChangedPredicate{}, modelMeshRolePredicates))).
		Owns(
			&rbacv1.RoleBinding{},
			builder.WithPredicates(predicate.Or(predicate.GenerationChangedPredicate{}, modelMeshRBPredicates))).
		Owns(
			&rbacv1.ClusterRole{},
			builder.WithPredicates(predicate.Or(predicate.GenerationChangedPredicate{}, modelMeshRolePredicates))).
		Owns(
			&rbacv1.ClusterRoleBinding{},
			builder.WithPredicates(predicate.Or(predicate.GenerationChangedPredicate{}, modelMeshRBPredicates))).
		Owns(
			&networkingv1.NetworkPolicy{},
			builder.WithPredicates(networkpolicyPredicates),
		).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(m)
}

// reduce unnecessary reconcile triggered by modelmesh's deployment change due to ManagedByODHOperator annotation.
var componentDeploymentPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		namespace := e.ObjectNew.GetNamespace()
		if (namespace == "opendatahub" || namespace == "redhat-ods-applications") && e.ObjectNew.GetLabels()[labels.K8SCommon.PartOf] == ComponentName {
			oldManaged, oldExists := e.ObjectOld.GetAnnotations()[annotations.ManagedByODHOperator]
			newManaged := e.ObjectNew.GetAnnotations()[annotations.ManagedByODHOperator]
			// only reoncile if annotation from "not exist" to "set to true", or from "non-true" value to "true"
			if newManaged == "true" && (!oldExists || oldManaged != "true") {
				return true
			}
			return false
		}
		return true
	},
}

// ignore label updates if it is from application namespace.
var modelMeshGeneralPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		if strings.Contains(e.ObjectNew.GetName(), "odh-model-controller") || strings.Contains(e.ObjectNew.GetName(), "kserve") {
			return false
		}
		return true
	},
}
var modelMeshRBPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		notAllowedNames := []string{"leader-election-rolebinding", "proxy-rolebinding", "odh-model-controller-rolebinding-opendatahub"}
		for _, notallowedName := range notAllowedNames {
			if e.ObjectNew.GetName() == notallowedName {
				return false
			}
		}
		return true
	},
}

// a workaround for 2.5 due to odh-model-controller serivceaccount keeps updates with label.
var saPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		namespace := e.ObjectNew.GetNamespace()
		if e.ObjectNew.GetName() == "odh-model-controller" && (namespace == "redhat-ods-applications" || namespace == "opendatahub") {
			return false
		}
		return true
	},
}

// a workaround for 2.5 due to modelmesh-servingruntime.serving.kserve.io keeps updates.
var modelMeshwebhookPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		return e.ObjectNew.GetName() != "modelmesh-servingruntime.serving.kserve.io"
	},
}
var modelMeshRolePredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		notAllowedNames := []string{"leader-election-role", "proxy-role", "metrics-reader", "kserve-prometheus-k8s", "odh-model-controller-role"}
		for _, notallowedName := range notAllowedNames {
			if e.ObjectNew.GetName() == notallowedName {
				return false
			}
		}
		return true
	},
}

// a workaround for modelmesh and kserve both create same odh-model-controller NWP.
var networkpolicyPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		return e.ObjectNew.GetName() != "odh-model-controller"
	},
}

func (m *ModelMeshServingReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// Fetch the ModelMeshServingComponent instance to know created or deleted
	obj := &dsccomponentv1alpha1.ModelMeshServing{}
	err := m.Client.Get(ctx, request.NamespacedName, obj)

	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) || obj.GetDeletionTimestamp() != nil {
			m.Log.Info("ModelMeshServing CR has been deletem.", "Request.Name", request.Name)
			if err = m.DeployManifests(ctx, m.Client, m.Log, obj, obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	m.Log.Info("ModelMeshServing CR has been createm.", "Request.Name", request.Name)
	return ctrl.Result{}, m.DeployManifests(ctx, m.Client, m.Log, obj, obj.Spec.ComponentSpec, true)
}

func (m *ModelMeshServingReconciler) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// Go through each manifest and set the overlays if defined
	for _, subcomponent := range m.DevFlags.Manifests {
		if strings.Contains(subcomponent.URI, DependentComponentName) {
			// Download subcomponent
			if err := deploy.DownloadManifests(ctx, DependentComponentName, subcomponent); err != nil {
				return err
			}
			// If overlay is defined, update paths
			defaultKustomizePath := "base"
			if subcomponent.SourcePath != "" {
				defaultKustomizePath = subcomponent.SourcePath
			}
			DependentPath = filepath.Join(deploy.DefaultManifestPath, DependentComponentName, defaultKustomizePath)
		}

		if strings.Contains(subcomponent.URI, ComponentName) {
			// Download subcomponent
			if err := deploy.DownloadManifests(ctx, ComponentName, subcomponent); err != nil {
				return err
			}
			// If overlay is defined, update paths
			defaultKustomizePath := "overlays/odh"
			if subcomponent.SourcePath != "" {
				defaultKustomizePath = subcomponent.SourcePath
			}
			Path = filepath.Join(deploy.DefaultManifestPath, ComponentName, defaultKustomizePath)
		}
	}
	return nil
}

func (m *ModelMeshServingReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete ModelMeshServing Component CR
	mmCR := &dsccomponentv1alpha1.ModelMeshServing{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ModelMeshServing",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.ModelMeshServingComponentSpec{
			ComponentSpec: dsccomponentv1alpha1.ComponentSpec{
				Platform:      dsci.Status.Release.Name,
				ComponentName: ComponentName,
				DSCISpec:      dsci.Spec,
				DSCComponentSpec: dsccomponentv1alpha1.DSCComponentSpec{
					DSCDevFlags: dsccomponentv1alpha1.DSCDevFlags{
						LoggerMode: "default",
					},
				},
			},
		},
	}
	if enabled {
		cli.Create(ctx, mmCR)
	} else {
		cli.Delete(ctx, mmCR)
	}
	return nil
}
func (m *ModelMeshServingReconciler) GetComponentName() string {
	return ComponentName
}

func (m *ModelMeshServingReconciler) DeployManifests(ctx context.Context, cli client.Client, l logr.Logger,
	owner metav1.Object, componentSpec *dsccomponentv1alpha1.ComponentSpec,
	_ bool,
) error {
	var imageParamMap = map[string]string{
		"odh-mm-rest-proxy":             "RELATED_IMAGE_ODH_MM_REST_PROXY_IMAGE",
		"odh-modelmesh-runtime-adapter": "RELATED_IMAGE_ODH_MODELMESH_RUNTIME_ADAPTER_IMAGE",
		"odh-modelmesh":                 "RELATED_IMAGE_ODH_MODELMESH_IMAGE",
		"odh-modelmesh-controller":      "RELATED_IMAGE_ODH_MODELMESH_CONTROLLER_IMAGE",
		"odh-model-controller":          "RELATED_IMAGE_ODH_MODEL_CONTROLLER_IMAGE",
	}

	// odh-model-controller to use
	var dependentImageParamMap = map[string]string{
		"odh-model-controller": "RELATED_IMAGE_ODH_MODEL_CONTROLLER_IMAGE",
	}

	enabled := m.GetManagementState() == operatorv1.Managed

	// Update Default rolebinding
	if enabled {
		// if m.DevFlags != nil {
		// 	// Download manifests and update paths
		// 	if err := m.OverrideManifests(ctx, componentSpec.Platform); err != nil {
		// 		return err
		// 	}
		// }

		if err := cluster.UpdatePodSecurityRolebinding(ctx, cli, componentSpec.DSCISpec.ApplicationsNamespace,
			"modelmesh",
			"modelmesh-controller",
			"odh-prometheus-operator",
			"prometheus-custom"); err != nil {
			return err
		}
		// Update image parameters
		// if m.DevFlags == nil || len(m.DevFlags.Manifests) == 0 {
		// 	if err := deploy.ApplyParams(Path, imageParamMap); err != nil {
		// 		return fmt.Errorf("failed update image from %s : %w", Path, err)
		// 	}
		// }
	}

	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, Path, componentSpec.DSCISpec.ApplicationsNamespace, ComponentName, enabled); err != nil {
		return fmt.Errorf("failed to apply manifests from %s : %w", Path, err)
	}
	l.WithValues("Path", Path).Info("apply manifests done for modelmesh")
	// For odh-model-controller
	if enabled {
		if err := cluster.UpdatePodSecurityRolebinding(ctx, cli, componentSpec.DSCISpec.ApplicationsNamespace,
			"odh-model-controller"); err != nil {
			return err
		}
		// Update image parameters for odh-model-controller
		if componentSpec.DSCComponentSpec.DSCDevFlags.Manifests == nil {
			if err := deploy.ApplyParams(DependentPath, dependentImageParamMap); err != nil {
				return err
			}
		}
	}
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, DependentPath, componentSpec.DSCISpec.ApplicationsNamespace, m.GetComponentName(), enabled); err != nil {
		// explicitly ignore error if error contains keywords "spec.selector" and "field is immutable" and return all other error.
		if !strings.Contains(err.Error(), "spec.selector") || !strings.Contains(err.Error(), "field is immutable") {
			return err
		}
	}

	l.WithValues("Path", DependentPath).Info("apply manifests done for odh-model-controller")

	if enabled {
		if err := cluster.WaitForDeploymentAvailable(ctx, cli, ComponentName, componentSpec.DSCISpec.ApplicationsNamespace, 20, 2); err != nil {
			return fmt.Errorf("deployment for %s is not ready to server: %w", ComponentName, err)
		}
	}

	return nil
}
