package kueue

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/components/kueue"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KueueReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (k *KueueReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.Kueue{}).
		Owns(&corev1.Secret{}).
		Owns(
			&appsv1.Deployment{}, builder.WithPredicates(componentDeploymentPredicates)).
		Owns(
			&corev1.Service{}).
		Owns(
			&corev1.ServiceAccount{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(k)
}

// reduce unnecessary reconcile triggered by kueue's deployment change due to ManagedByODHOperator annotation.
var componentDeploymentPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		namespace := e.ObjectNew.GetNamespace()
		if (namespace == "opendatahub" || namespace == "redhat-ods-applications") && e.ObjectNew.GetLabels()[labels.K8SCommon.PartOf] == kueue.ComponentName {
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

func (k *KueueReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// Fetch the KueueComponent instance to know created or deleted
	obj := &dsccomponentv1alpha1.Kueue{}
	err := k.Client.Get(ctx, request.NamespacedName, obj)
	if obj.GetName() != request.Name && obj.GetOwnerReferences()[0].Name != "default-dsc" {
		return ctrl.Result{}, nil
	}
	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) || obj.GetDeletionTimestamp() != nil {
			k.Log.Info("Kueue CR has been deletek.", "Request.Name", request.Name)
			if err = k.DeployManifests(ctx, k.Client, k.Log, obj, obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	k.Log.Info("Kueue CR has been createk.", "Request.Name", request.Name)
	k.DeployManifests(ctx, k.Client, k.Log, obj, obj.Spec.ComponentSpec, true)
}

var (
	ComponentName = "kueue"
	Path          = deploy.DefaultManifestPath + "/" + ComponentName + "/rhoai" // same path for both odh and rhoai
)

func (k *KueueReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete Kueue Component CR
	kueueCR := &dsccomponentv1alpha1.Kueue{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Kueue",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default-kueue",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.KueueComponentSpec{
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
		cli.Create(ctx, kueueCR)
	} else {
		cli.Delete(ctx, kueueCR)
	}
	return nil
}
func (k *KueueReconciler) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(k.DevFlags.Manifests) != 0 {
		manifestConfig := k.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentName, manifestConfig); err != nil {
			return err
		}
		// If overlay is defined, update paths
		defaultKustomizePath := "rhoai"
		if manifestConfig.SourcePath != "" {
			defaultKustomizePath = manifestConfig.SourcePath
		}
		Path = filepath.Join(deploy.DefaultManifestPath, ComponentName, defaultKustomizePath)
	}

	return nil
}

func (k *KueueReconciler) GetComponentName() string {
	return ComponentName
}

func (k *KueueReconciler) DeployManifests(ctx context.Context, cli client.Client, l logr.Logger,
	owner metav1.Object, componentSpec *dsccomponentv1alpha1.ComponentSpec, _ bool) error {
	var imageParamMap = map[string]string{
		"odh-kueue-controller-image": "RELATED_IMAGE_ODH_KUEUE_CONTROLLER_IMAGE", // new kueue image
	}

	enabled := k.GetManagementState() == operatorv1.Managed
	monitoringEnabled := componentSpec.DSCISpec.Monitoring.ManagementState == operatorv1.Managed
	if enabled {
		if k.DevFlags != nil {
			// Download manifests and update paths
			if err := k.OverrideManifests(ctx, componentSpec.Platform); err != nil {
				return err
			}
		}
		if k.DevFlags == nil || len(k.DevFlags.Manifests) == 0 {
			if err := deploy.ApplyParams(Path, imageParamMap); err != nil {
				return fmt.Errorf("failed to update image from %s : %w", Path, err)
			}
		}
	}
	// Deploy Kueue Operator
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, Path, componentSpec.DSCISpec.ApplicationsNamespace, ComponentName, enabled); err != nil {
		return fmt.Errorf("failed to apply manifetss %s: %w", Path, err)
	}
	l.Info("apply manifests done")

	if enabled {
		if err := cluster.WaitForDeploymentAvailable(ctx, cli, ComponentName, componentSpec.DSCISpec.ApplicationsNamespace, 20, 2); err != nil {
			return fmt.Errorf("deployment for %s is not ready to server: %w", ComponentName, err)
		}
	}

	// CloudService Monitoring handling
	if componentSpec.Platform == cluster.ManagedRhods {
		if err := k.UpdatePrometheusConfig(cli, l, enabled && monitoringEnabled, ComponentName); err != nil {
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
