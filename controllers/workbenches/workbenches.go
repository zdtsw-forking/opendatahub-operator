package workbenches

import (
	"context"

	"github.com/go-logr/logr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
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
	"github.com/opendatahub-io/opendatahub-operator/v2/components/workbenches"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
		dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
)


var (
	ComponentName          = "workbenches"
	DependentComponentName = "notebooks"
	// manifests for nbc in ODH and RHOAI + downstream use it for imageparams.
	notebookControllerPath = deploy.DefaultManifestPath + "/odh-notebook-controller/odh-notebook-controller/base"
	// manifests for ODH nbc + downstream use it for imageparams.
	kfnotebookControllerPath = deploy.DefaultManifestPath + "/odh-notebook-controller/kf-notebook-controller/overlays/openshift"
	// notebook image manifests.
	notebookImagesPath = deploy.DefaultManifestPath + "/notebooks/overlays/additional"
)
type WorkbenchReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (w *WorkbenchReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.Workbench{}).
		Owns(&admissionregistrationv1.MutatingWebhookConfiguration{}).
		Owns(&admissionregistrationv1.ValidatingWebhookConfiguration{}).
		Owns(&corev1.Secret{}).
		Owns(
			&appsv1.Deployment{}, builder.WithPredicates(componentDeploymentPredicates)).
		Owns(
			&corev1.Service{}).
		Owns(
			&corev1.ServiceAccount{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(w)
}

// reduce unnecessary reconcile triggered by wb's deployment change due to ManagedByODHOperator annotation.
var componentDeploymentPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		namespace := e.ObjectNew.GetNamespace()
		if (namespace == "opendatahub" || namespace == "redhat-ods-applications") && e.ObjectNew.GetLabels()[labels.K8SCommon.PartOf] == workbenches.ComponentName {
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

func (w *WorkbenchReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// Fetch the WorkbenchComponent instance to know created or deleted
	obj := &dsccomponentv1alpha1.Workbench{}
	err := w.Client.Get(ctx, request.NamespacedName, obj)

	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) || obj.GetDeletionTimestamp() != nil {
			w.Log.Info("Workbench CR has been deleted.", "Request.Name", request.Name)
			if err = w.DeployManifests(ctx, w.Client, w.Log, obj, &obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	w.Log.Info("Workbench CR has been createw.", "Request.Name", request.Name)
	w.DeployManifests(ctx, w.Client, w.Log, obj, &obj.Spec.ComponentSpec, true)
}

func (w *WorkbenchReconciler) OverrideManifests(ctx context.Context, platform cluster.Platform) error {
	// Download manifests if defined by devflags
	// Go through each manifest and set the overlays if defined
	// first on odh-notebook-controller and kf-notebook-controller last to notebook-images
	for _, subcomponent := range w.DevFlags.Manifests {
		if strings.Contains(subcomponent.ContextDir, "components/odh-notebook-controller") {
			// Download subcomponent
			if err := deploy.DownloadManifests(ctx, "odh-notebook-controller/odh-notebook-controller", subcomponent); err != nil {
				return err
			}
			// If overlay is defined, update paths
			defaultKustomizePathNbc := "base"
			if subcomponent.SourcePath != "" {
				defaultKustomizePathNbc = subcomponent.SourcePath
			}
			notebookControllerPath = filepath.Join(deploy.DefaultManifestPath, "odh-notebook-controller/odh-notebook-controller", defaultKustomizePathNbc)
		}

		if strings.Contains(subcomponent.ContextDir, "components/notebook-controller") {
			// Download subcomponent
			if err := deploy.DownloadManifests(ctx, "odh-notebook-controller/kf-notebook-controller", subcomponent); err != nil {
				return err
			}
			// If overlay is defined, update paths
			defaultKustomizePathKfNbc := "overlays/openshift"
			if subcomponent.SourcePath != "" {
				defaultKustomizePathKfNbc = subcomponent.SourcePath
			}
			kfnotebookControllerPath = filepath.Join(deploy.DefaultManifestPath, "odh-notebook-controller/kf-notebook-controller", defaultKustomizePathKfNbc)
		}
		if strings.Contains(subcomponent.URI, DependentComponentName) {
			// Download subcomponent
			if err := deploy.DownloadManifests(ctx, DependentComponentName, subcomponent); err != nil {
				return err
			}
			// If overlay is defined, update paths
			defaultKustomizePath := "overlays/additional"
			if subcomponent.SourcePath != "" {
				defaultKustomizePath = subcomponent.SourcePath
			}
			notebookImagesPath = filepath.Join(deploy.DefaultManifestPath, DependentComponentName, defaultKustomizePath)
		}
	}
	return nil
}

func (w *WorkbenchReconciler) GetComponentName() string {
	return ComponentName
}

func (d *WorkbenchReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	// create/delete Workbenches Component CR
	wbCR := &dsccomponentv1alpha1.DSCWorkbench{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DSCRay",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default-workbench",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.WBComponentSpec{
			ComponentSpec: dsccomponentv1alpha1.ComponentSpec{
				Platform:              dsci.Status.Release.Name,
				ComponentName:         ComponentName,
				ApplicationsNamespace: dsci.Spec.ApplicationsNamespace,
				Monitoring:            dsci.Spec.Monitoring,
			},
		},
	}
	if enabled {
		cli.Create(ctx, wbCR)
	} else {
		cli.Delete(ctx, wbCR)
	}
	return nil
}
func (w *WorkbenchReconciler) DeployManifests(ctx context.Context, cli client.Client, l logr.Logger,
	owner metav1.Object, componentSpec *dsccomponentv1alpha1.ComponentSpec, _ bool) error {
	var imageParamMap = map[string]string{
		"odh-notebook-controller-image":    "RELATED_IMAGE_ODH_NOTEBOOK_CONTROLLER_IMAGE",
		"odh-kf-notebook-controller-image": "RELATED_IMAGE_ODH_KF_NOTEBOOK_CONTROLLER_IMAGE",
	}

	// Set default notebooks namespace
	// Create rhods-notebooks namespace in managed platforms
	enabled := w.GetManagementState() == operatorv1.Managed
	monitoringEnabled := componentSpec.Monitoring.ManagementState == operatorv1.Managed
	if enabled {
		if w.DevFlags != nil {
			// Download manifests and update paths
			if err := w.OverrideManifests(ctx, componentSpec.Platform); err != nil {
				return err
			}
		}
		if componentSpec.Platform == cluster.SelfManagedRhods || componentSpec.Platform == cluster.ManagedRhods {
			// Intentionally leaving the ownership unset for this namespace.
			// Specifying this label triggers its deletion when the operator is uninstalled.
			_, err := cluster.CreateNamespace(ctx, cli, cluster.DefaultNotebooksNamespace, cluster.WithLabels(labels.ODH.OwnedNamespace, "true"))
			if err != nil {
				return err
			}
		}
		// Update Default rolebinding
		err := cluster.UpdatePodSecurityRolebinding(ctx, cli, componentSpec.ApplicationsNamespace, "notebook-controller-service-account")
		if err != nil {
			return err
		}
	}

	// Update image parameters for nbc
	if enabled {
		if w.DevFlags == nil || len(w.DevFlags.Manifests) == 0 {
			// for kf-notebook-controller image
			if err := deploy.ApplyParams(notebookControllerPath, imageParamMap); err != nil {
				return fmt.Errorf("failed to update image %s: %w", notebookControllerPath, err)
			}
			// for odh-notebook-controller image
			if err := deploy.ApplyParams(kfnotebookControllerPath, imageParamMap); err != nil {
				return fmt.Errorf("failed to update image %s: %w", kfnotebookControllerPath, err)
			}
		}
	}
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner,
		notebookControllerPath,
		componentSpec.ApplicationsNamespace,
		ComponentName, enabled); err != nil {
		return fmt.Errorf("failed to apply manifetss %s: %w", notebookControllerPath, err)
	}
	l.WithValues("Path", notebookControllerPath).Info("apply manifests done notebook controller done")

	if err := deploy.DeployManifestsFromPath(ctx, cli, owner,
		kfnotebookControllerPath,
		componentSpec.ApplicationsNamespace,
		ComponentName, enabled); err != nil {
		return fmt.Errorf("failed to apply manifetss %s: %w", kfnotebookControllerPath, err)
	}
	l.WithValues("Path", kfnotebookControllerPath).Info("apply manifests done kf-notebook controller done")

	if err := deploy.DeployManifestsFromPath(ctx, cli, owner,
		notebookImagesPath,
		componentSpec.ApplicationsNamespace,
		ComponentName, enabled); err != nil {
		return err
	}
	l.WithValues("Path", notebookImagesPath).Info("apply manifests done notebook image done")

	// Wait for deployment available
	if enabled {
		if err := cluster.WaitForDeploymentAvailable(ctx, cli, ComponentName, componentSpec.ApplicationsNamespace, 10, 2); err != nil {
			return fmt.Errorf("deployments for %s are not ready to server: %w", ComponentName, err)
		}
	}

	// CloudService Monitoring handling
	if componentSpec.Platform == cluster.ManagedRhods {
		if err := w.UpdatePrometheusConfig(cli, l, enabled && monitoringEnabled, ComponentName); err != nil {
			return err
		}
		if err := deploy.DeployManifestsFromPath(ctx, cli, owner,
			filepath.Join(deploy.DefaultManifestPath, "monitoring", "prometheus", "apps"),
			componentSpec.Monitoring.Namespace,
			"prometheus", true); err != nil {
			return err
		}
		l.Info("updating SRE monitoring done")
	}
	return nil
}

