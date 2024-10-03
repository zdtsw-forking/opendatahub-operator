package ray

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
)

type RayReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *RayReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.Ray{}).
		Owns(&corev1.Secret{}).
		Owns(
			&appsv1.Deployment{}, builder.WithPredicates(componentDeploymentPredicates)).
		Owns(
			&corev1.Service{}).
		Owns(
			&corev1.ServiceAccount{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}

// reduce unnecessary reconcile triggered by ray's deployment change due to ManagedByODHOperator annotation.
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

func (r *RayReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// Fetch the RayComponent instance to know created or deleted
	obj := &dsccomponentv1alpha1.Ray{}
	err := r.Client.Get(ctx, request.NamespacedName, obj)

	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) || obj.GetDeletionTimestamp() != nil {
			r.Log.Info("Ray CR has been deleter.", "Request.Name", request.Name)
			if err = r.DeployManifests(ctx, r.Client, r.Log, obj, obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	r.Log.Info("Ray CR has been creater.", "Request.Name", request.Name)
	return ctrl.Result{}, r.DeployManifests(ctx, r.Client, r.Log, obj, obj.Spec.ComponentSpec, true)
}

var (
	ComponentName = "ray"
	RayPath       = deploy.DefaultManifestPath + "/" + ComponentName + "/openshift"
)

func (r *RayReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete Ray Component CR
	rayCR := &dsccomponentv1alpha1.Ray{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ray",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.RayComponentSpec{
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
		cli.Create(ctx, rayCR)
	} else {
		cli.Delete(ctx, rayCR)
	}
	return nil
}

func (r *RayReconciler) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(r.DevFlags.Manifests) != 0 {
		manifestConfig := r.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentName, manifestConfig); err != nil {
			return err
		}
		// If overlay is defined, update paths
		defaultKustomizePath := "openshift"
		if manifestConfig.SourcePath != "" {
			defaultKustomizePath = manifestConfig.SourcePath
		}
		RayPath = filepath.Join(deploy.DefaultManifestPath, ComponentName, defaultKustomizePath)
	}

	return nil
}

func (r *RayReconciler) GetComponentName() string {
	return ComponentName
}

func (r *RayReconciler) DeployManifests(ctx context.Context, cli client.Client, l logr.Logger,
	owner metav1.Object, componentSpec *dsccomponentv1alpha1.ComponentSpec, _ bool) error {
	var imageParamMap = map[string]string{
		"odh-kuberay-operator-controller-image": "RELATED_IMAGE_ODH_KUBERAY_OPERATOR_CONTROLLER_IMAGE",
	}
	obj := (owner).(*dsccomponentv1alpha1.Ray)
	enabled := r.GetManagementState() == operatorv1.Managed

	if enabled {
		// if r.DevFlags != nil {
		// 	// Download manifests and update paths
		// 	if err := r.OverrideManifests(ctx, componentSpec.Platform); err != nil {
		// 		return err
		// 	}
		// }
		// if r.DevFlags == nil || len(r.DevFlags.Manifests) == 0 {
		// 	if err := deploy.ApplyParams(RayPath, imageParamMap, map[string]string{"namespace": componentSpec.DSCISpec.ApplicationsNamespace}); err != nil {
		// 		return fmt.Errorf("failed to update image from %s : %w", RayPath, err)
		// 	}
		// }
	}
	// Deploy Ray Operator
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, RayPath, componentSpec.DSCISpec.ApplicationsNamespace, ComponentName, enabled); err != nil {
		return fmt.Errorf("failed to apply manifets from %s : %w", RayPath, err)
	}
	l.Info("apply manifests done")

	return nil
}
