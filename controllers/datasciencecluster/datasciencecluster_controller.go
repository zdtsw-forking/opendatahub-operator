/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package datasciencecluster contains controller logic of CRD DataScienceCluster
package datasciencecluster

import (
	"context"
	"errors"
	"fmt"
	"time"

	operatorv1 "github.com/openshift/api/operator/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	componentApi "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	dscv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/datasciencecluster/v1"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	datasciencepipelinesctrl "github.com/opendatahub-io/opendatahub-operator/v2/controllers/components/datasciencepipelines"
	"github.com/opendatahub-io/opendatahub-operator/v2/controllers/status"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	cr "github.com/opendatahub-io/opendatahub-operator/v2/pkg/componentsregistry"
	odhClient "github.com/opendatahub-io/opendatahub-operator/v2/pkg/controller/client"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/upgrade"
)

// DataScienceClusterReconciler reconciles a DataScienceCluster object.
type DataScienceClusterReconciler struct {
	*odhClient.Client
	Scheme *runtime.Scheme
	// Recorder to generate events
	Recorder           record.EventRecorder
	DataScienceCluster *DataScienceClusterConfig
}

// DataScienceClusterConfig passing Spec of DSCI for reconcile DataScienceCluster.
type DataScienceClusterConfig struct {
	DSCISpec *dsciv1.DSCInitializationSpec
}

