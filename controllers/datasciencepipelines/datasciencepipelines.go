package datasciencepipelines

import (
	"context"
	"fmt"
	"path/filepath"

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
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/components/datasciencepipelines"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DataSciencePipelineReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

var (
	ComponentName   = "data-science-pipelines-operator"
	Path            = deploy.DefaultManifestPath + "/" + ComponentName + "/base"
	OverlayPath     = deploy.DefaultManifestPath + "/" + ComponentName + "/overlays"
	ArgoWorkflowCRD = "workflows.argoproj.io"
)

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
	return []reconcile.Request{}
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
			if err = d.DeployManifests(ctx, d.Client, d.Log, obj, &obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	d.Log.Info("DataSciencePipeline CR has been created/updated.", "Request.Name", request.Name)
	d.DeployManifests(ctx, d.Client, d.Log, obj, &obj.Spec.ComponentSpec, true)
}

func (d *DataSciencePipelineReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete DataSciencePipelines Component CR
	dspCR := &dsccomponentv1alpha1.DataSciencePipeline{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DataSciencePipeline",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default-dsp",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.DataSciencePipelineSpec{
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
		cli.Create(ctx, dspCR)
	} else {
		cli.Delete(ctx, dspCR)
	}
	return nil
}
func (d *DataSciencePipelineReconciler) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(d.DevFlags.Manifests) != 0 {
		manifestConfig := d.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentName, manifestConfig); err != nil {
			return err
		}
		// If overlay is defined, update paths
		defaultKustomizePath := "base"
		if manifestConfig.SourcePath != "" {
			defaultKustomizePath = manifestConfig.SourcePath
		}
		Path = filepath.Join(deploy.DefaultManifestPath, ComponentName, defaultKustomizePath)
	}

	return nil
}

func (d *DataSciencePipelineReconciler) GetComponentName() string {
	return ComponentName
}

func (d *DataSciencePipelineReconciler) DeployManifests(ctx context.Context,
	cli client.Client,
	l logr.Logger,
	owner metav1.Object,
	componentSpec *dsccomponentv1alpha1.ComponentSpec,
	currentComponentExist bool,
) error {
	var imageParamMap = map[string]string{
		// v1
		"IMAGES_APISERVER":         "RELATED_IMAGE_ODH_ML_PIPELINES_API_SERVER_IMAGE",
		"IMAGES_ARTIFACT":          "RELATED_IMAGE_ODH_ML_PIPELINES_ARTIFACT_MANAGER_IMAGE",
		"IMAGES_PERSISTENTAGENT":   "RELATED_IMAGE_ODH_ML_PIPELINES_PERSISTENCEAGENT_IMAGE",
		"IMAGES_SCHEDULEDWORKFLOW": "RELATED_IMAGE_ODH_ML_PIPELINES_SCHEDULEDWORKFLOW_IMAGE",
		"IMAGES_CACHE":             "RELATED_IMAGE_ODH_ML_PIPELINES_CACHE_IMAGE",
		"IMAGES_DSPO":              "RELATED_IMAGE_ODH_DATA_SCIENCE_PIPELINES_OPERATOR_CONTROLLER_IMAGE",
		// v2
		"IMAGESV2_ARGO_APISERVER":          "RELATED_IMAGE_ODH_ML_PIPELINES_API_SERVER_V2_IMAGE",
		"IMAGESV2_ARGO_PERSISTENCEAGENT":   "RELATED_IMAGE_ODH_ML_PIPELINES_PERSISTENCEAGENT_V2_IMAGE",
		"IMAGESV2_ARGO_SCHEDULEDWORKFLOW":  "RELATED_IMAGE_ODH_ML_PIPELINES_SCHEDULEDWORKFLOW_V2_IMAGE",
		"IMAGESV2_ARGO_ARGOEXEC":           "RELATED_IMAGE_ODH_DATA_SCIENCE_PIPELINES_ARGO_ARGOEXEC_IMAGE",
		"IMAGESV2_ARGO_WORKFLOWCONTROLLER": "RELATED_IMAGE_ODH_DATA_SCIENCE_PIPELINES_ARGO_WORKFLOWCONTROLLER_IMAGE",
		"V2_DRIVER_IMAGE":                  "RELATED_IMAGE_ODH_ML_PIPELINES_DRIVER_IMAGE",
		"V2_LAUNCHER_IMAGE":                "RELATED_IMAGE_ODH_ML_PIPELINES_LAUNCHER_IMAGE",
		"IMAGESV2_ARGO_MLMDGRPC":           "RELATED_IMAGE_ODH_MLMD_GRPC_SERVER_IMAGE",
	}

	enabled := d.GetManagementState() == operatorv1.Managed
	monitoringEnabled := componentSpec.DSCISpec.Monitoring.ManagementState == operatorv1.Managed

	if enabled {
		if d.DevFlags != nil {
			// Download manifests and update paths
			if err := d.OverrideManifests(ctx, componentSpec.Platform); err != nil {
				return err
			}
		}
		// skip check if the dependent operator has beeninstalled, this is done in dashboard
		// Update image parameters only when we do not have customized manifests set
		if d.DevFlags == nil || len(d.DevFlags.Manifests) == 0 {
			if err := deploy.ApplyParams(Path, imageParamMap); err != nil {
				return fmt.Errorf("failed to update image from %s : %w", Path, err)
			}
		}
		// Check for existing Argo Workflows
		if err := UnmanagedArgoWorkFlowExists(ctx, cli); err != nil {
			return err
		}
	}

	// new overlay
	manifestsPath := filepath.Join(OverlayPath, "rhoai")
	if componentSpec.Platform == cluster.OpenDataHub || componentSpec.Platform == "" {
		manifestsPath = filepath.Join(OverlayPath, "odh")
	}
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, manifestsPath, componentSpec.DSCISpec.ApplicationsNamespace, ComponentName, enabled); err != nil {
		return err
	}
	l.Info("apply manifests done")

	// Wait for deployment available
	if enabled {
		if err := cluster.WaitForDeploymentAvailable(ctx, cli, ComponentName, componentSpec.DSCISpec.ApplicationsNamespace, 20, 2); err != nil {
			return fmt.Errorf("deployment for %s is not ready to server: %w", ComponentName, err)
		}
	}

	// CloudService Monitoring handling
	if componentSpec.Platform == cluster.ManagedRhods {
		if err := d.UpdatePrometheusConfig(cli, l, enabled && monitoringEnabled, ComponentName); err != nil {
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
