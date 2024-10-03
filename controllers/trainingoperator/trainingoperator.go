package trainingoperator

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

type TOReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (t *TOReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.KFTO{}).
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

// reduce unnecessary reconcile triggered by kfto's deployment change due to ManagedByODHOperator annotation.
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

func (t *TOReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// Fetch the KFTOComponent instance to know created or deleted
	obj := &dsccomponentv1alpha1.KFTO{}
	err := t.Client.Get(ctx, request.NamespacedName, obj)

	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) || obj.GetDeletionTimestamp() != nil {
			t.Log.Info("TrainingOperator CR has been deletet.", "Request.Name", request.Name)
			if err = t.DeployManifests(ctx, t.Client, t.Log, obj, &obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	t.Log.Info("TrainingOperator CR has been createt.", "Request.Name", request.Name)
	return ctrl.Result{}, t.DeployManifests(ctx, t.Client, t.Log, obj, &obj.Spec.ComponentSpec, true)
}

var (
	ComponentName        = "trainingoperator"
	TrainingOperatorPath = deploy.DefaultManifestPath + "/" + ComponentName + "/rhoai"
)

func (t *TOReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete TrainingOperator Component CR
	toCR := &dsccomponentv1alpha1.TrainingOperator{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DSCTO",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.KFTOComponentSpec{
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
		cli.Create(ctx, toCR)
	} else {
		cli.Delete(ctx, toCR)
	}
	return nil
}

func (r *TOReconciler) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(r.DevFlags.Manifests) != 0 {
		manifestConfig := r.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentName, manifestConfig); err != nil {
			return err
		}
		// If overlay is defined, update paths
		defaultKustomizePath := "rhoai"
		if manifestConfig.SourcePath != "" {
			defaultKustomizePath = manifestConfig.SourcePath
		}
		TrainingOperatorPath = filepath.Join(deploy.DefaultManifestPath, ComponentName, defaultKustomizePath)
	}

	return nil
}

func (r *TOReconciler) GetComponentName() string {
	return ComponentName
}

func (t *TOReconciler) DeployManifests(ctx context.Context, cli client.Client, l logr.Logger,
	owner metav1.Object, componentSpec *dsccomponentv1alpha1.ComponentSpec, _ bool) error {
	var imageParamMap = map[string]string{
		"odh-training-operator-controller-image": "RELATED_IMAGE_ODH_TRAINING_OPERATOR_IMAGE",
	}
	obj := (owner).(*dsccomponentv1alpha1.TrustyAI)
	enabled := obj.GetManagementState() == operatorv1.Managed

	if enabled {
		// TODO
		// if t.DevFlags != nil {
		// 	// Download manifests and update paths
		// 	if err := t.OverrideManifests(ctx, componentSpec.Platform); err != nil {
		// 		return err
		// 	}
		// }
		// if t.DevFlags == nil || len(t.DevFlags.Manifests) == 0 {
		// 	if err := deploy.ApplyParams(TrainingOperatorPath, imageParamMap); err != nil {
		// 		return err
		// 	}
		// }
	}
	// Deploy Training Operator
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, TrainingOperatorPath, componentSpec.DSCISpec.ApplicationsNamespace, ComponentName, enabled); err != nil {
		return err
	}
	l.Info("apply manifests done")

	return nil
}