const (
	finalizerName = "datasciencecluster.opendatahub.io/finalizer"
)

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DataScienceClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("DataScienceCluster")
	log.Info("Reconciling DataScienceCluster resources", "Request.Name", req.Name)

	// Get information on version and platform
	currentOperatorRelease := cluster.GetRelease()
	// Set platform
	platform := currentOperatorRelease.Name

	instances := &dscv1.DataScienceClusterList{}

	if err := r.Client.List(ctx, instances); err != nil {
		return ctrl.Result{}, err
	}

	if len(instances.Items) == 0 {
		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected.
		// For additional cleanup logic use operatorUninstall function.
		// Return and don't requeue
		if upgrade.HasDeleteConfigMap(ctx, r.Client) {
			if uninstallErr := upgrade.OperatorUninstall(ctx, r.Client, platform); uninstallErr != nil {
				return ctrl.Result{}, fmt.Errorf("error while operator uninstall: %w", uninstallErr)
			}
		}

		return ctrl.Result{}, nil
	}

	instance := &instances.Items[0]

	// allComponents, err := instance.GetComponents()
	// if err != nil {
	//	return ctrl.Result{}, err
	// }

	// If DSC CR exist and deletion CM exist
	// delete DSC CR and let reconcile requeue
	// sometimes with finalizer DSC CR won't get deleted, force to remove finalizer here
	if upgrade.HasDeleteConfigMap(ctx, r.Client) {
		if controllerutil.ContainsFinalizer(instance, finalizerName) {
			if controllerutil.RemoveFinalizer(instance, finalizerName) {
				if err := r.Update(ctx, instance); err != nil {
					log.Info("Error to remove DSC finalizer", "error", err)
					return ctrl.Result{}, err
				}
				log.Info("Removed finalizer for DataScienceCluster", "name", instance.Name, "finalizer", finalizerName)
			}
		}
		if err := r.Client.Delete(ctx, instance, []client.DeleteOption{}...); err != nil {
			if !k8serr.IsNotFound(err) {
				return reconcile.Result{}, err
			}
		}
		// for _, component := range allComponents {
		//	if err := component.Cleanup(ctx, r.Client, instance, r.DataScienceCluster.DSCISpec); err != nil {
		//		return ctrl.Result{}, err
		//	}
		// }
		return reconcile.Result{Requeue: true}, nil
	}

	// Verify a valid DSCInitialization instance is created
	dsciInstances := &dsciv1.DSCInitializationList{}
	err := r.Client.List(ctx, dsciInstances)
	if err != nil {
		log.Error(err, "Failed to retrieve DSCInitialization resource.", "DSCInitialization Request.Name", req.Name)
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, "DSCInitializationReconcileError", "Failed to retrieve DSCInitialization instance")

		return ctrl.Result{}, err
	}

	// Update phase to error state if DataScienceCluster is created without valid DSCInitialization
	switch len(dsciInstances.Items) { // only handle number as 0 or 1, others won't be existed since webhook block creation
	case 0:
		reason := status.ReconcileFailed
		message := "Failed to get a valid DSCInitialization instance, please create a DSCI instance"
		log.Info(message)
		instance, err = status.UpdateWithRetry(ctx, r.Client, instance, func(saved *dscv1.DataScienceCluster) {
			status.SetProgressingCondition(&saved.Status.Conditions, reason, message)
			// Patch Degraded with True status
			status.SetCondition(&saved.Status.Conditions, "Degraded", reason, message, corev1.ConditionTrue)
			saved.Status.Phase = status.PhaseError
		})
		if err != nil {
			r.reportError(ctx, err, instance, "failed to update DataScienceCluster condition")

			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case 1:
		dscInitializationSpec := dsciInstances.Items[0].Spec
		dscInitializationSpec.DeepCopyInto(r.DataScienceCluster.DSCISpec)
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(instance, finalizerName) {
			log.Info("Adding finalizer for DataScienceCluster", "name", instance.Name, "finalizer", finalizerName)
			controllerutil.AddFinalizer(instance, finalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		log.Info("Finalization DataScienceCluster start deleting instance", "name", instance.Name, "finalizer", finalizerName)
		// for _, component := range allComponents {
		//	if err := component.Cleanup(ctx, r.Client, instance, r.DataScienceCluster.DSCISpec); err != nil {
		//		return ctrl.Result{}, err
		//	}
		// }
		if controllerutil.ContainsFinalizer(instance, finalizerName) {
			controllerutil.RemoveFinalizer(instance, finalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		if upgrade.HasDeleteConfigMap(ctx, r.Client) {
			// if delete configmap exists, requeue the request to handle operator uninstall
			return reconcile.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, nil
	}

	// Start reconciling
	if instance.Status.Conditions == nil {
		reason := status.ReconcileInit
		message := "Initializing DataScienceCluster resource"
		instance, err = status.UpdateWithRetry(ctx, r.Client, instance, func(saved *dscv1.DataScienceCluster) {
			status.SetProgressingCondition(&saved.Status.Conditions, reason, message)
			saved.Status.Phase = status.PhaseProgressing
			saved.Status.Release = currentOperatorRelease
		})
		if err != nil {
			_ = r.reportError(ctx, err, instance, fmt.Sprintf("failed to add conditions to status of DataScienceCluster resource name %s", req.Name))

			return ctrl.Result{}, err
		}
	}

	// all DSC defined components
	componentErrors := cr.ForEach(func(component cr.ComponentHandler) error {
		var err error
		instance, err = r.ReconcileComponent(ctx, instance, component)
		return err
	})

	// Process errors for components
	if componentErrors != nil {
		log.Info("DataScienceCluster Deployment Incomplete.")
		instance, err = status.UpdateWithRetry(ctx, r.Client, instance, func(saved *dscv1.DataScienceCluster) {
			status.SetCompleteCondition(&saved.Status.Conditions, status.ReconcileCompletedWithComponentErrors,
				fmt.Sprintf("DataScienceCluster resource reconciled with component errors: %v", componentErrors))
			saved.Status.Phase = status.PhaseReady
			saved.Status.Release = currentOperatorRelease
		})
		if err != nil {
			log.Error(err, "failed to update DataScienceCluster conditions with incompleted reconciliation")

			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(instance, corev1.EventTypeNormal, "DataScienceClusterComponentFailures",
			"DataScienceCluster instance %s created, but have some failures in component %v", instance.Name, componentErrors)

		return ctrl.Result{RequeueAfter: time.Second * 30}, componentErrors
	}

	// finalize reconciliation
	instance, err = status.UpdateWithRetry(ctx, r.Client, instance, func(saved *dscv1.DataScienceCluster) {
		status.SetCompleteCondition(&saved.Status.Conditions, status.ReconcileCompleted, "DataScienceCluster resource reconciled successfully")
		saved.Status.Phase = status.PhaseReady
		saved.Status.Release = currentOperatorRelease
	})

	if err != nil {
		log.Error(err, "failed to update DataScienceCluster conditions after successfully completed reconciliation")

		return ctrl.Result{}, err
	}

	log.Info("DataScienceCluster Deployment Completed.")
	r.Recorder.Eventf(instance, corev1.EventTypeNormal, "DataScienceClusterCreationSuccessful",
		"DataScienceCluster instance %s created and deployed successfully", instance.Name)

	return ctrl.Result{}, nil
}

func (r *DataScienceClusterReconciler) ReconcileComponent(
	ctx context.Context,
	instance *dscv1.DataScienceCluster,
	component cr.ComponentHandler,
) (*dscv1.DataScienceCluster, error) {
	log := logf.FromContext(ctx)
	componentName := component.GetName()

	log.Info("Starting reconciliation of component: " + componentName)

	enabled := component.GetManagementState(instance) == operatorv1.Managed

	componentCR := component.NewCRObject(instance)
	err := r.apply(ctx, instance, componentCR)
	if err != nil {
		log.Error(err, "Failed to reconciled component CR: "+componentName)
		instance = r.reportError(ctx, err, instance, fmt.Sprintf("failed to reconciled %s by DSC", componentName))
		instance, _ = status.UpdateWithRetry(ctx, r.Client, instance, func(saved *dscv1.DataScienceCluster) {
			status.SetComponentCondition(&saved.Status.Conditions, componentName, status.ReconcileFailed, fmt.Sprintf("Component reconciliation failed: %v", err), corev1.ConditionFalse)
		})
		return instance, err
	}

	// TODO: check component status before update DSC status to successful .GetStatus().Phase == "Ready"
	log.Info("component reconciled successfully: " + componentName)
	instance, err = status.UpdateWithRetry(ctx, r.Client, instance, func(saved *dscv1.DataScienceCluster) {
		if saved.Status.InstalledComponents == nil {
			saved.Status.InstalledComponents = make(map[string]bool)
		}
		// only set non-modelcontroller component into DSC .status.InstalledComponents map
		if componentName != componentApi.ModelControllerComponentName {
			saved.Status.InstalledComponents[componentName] = enabled
		}
		if enabled {
			status.SetComponentCondition(&saved.Status.Conditions, componentName, status.ReconcileCompleted, "Component reconciled successfully", corev1.ConditionTrue)
		} else {
			status.RemoveComponentCondition(&saved.Status.Conditions, componentName)
		}
	})
	if err != nil {
		return instance, fmt.Errorf("failed to update DataScienceCluster status after reconciling %s: %w", componentName, err)
	}

	// TODO: report failed component status .GetStatus().Phase == "NotReady" not only creation
	if err != nil {
		log.Error(err, "Failed to reconcile component: "+componentName)
		instance = r.reportError(ctx, err, instance, fmt.Sprintf("failed to reconcile %s", componentName))
		instance, _ = status.UpdateWithRetry(ctx, r.Client, instance, func(saved *dscv1.DataScienceCluster) {
			status.SetComponentCondition(&saved.Status.Conditions, componentName, status.ReconcileFailed, fmt.Sprintf("Component reconciliation failed: %v", err), corev1.ConditionFalse)
		})
		return instance, err
	}
	return instance, nil
}

func (r *DataScienceClusterReconciler) reportError(ctx context.Context, err error, instance *dscv1.DataScienceCluster, message string) *dscv1.DataScienceCluster {
	logf.FromContext(ctx).Error(err, message, "instance.Name", instance.Name)
	r.Recorder.Eventf(instance, corev1.EventTypeWarning, "DataScienceClusterReconcileError",
		"%s for instance %s", message, instance.Name)
	return instance
}

var configMapPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		// Do not reconcile on prometheus configmap update, since it is handled by DSCI
		if e.ObjectNew.GetName() == "prometheus" && e.ObjectNew.GetNamespace() == "redhat-ods-monitoring" {
			return false
		}
		// TODO: cleanup below
		// Do not reconcile on kserver's inferenceservice-config CM updates, for rawdeployment
		namespace := e.ObjectNew.GetNamespace()
		if e.ObjectNew.GetName() == "inferenceservice-config" && (namespace == "redhat-ods-applications" || namespace == "opendatahub") {
			return false
		}
		return true
	},
}

// apply either create component CR with owner set or delete component CR if it is marked with annotation.
func (r *DataScienceClusterReconciler) apply(ctx context.Context, dsc *dscv1.DataScienceCluster, obj client.Object) error {
	if obj.GetObjectKind().GroupVersionKind().Empty() {
		return errors.New("no groupversionkind defined")
	}
	if err := ctrl.SetControllerReference(dsc, obj, r.Scheme); err != nil {
		return err
	}

	managementStateAnn, exists := obj.GetAnnotations()[annotations.ManagementStateAnnotation]
	if exists && managementStateAnn == string(operatorv1.Removed) {
		err := r.Client.Delete(ctx, obj)
		if k8serr.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.Client.Apply(ctx, obj, client.FieldOwner(dsc.Name), client.ForceOwnership); err != nil {
		return client.IgnoreNotFound(err)
	}

	return nil
}

// reduce unnecessary reconcile triggered by odh component's deployment change due to ManagedByODHOperator annotation.
var componentDeploymentPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		namespace := e.ObjectNew.GetNamespace()
		if namespace == "opendatahub" || namespace == "redhat-ods-applications" {
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

// SetupWithManager sets up the controller with the Manager.
func (r *DataScienceClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dscv1.DataScienceCluster{}).
		Owns(
			&corev1.ConfigMap{},
			builder.WithPredicates(configMapPredicates),
		).
		Owns(
			&appsv1.Deployment{},
			builder.WithPredicates(componentDeploymentPredicates)).
		Owns(&admissionregistrationv1.MutatingWebhookConfiguration{}).
		Owns(&admissionregistrationv1.ValidatingWebhookConfiguration{}).
		Owns(&corev1.ServiceAccount{}).
		// components CRs
		Owns(&componentApi.Dashboard{}).
		Owns(&componentApi.Workbenches{}).
		Owns(&componentApi.Ray{}).
		Owns(&componentApi.ModelRegistry{}).
		Owns(&componentApi.TrustyAI{}).
		Owns(&componentApi.Kueue{}).
		Owns(&componentApi.CodeFlare{}).
		Owns(&componentApi.TrainingOperator{}).
		Owns(&componentApi.DataSciencePipelines{}).
		Owns(&componentApi.Kserve{}).
		Owns(&componentApi.ModelMeshServing{}).
		Owns(&componentApi.ModelController{}).
		Watches(
			&dsciv1.DSCInitialization{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
				return r.watchDataScienceClusterForDSCI(ctx, a)
			},
			)).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
				return r.watchDataScienceClusterResources(ctx, a)
			}),
			builder.WithPredicates(configMapPredicates),
		).
		Watches(
			&apiextensionsv1.CustomResourceDefinition{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
				return r.watchDataScienceClusterResources(ctx, a)
			}),
			builder.WithPredicates(argoWorkflowCRDPredicates),
		).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
				return r.watchDefaultIngressSecret(ctx, a)
			}),
			builder.WithPredicates(defaultIngressCertSecretPredicates)).
		// this predicates prevents meaningless reconciliations from being triggered
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}

