package datasciencepipelines

import (
	"context"

	"github.com/go-logr/logr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	dsccomponentv1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	"github.com/opendatahub-io/opendatahub-operator/v2/components/datasciencepipelines"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
)

type DataSciencePipelineReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (d *DataSciencePipelineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.DataSciencePipeline{}).
		Owns(&corev1.Secret{}).
		Owns(&admissionregistrationv1.MutatingWebhookConfiguration{}).
		Owns(&admissionregistrationv1.ValidatingWebhookConfiguration{}).
		Owns(
			&appsv1.Deployment{}, builder.WithPredicates(componentDeploymentPredicates)).
		Owns(
			&corev1.Service{}).
		Owns(
			&corev1.ServiceAccount{}).
		Watches(
			&apiextensionsv1.CustomResourceDefinition{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
				return d.watchCRD(ctx, a)
			}),
			builder.WithPredicates(argoWorkflowCRDPredicates),
		).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(d)
}

// reduce unnecessary reconcile triggered by DSP's deployment change due to ManagedByODHOperator annotation.
var componentDeploymentPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		namespace := e.ObjectNew.GetNamespace()
		if (namespace == "opendatahub" || namespace == "redhat-ods-applications") && e.ObjectNew.GetLabels()[labels.K8SCommon.PartOf] == datasciencepipelines.ComponentName {
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

func (d *DataSciencePipelineReconciler) watchCRD(ctx context.Context, a client.Object) []reconcile.Request {
	if a.GetName() == "ArgoWorkflowCRD" {
		return []reconcile.Request{{
			NamespacedName: types.NamespacedName{Name: "ArgoWorkflowCRD"},
		}}
	}
	return nil
}

// argoWorkflowCRDPredicates filters the delete events to trigger reconcile when Argo Workflow CRD is deleted.
var argoWorkflowCRDPredicates = predicate.Funcs{
	DeleteFunc: func(e event.DeleteEvent) bool {
		if e.Object.GetName() == datasciencepipelines.ArgoWorkflowCRD {
			labelList := e.Object.GetLabels()
			// CRD to be deleted with label "app.opendatahub.io/datasciencepipeline":"true", should not trigger reconcile
			if value, exist := labelList[labels.ODH.Component(datasciencepipelines.ComponentName)]; exist && value == "true" {
				return false
			}
		}
		// CRD to be deleted either not with label or label value is not "true", should trigger reconcile
		return true
	},
}

func (d *DataSciencePipelineReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// Fetch the DSPComponent instance to know created or deleted
	obj := &dsccomponentv1alpha1.DataSciencePipeline{}
	err := d.Client.Get(ctx, request.NamespacedName, obj)
	if obj.GetName() != request.Name && obj.GetOwnerReferences()[0].Name != "default-dsc" {
		return ctrl.Result{}, nil
	}
	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) || obj.GetDeletionTimestamp() != nil {
			d.Log.Info("DataSciencePipeline CR has been deleted.", "Request.Name", request.Name)
			if err = ReconcileComponent(ctx, d.Client, d.Log, obj, obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	d.Log.Info("DataSciencePipeline CR has been created/updated.", "Request.Name", request.Name)
	ReconcileComponent(ctx, d.Client, d.Log, obj, obj.Spec.ComponentSpec, true)
}
