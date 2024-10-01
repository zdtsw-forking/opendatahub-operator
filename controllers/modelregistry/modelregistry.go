package modelregistry

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/components/modelregistry"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ModelRegistryReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (m *ModelRegistryReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.ModelReg{}).
		Owns(&admissionregistrationv1.MutatingWebhookConfiguration{}).
		Owns(&admissionregistrationv1.ValidatingWebhookConfiguration{}).
		Owns(&corev1.Secret{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
				return m.watchDefaultIngressSecret(ctx, a)
			}),
			builder.WithPredicates(defaultIngressCertSecretPredicates)).
		Owns(
			&appsv1.Deployment{}, builder.WithPredicates(componentDeploymentPredicates)).
		Owns(
			&corev1.Service{}).
		Owns(
			&corev1.ServiceAccount{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(m)
}

func (r *ModelRegistryReconciler) watchDefaultIngressSecret(ctx context.Context, a client.Object) []reconcile.Request {
	requestName, err := r.getRequestName(ctx)
	if err != nil {
		return []reconcile.Request{}
	}
	// When ingress secret gets created/deleted, trigger reconcile function
	ingressCtrl, err := cluster.FindAvailableIngressController(ctx, r.Client)
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
		if (namespace == "opendatahub" || namespace == "redhat-ods-applications") && e.ObjectNew.GetLabels()[labels.K8SCommon.PartOf] == modelregistry.ComponentName {
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

func (m *ModelRegistryReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// Fetch the ModelRegComponent instance to know created or deleted
	obj := &dsccomponentv1alpha1.ModelReg{}
	err := m.Client.Get(ctx, request.NamespacedName, obj)
	if obj.GetName() != request.Name && obj.GetOwnerReferences()[0].Name != "default-dsc" {
		return ctrl.Result{}, nil
	}
	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) || obj.GetDeletionTimestamp() != nil {
			m.Log.Info("ModelRegistry CR has been deletem.", "Request.Name", request.Name)
			if err = DeployManifests(ctx, m.Client, m.Log, obj, obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	m.Log.Info("ModelRegistry CR has been createm.", "Request.Name", request.Name)
	DeployManifests(ctx, m.Client, m.Log, obj, obj.Spec.ComponentSpec, true)
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

const DefaultModelRegistryCert = "default-modelregistry-cert"

var (
	ComponentName                   = "model-registry-operator"
	DefaultModelRegistriesNamespace = "odh-model-registries"
	Path                            = deploy.DefaultManifestPath + "/" + ComponentName + "/overlays/odh"
	// we should not apply this label to the namespace, as it triggered namspace deletion during operator uninstall
	// modelRegistryLabels = cluster.WithLabels(
	//      labels.ODH.OwnedNamespace, "true",
	// ).
)

// ModelRegistry struct holds the configuration for the ModelRegistry component.
// The property `registriesNamespace` is immutable when `managementState` is `Managed`

// +kubebuilder:object:generate=true
// +kubebuilder:validation:XValidation:rule="(self.managementState != 'Managed') || (oldSelf.registriesNamespace == ”) || (oldSelf.managementState != 'Managed')|| (self.registriesNamespace == oldSelf.registriesNamespace)",message="RegistriesNamespace is immutable when model registry is Managed"
//
//nolint:lll

func (m *ModelRegistryReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete Ray Component CR
	rayCR := &dsccomponentv1alpha1.ModelReg{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ModelReg",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default-modelreg",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.ModelRegComponentSpec{
			ComponentSpec: dsccomponentv1alpha1.ComponentSpec{
				Platform:              dsci.Status.Release.Name,
				ComponentName:         ComponentName,
				DSCInitializationSpec: dsci.Spec,
				ComponentDevFlags: dsccomponentv1alpha1.DevFlags{
					LoggerMode: dsci.Spec.DevFlags.LogMode,
				},
			},
		},
	}
	if enabled {
		cli.Create(ctx, rayCR)
	} else {
		cli.Delete(ctx, rayCR)
	}
	return nil
}

func (m *ModelRegistryReconciler) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(m.DevFlags.Manifests) != 0 {
		manifestConfig := m.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentName, manifestConfig); err != nil {
			return err
		}
		// If overlay is defined, update paths
		defaultKustomizePath := "overlays/odh"
		if manifestConfig.SourcePath != "" {
			defaultKustomizePath = manifestConfig.SourcePath
		}
		Path = filepath.Join(deploy.DefaultManifestPath, ComponentName, defaultKustomizePath)
	}

	return nil
}

func (m *ModelRegistryReconciler) GetComponentName() string {
	return ComponentName
}

func (m *ModelRegistryReconciler) DeployManifests(ctx context.Context, cli client.Client, l logr.Logger,
	owner metav1.Object, componentSpec *dsccomponentv1alpha1.ComponentSpec, _ bool) error {
	var imageParamMap = map[string]string{
		"IMAGES_MODELREGISTRY_OPERATOR": "RELATED_IMAGE_ODH_MODEL_REGISTRY_OPERATOR_IMAGE",
		"IMAGES_GRPC_SERVICE":           "RELATED_IMAGE_ODH_MLMD_GRPC_SERVER_IMAGE",
		"IMAGES_REST_SERVICE":           "RELATED_IMAGE_ODH_MODEL_REGISTRY_IMAGE",
	}
	enabled := m.GetManagementState() == operatorv1.Managed
	monitoringEnabled := componentSpec.Monitoring.ManagementState == operatorv1.Managed

	if enabled {
		// return error if ServiceMesh is not enabled, as it's a required feature
		if componentSpec.ServiceMesh == nil || componentSpec.ServiceMesh.ManagementState != operatorv1.Managed {
			return errors.New("ServiceMesh needs to be set to 'Managed' in DSCI CR, it is required by Model Registry")
		}

		if err := m.createDependencies(ctx, cli, componentSpec); err != nil {
			return err
		}

		if m.DevFlags != nil {
			// Download manifests and update paths
			if err := m.OverrideManifests(ctx, componentSpec.Platform); err != nil {
				return err
			}
		}

		// Update image parameters only when we do not have customized manifests set
		if m.DevFlags == nil || len(m.DevFlags.Manifests) == 0 {
			extraParamsMap := map[string]string{
				"DEFAULT_CERT": DefaultModelRegistryCert,
			}
			if err := deploy.ApplyParams(Path, imageParamMap, extraParamsMap); err != nil {
				return fmt.Errorf("failed to update image from %s : %w", Path, err)
			}
		}

		// Create model registries namespace
		// We do not delete this namespace even when ModelRegistry is Removed or when operator is uninstalled.
		ns, err := cluster.CreateNamespace(ctx, cli, m.RegistriesNamespace)
		if err != nil {
			return err
		}
		l.Info("created model registry namespace", "namespace", m.RegistriesNamespace)
		// create servicemeshmember here, for now until post MVP solution
		err = enrollToServiceMesh(ctx, cli, componentSpec, ns)
		if err != nil {
			return err
		}
		l.Info("created model registry servicemesh member", "namespace", m.RegistriesNamespace)
	} else {
		err := m.removeDependencies(ctx, cli, componentSpec)
		if err != nil {
			return err
		}
	}

	// Deploy ModelRegistry Operator
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, Path, componentSpec.DSCISpec.ApplicationsNamespace, m.GetComponentName(), enabled); err != nil {
		return err
	}
	l.Info("apply manifests done")

	// Create additional model registry resources, componentEnabled=true because these extras are never deleted!
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, Path+"/extras", componentSpec.DSCISpec.ApplicationsNamespace, m.GetComponentName(), true); err != nil {
		return err
	}
	l.Info("apply extra manifests done")

	if enabled {
		if err := cluster.WaitForDeploymentAvailable(ctx, cli, m.GetComponentName(), componentSpec.DSCISpec.ApplicationsNamespace, 10, 1); err != nil {
			return fmt.Errorf("deployment for %s is not ready to server: %w", ComponentName, err)
		}
	}

	// CloudService Monitoring handling
	if componentSpec.Platform == cluster.ManagedRhods {
		if err := m.UpdatePrometheusConfig(cli, l, enabled && monitoringEnabled, ComponentName); err != nil {
			return err
		}
		if err := deploy.DeployManifestsFromPath(ctx, cli, owner,
			filepath.Join(deploy.DefaultManifestPath, "monitoring", "prometheus", "apps"),
			componentSpec.DSCISpec.Monitoring.Namespace,
			"prometheus", true); err != nil {
			return err
		}
		l.Info("updating SRE monitoring done")
	}
	return nil
}