func (r *DataScienceClusterReconciler) watchDataScienceClusterForDSCI(ctx context.Context, a client.Object) []reconcile.Request {
	requestName, err := r.getRequestName(ctx)
	if err != nil {
		return nil
	}
	// When DSCI CR gets created, trigger reconcile function
	if a.GetObjectKind().GroupVersionKind().Kind == "DSCInitialization" || a.GetName() == "default-dsci" {
		return []reconcile.Request{{
			NamespacedName: types.NamespacedName{Name: requestName},
		}}
	}
	return nil
}

func (r *DataScienceClusterReconciler) watchDataScienceClusterResources(ctx context.Context, a client.Object) []reconcile.Request {
	requestName, err := r.getRequestName(ctx)
	if err != nil {
		return nil
	}

	if a.GetObjectKind().GroupVersionKind().Kind == "CustomResourceDefinition" || a.GetName() == "ArgoWorkflowCRD" {
		return []reconcile.Request{{
			NamespacedName: types.NamespacedName{Name: requestName},
		}}
	}

	// Trigger reconcile function when uninstall configmap is created
	operatorNs, err := cluster.GetOperatorNamespace()
	if err != nil {
		return nil
	}
	if a.GetNamespace() == operatorNs {
		cmLabels := a.GetLabels()
		if val, ok := cmLabels[upgrade.DeleteConfigMapLabel]; ok && val == "true" {
			return []reconcile.Request{{
				NamespacedName: types.NamespacedName{Name: requestName},
			}}
		}
	}
	return nil
}

