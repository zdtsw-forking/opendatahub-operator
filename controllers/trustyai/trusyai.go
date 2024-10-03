package trustyai

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
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
)

type TrustyAIReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

var (
	ComponentName     = "trustyai"
	ComponentPathName = "trustyai-service-operator"
	PathUpstream      = deploy.DefaultManifestPath + "/" + ComponentPathName + "/overlays/odh"
	PathDownstream    = deploy.DefaultManifestPath + "/" + ComponentPathName + "/overlays/rhoai"
	OverridePath      = ""
)

// SetupWithManager sets up the controller with the Manager.
func (t *TrustyAIReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.TrustyAI{}).
		Owns(&corev1.Secret{}).
		Owns(
			&appsv1.Deployment{}, builder.WithPredicates(componentDeploymentPredicates)).
		Owns(
			&corev1.Service{}).
		Owns(
			&corev1.ServiceAccount{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(t)
}

// reduce unnecessary reconcile triggered by trustyai's deployment change due to ManagedByODHOperator annotation.
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

func (t *TrustyAIReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	obj := &dsccomponentv1alpha1.TrustyAI{}
	err := t.Client.Get(ctx, request.NamespacedName, obj)
	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) || obj.GetDeletionTimestamp() != nil {
			t.Log.Info("TrustyAI CR has been deletet.", "Request.Name", request.Name)
			if err = t.DeployManifests(ctx, t.Client, t.Log, obj, &obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	t.Log.Info("TrustyAI CR has been createt.", "Request.Name", request.Name)
	return ctrl.Result{}, t.DeployManifests(ctx, t.Client, t.Log, obj, &obj.Spec.ComponentSpec, true)
}

func (t *TrustyAIReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete TrustyAI Component CR
	trustyaiCR := &dsccomponentv1alpha1.TrustyAI{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrustyAI",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.TrustyAIComponentSpec{
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
		cli.Create(ctx, trustyaiCR)
	} else {
		cli.Delete(ctx, trustyaiCR)
	}
	return nil
}
func (t *TrustyAIReconciler) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(t.DevFlags.Manifests) != 0 {
		manifestConfig := t.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentPathName, manifestConfig); err != nil {
			return err
		}
		// If overlay is defined, update paths
		defaultKustomizePath := "base"
		if manifestConfig.SourcePath != "" {
			defaultKustomizePath = manifestConfig.SourcePath
		}
		OverridePath = filepath.Join(deploy.DefaultManifestPath, ComponentPathName, defaultKustomizePath)
	}
	return nil
}

func (t *TrustyAIReconciler) GetComponentName() string {
	return ComponentName
}

func (t *TrustyAIReconciler) DeployManifests(ctx context.Context, cli client.Client, l logr.Logger,
	owner metav1.Object, componentSpec *dsccomponentv1alpha1.ComponentSpec, _ bool) error {
	var imageParamMap = map[string]string{
		"trustyaiServiceImage":  "RELATED_IMAGE_ODH_TRUSTYAI_SERVICE_IMAGE",
		"trustyaiOperatorImage": "RELATED_IMAGE_ODH_TRUSTYAI_SERVICE_OPERATOR_IMAGE",
	}
	entryPath := map[cluster.Platform]string{
		cluster.SelfManagedRhods: PathDownstream,
		cluster.ManagedRhods:     PathDownstream,
		cluster.OpenDataHub:      PathUpstream,
		cluster.Unknown:          PathUpstream,
	}[componentSpec.Platform]
	TrustyAI
	enabled := t.GetManagementState() == operatorv1.Managed

	if enabled {
		// TODO
		// if t.DevFlags != nil {
		// 	// Download manifests and update paths
		// 	if err := t.OverrideManifests(ctx, componentSpec.Platform); err != nil {
		// 		return err
		// 	}
		// 	if OverridePath != "" {
		// 		entryPath = OverridePath
		// 	}
		// }
		// if t.DevFlags == nil || len(t.DevFlags.Manifests) == 0 {
		// 	if err := deploy.ApplyParams(entryPath, imageParamMap); err != nil {
		// 		return fmt.Errorf("failed to update image %s: %w", entryPath, err)
		// 	}
		// }
	}
	// Deploy TrustyAI Operator
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, entryPath, componentSpec.DSCISpec.ApplicationsNamespace, t.GetComponentName(), enabled); err != nil {
		return err
	}
	l.Info("apply manifests done")

	return nil
}
