package kserve

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
)

type KserveReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:validation:Pattern=`^(Serverless|RawDeployment)$`
type DefaultDeploymentMode string

var (
	ComponentName          = "kserve"
	Path                   = deploy.DefaultManifestPath + "/" + ComponentName + "/overlays/odh"
	DependentComponentName = "odh-model-controller"
	DependentPath          = deploy.DefaultManifestPath + "/" + DependentComponentName + "/base"
	ServiceMeshOperator    = "servicemeshoperator"
	ServerlessOperator     = "serverless-operator"
	// Serverless will be used as the default deployment mode for Kserve. This requires Serverless and ServiceMesh operators configured as dependencies.
	Serverless DefaultDeploymentMode = "Serverless"
	// RawDeployment will be used as the default deployment mode for Kserve.
	RawDeployment DefaultDeploymentMode = "RawDeployment"
)

func (k *KserveReconciler) Cleanup(ctx context.Context, cli client.Client, owner metav1.Object, instance *dsccomponentv1alpha1.ComponentSpec) error {
	if removeServerlessErr := k.removeServerlessFeatures(ctx, cli, owner, instance); removeServerlessErr != nil {
		return removeServerlessErr
	}

	return k.removeServiceMeshConfigurations(ctx, cli, owner, instance)
}

// SetupWithManager sets up the controller with the Manager.
func (k *KserveReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.Kserve{}).
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
		Complete(k)
}

func (k *KserveReconciler) watchDefaultIngressSecret(ctx context.Context, a client.Object) []reconcile.Request {
	requestName := k.GetComponentName()

	// When ingress secret gets created/deleted, trigger reconcile function
	ingressCtrl, err := cluster.FindAvailableIngressController(ctx, k.Client)
	if err != nil {
		return []reconcile.Request{}
	}
	defaultIngressSecretName := cluster.GetDefaultIngressCertSecretName(ingressCtrl)
	if a.GetName() == defaultIngressSecretName && a.GetNamespace() == "openshift-ingress" {
		return []reconcile.Request{{
			NamespacedName: types.NamespacedName{Name: requestName},
		}}
	}
	return []reconcile.Request{}
}

// reduce unnecessary reconcile triggered by ModelReg's deployment change due to ManagedByODHOperator annotation.
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

var configMapPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		// Do not reconcile on kserver's inferenceservice-config CM updates, for rawdeployment
		namespace := e.ObjectNew.GetNamespace()
		if e.ObjectNew.GetName() == "inferenceservice-config" && (namespace == "redhat-ods-applications" || namespace == "opendatahub") { //nolint:goconst
			return false
		}
		return true
	},
}

func (k *KserveReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// Fetch the KserveComponent instance to know created or deleted
	obj := &dsccomponentv1alpha1.Kserve{}
	err := k.Client.Get(ctx, request.NamespacedName, obj)

	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) || obj.GetDeletionTimestamp() != nil {
			k.Log.Info("Kserve CR has been deletek.", "Request.Name", request.Name)
			if err = k.DeployManifests(ctx, k.Client, k.Log, obj, &obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	k.Log.Info("Kserve CR has been createk.", "Request.Name", request.Name)
	return ctrl.Result{}, k.DeployManifests(ctx, k.Client, k.Log, obj, &obj.Spec.ComponentSpec, true)
}

// defaultIngressCertSecretPredicates filters delete and create events to trigger reconcile when default ingress cert secret is expired
// or created.
var defaultIngressCertSecretPredicates = predicate.Funcs{
	CreateFunc: func(createEvent event.CreateEvent) bool {
		return true

	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		return true
	},
}

func (k *KserveReconciler) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// Download manifests if defined by devflags
	// Go through each manifest and set the overlays if defined
	for _, subcomponent := range k.DevFlags.Manifests {
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

func (k *KserveReconciler) GetComponentName() string {
	return ComponentName
}

func (k *KserveReconciler) DeployManifests(ctx context.Context, cli client.Client,
	l logr.Logger, owner metav1.Object, componentSpec *dsccomponentv1alpha1.ComponentSpec, _ bool) error {
	// dependentParamMap for odh-model-controller to use.
	var dependentParamMap = map[string]string{
		"odh-model-controller": "RELATED_IMAGE_ODH_MODEL_CONTROLLER_IMAGE",
	}

	enabled := k.GetManagementState() == operatorv1.Managed
	obj := (owner).(*dsccomponentv1alpha1.Kserve)
	if !enabled {
		if err := k.removeServerlessFeatures(ctx, cli, owner, componentSpec); err != nil {
			return err
		}
	} else {
		// Configure dependencies
		if err := k.configureServerless(ctx, cli, l, owner, componentSpec); err != nil {
			return err
		}
		// if k.DevFlags != nil {
		// 	// Download manifests and update paths
		// 	if err := k.OverrideManifests(ctx, componentSpec.Platform); err != nil {
		// 		return err
		// 	}
		// }
	}

	if err := k.configureServiceMesh(ctx, cli, owner, componentSpec); err != nil {
		return fmt.Errorf("failed configuring service mesh while reconciling kserve component. cause: %w", err)
	}

	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, Path, componentSpec.DSCISpec.ApplicationsNamespace, ComponentName, enabled); err != nil {
		return fmt.Errorf("failed to apply manifests from %s : %w", Path, err)
	}

	l.WithValues("Path", Path).Info("apply manifests done for kserve")

	if enabled {
		if err := k.setupKserveConfig(ctx, cli, l, componentSpec); err != nil {
			return err
		}

		// For odh-model-controller
		if err := cluster.UpdatePodSecurityRolebinding(ctx, cli, componentSpec.DSCISpec.ApplicationsNamespace, "odh-model-controller"); err != nil {
			return err
		}
		// Update image parameters for odh-model-controller
		// if k.DevFlags == nil || len(k.DevFlags.Manifests) == 0 {
		// 	if err := deploy.ApplyParams(DependentPath, dependentParamMap); err != nil {
		// 		return fmt.Errorf("failed to update image %s: %w", DependentPath, err)
		// 	}
		// }
	}

	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, DependentPath, componentSpec.DSCISpec.ApplicationsNamespace, ComponentName, enabled); err != nil {
		if !strings.Contains(err.Error(), "spec.selector") || !strings.Contains(err.Error(), "field is immutable") {
			// explicitly ignore error if error contains keywords "spec.selector" and "field is immutable" and return all other error.
			return err
		}
	}
	l.WithValues("Path", Path).Info("apply manifests done for odh-model-controller")

	return nil
}

func (k *KserveReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete Kserve Component CR
	kserveCR := &dsccomponentv1alpha1.Kserve{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Kserve",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.KserveComponentSpec{
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
		cli.Create(ctx, kserveCR)
	} else {
		cli.Delete(ctx, kserveCR)
	}
	return nil
}