func (r *DataScienceClusterReconciler) getRequestName(ctx context.Context) (string, error) {
	instanceList := &dscv1.DataScienceClusterList{}
	err := r.Client.List(ctx, instanceList)
	if err != nil {
		return "", err
	}

	switch {
	case len(instanceList.Items) == 1:
		return instanceList.Items[0].Name, nil
	case len(instanceList.Items) == 0:
		return "default-dsc", nil
	default:
		return "", errors.New("multiple DataScienceCluster instances found")
	}
}

// argoWorkflowCRDPredicates filters the delete events to trigger reconcile when Argo Workflow CRD is deleted.
var argoWorkflowCRDPredicates = predicate.Funcs{
	DeleteFunc: func(e event.DeleteEvent) bool {
		if e.Object.GetName() == datasciencepipelinesctrl.ArgoWorkflowCRD {
			labelList := e.Object.GetLabels()
			// CRD to be deleted with label "app.opendatahub.io/datasciencepipeline":"true", should not trigger reconcile
			if value, exist := labelList[labels.ODH.Component(componentApi.DataSciencePipelinesComponentName)]; exist && value == "true" {
				return false
			}
		}
		// CRD to be deleted either not with label or label value is not "true", should trigger reconcile
		return true
	},
}

func (r *DataScienceClusterReconciler) watchDefaultIngressSecret(ctx context.Context, a client.Object) []reconcile.Request {
	requestName, err := r.getRequestName(ctx)
	if err != nil {
		return nil
	}
	// When ingress secret gets created/deleted, trigger reconcile function
	ingressCtrl, err := cluster.FindAvailableIngressController(ctx, r.Client)
	if err != nil {
		return nil
	}
	defaultIngressSecretName := cluster.GetDefaultIngressCertSecretName(ingressCtrl)
	if a.GetName() == defaultIngressSecretName && a.GetNamespace() == "openshift-ingress" {
		return []reconcile.Request{{
			NamespacedName: types.NamespacedName{Name: requestName},
		}}
	}
	return nil
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
