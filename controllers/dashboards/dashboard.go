package dashboard

import (
	"context"

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
	dscv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/datasciencecluster/v1"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/dscinitialization/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/components"
	"github.com/opendatahub-io/opendatahub-operator/v2/components/dashboard"
	annotations "github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
	operatorv1 "github.com/openshift/api/operator/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
		"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
		"github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
)

type DashboardReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (d *DashboardReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.Dashboard{}).
		Owns(&corev1.Secret{}).
		Owns(
			&appsv1.Deployment{}, builder.WithPredicates(componentDeploymentPredicates)).
		Owns(
			&corev1.Service{}).
		Owns(
			&corev1.ServiceAccount{}).
		Owns(
			&dscv1.DataScienceCluster{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(d)
}

// reduce unnecessary reconcile triggered by dashboard's deployment change due to ManagedByODHOperator annotation.
var componentDeploymentPredicates = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		namespace := e.ObjectNew.GetNamespace()
		if (namespace == "opendatahub" && e.ObjectNew.GetLabels()[labels.K8SCommon.PartOf] == dashboard.ComponentNameUpstream) ||
			(namespace == "redhat-ods-applications" && e.ObjectNew.GetLabels()[labels.K8SCommon.PartOf] == dashboard.ComponentNameDownstream) {
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

func (d *DashboardReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// Fetch the DashboardComponent instance to know created or deleted
	obj := &dsccomponentv1alpha1.Dashboard{}
	err := d.Client.Get(ctx, request.NamespacedName, obj)
	if obj.GetName() != request.Name && obj.GetOwnerReferences()[0].Name != "default-dsc" {
		return ctrl.Result{}, nil
	}
	var _ components.ComponentInterface = (*DashboardReconciler)(nil)

	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) || obj.GetDeletionTimestamp() != nil {
			d.Log.Info("Dashboard CR has been deleted.", "Request.Name", request.Name)
			if err = d.DeployManifests(ctx, d.Client, d.Log, obj, obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	d.Log.Info("Dashboard CR has been created.", "Request.Name", request.Name)

	instanceList := &dscv1.DataScienceClusterList{}
	if err := d.Client.List(ctx, instanceList); err != nil {
		d.Log.Error(err, "Failed to get DataScienceClusterList")
		return ctrl.Result{}, err
	}
	installedComponentValue, _ := instanceList.Items[0].Status.InstalledComponents["dashboard"]
	d.DeployManifests(ctx, d.Client, d.Log, obj, obj.Spec.ComponentSpec, installedComponentValue)
}


var (
	ComponentNameUpstream = "dashboard"
	PathUpstream          = deploy.DefaultManifestPath + "/" + ComponentNameUpstream + "/odh"

	ComponentNameDownstream = "rhods-dashboard"
	PathDownstream          = deploy.DefaultManifestPath + "/" + ComponentNameUpstream + "/rhoai"
	PathSelfDownstream      = PathDownstream + "/onprem"
	PathManagedDownstream   = PathDownstream + "/addon"
	OverridePath            = ""
)

// Verifies that Dashboard implements ComponentInterface.
//var _ components.ComponentInterface = (*Dashboard)(nil)

// Dashboard struct holds the configuration for the Dashboard component.
// +kubebuilder:object:generate=true
type Dashboard struct {
	components.Component `json:""`
}

func (d *DashboardReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {
	componentName := ComponentNameUpstream
	if dsci.Status.Release.Name == cluster.SelfManagedRhods || dsci.Status.Release.Name == cluster.ManagedRhods {
		componentName = ComponentNameDownstream
	}
	// create/delete DashboardComponent CR
	dashboardCR := &dsccomponentv1alpha1.Dashboard{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Dashboard",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default-dashboard",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.DashboardComponentSpec{
			ComponentSpec: dsccomponentv1alpha1.ComponentSpec{
				Platform:              dsci.Status.Release.Name,
				ComponentName:         "dashboard",
				DSCInitializationSpec: dsci.Spec,
				ComponentDevFlags: dsccomponentv1alpha1.DevFlags{
					LoggerMode: dsci.Spec.DevFlags.LogMode,
					// DSCDevFlags: dsccomponentv1alpha1.DSCDevFlags{
					// 	Manifests: [], // TODO
					// },
				},
			},
		},
	}
	if enabled {
		cli.Create(ctx, dashboardCR)
	} else {
		cli.Delete(ctx, dashboardCR)
	}
	return nil
}

func (d *DashboardReconciler) OverrideManifests(ctx context.Context, platform cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(d.DevFlags.Manifests) != 0 {
		manifestConfig := d.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentNameUpstream, manifestConfig); err != nil {
			return err
		}
		if manifestConfig.SourcePath != "" {
			OverridePath = filepath.Join(deploy.DefaultManifestPath, ComponentNameUpstream, manifestConfig.SourcePath)
		}
	}
	return nil
}
var (
	ComponentNameUpstream = "dashboard"
	PathUpstream          = deploy.DefaultManifestPath + "/" + ComponentNameUpstream + "/odh"

	ComponentNameDownstream = "rhods-dashboard"
	PathDownstream          = deploy.DefaultManifestPath + "/" + ComponentNameUpstream + "/rhoai"
	PathSelfDownstream      = PathDownstream + "/onprem"
	PathManagedDownstream   = PathDownstream + "/addon"
	OverridePath            = ""
)

func (d *DashboardReconciler) GetComponentName() string {
	return ComponentNameUpstream
}

func (d *DashboardReconciler) DeployManifests(ctx context.Context,
	cli client.Client,
	l logr.Logger,
	owner metav1.Object,
	componentSpec *dsccomponentv1alpha1.ComponentSpec,
	currentComponentExist bool,
) error {
	entryPath := map[cluster.Platform]string{
		cluster.SelfManagedRhods: PathDownstream + "/onprem",
		cluster.ManagedRhods:     PathDownstream + "/addon",
		cluster.OpenDataHub:      PathUpstream,
		cluster.Unknown:          PathUpstream,
	}[componentSpec.Platform]

	enabled :=  == operatorv1.Managed
	monitoringEnabled := componentSpec.DSCISpec.Monitoring.ManagementState == operatorv1.Managed
	imageParamMap := make(map[string]string)

	if enabled {
		// 1. cleanup OAuth client related secret and CR if dashboard is in 'installed false' status
		if err := d.cleanOauthClient(ctx, cli, componentSpec, currentComponentExist, l); err != nil {
			return err
		}
		if d.DevFlags != nil && len(d.DevFlags.Manifests) != 0 {
			// Download manifests and update paths
			if err := d.OverrideManifests(ctx, componentSpec.Platform); err != nil {
				return err
			}
			if OverridePath != "" {
				entryPath = OverridePath
			}
		} else { // Update image parameters if devFlags is not provided
			imageParamMap["odh-dashboard-image"] = "RELATED_IMAGE_ODH_DASHBOARD_IMAGE"
		}

		// 2. platform specific RBAC
		if componentSpec.Platform == cluster.OpenDataHub || componentSpec.Platform == "" {
			if err := cluster.UpdatePodSecurityRolebinding(ctx, cli, componentSpec.DSCISpec.ApplicationsNamespace, "odh-dashboard"); err != nil {
				return err
			}
		} else {
			if err := cluster.UpdatePodSecurityRolebinding(ctx, cli, componentSpec.DSCISpec.ApplicationsNamespace, "rhods-dashboard"); err != nil {
				return err
			}
		}

		// 3. Append or Update variable for component to consume
		extraParamsMap, err := updateKustomizeVariable(ctx, cli, componentSpec.Platform, componentSpec)
		if err != nil {
			return errors.New("failed to set variable for extraParamsMap")
		}

		// 4. update params.env regardless devFlags is provided of not
		if err := deploy.ApplyParams(entryPath, imageParamMap, extraParamsMap); err != nil {
			return fmt.Errorf("failed to update params.env  from %s : %w", entryPath, err)
		}
	}

	// common: Deploy odh-dashboard manifests
	// TODO: check if we can have the same component name odh-dashboard for both, or still keep rhods-dashboard for RHOAI
	switch componentSpec.Platform {
	case cluster.SelfManagedRhods, cluster.ManagedRhods:
		// anaconda
		if err := cluster.CreateSecret(ctx, cli, "anaconda-ce-access", componentSpec.DSCISpec.ApplicationsNamespace); err != nil {
			return fmt.Errorf("failed to create access-secret for anaconda: %w", err)
		}
		// Deploy RHOAI manifests
		if err := deploy.DeployManifestsFromPath(ctx, cli, owner, entryPath, componentSpec.DSCISpec.ApplicationsNamespace, ComponentNameDownstream, enabled); err != nil {
			return fmt.Errorf("failed to apply manifests from %s: %w", PathDownstream, err)
		}
		l.Info("apply manifests done")

		if enabled {
			if err := cluster.WaitForDeploymentAvailable(ctx, cli, ComponentNameDownstream, componentSpec.DSCISpec.ApplicationsNamespace, 20, 3); err != nil {
				return fmt.Errorf("deployment for %s is not ready to server: %w", ComponentNameDownstream, err)
			}
		}

		// CloudService Monitoring handling
		if componentSpec.Platform == cluster.ManagedRhods {
			if err := d.UpdatePrometheusConfig(cli, l, enabled && monitoringEnabled, ComponentNameDownstream); err != nil {
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

	default:
		// Deploy ODH manifests
		if err := deploy.DeployManifestsFromPath(ctx, cli, owner, entryPath, componentSpec.DSCISpec.ApplicationsNamespace, ComponentNameUpstream, enabled); err != nil {
			return err
		}
		l.Info("apply manifests done")
		if enabled {
			if err := cluster.WaitForDeploymentAvailable(ctx, cli, ComponentNameUpstream, componentSpec.DSCISpec.ApplicationsNamespace, 20, 3); err != nil {
				return fmt.Errorf("deployment for %s is not ready to server: %w", ComponentNameUpstream, err)
			}
		}

		return nil
	}
}

func (d *DashboardReconciler) cleanOauthClient(ctx context.Context, cli client.Client, componentSpec *dsccomponentv1alpha1.ComponentSpec, currentComponentExist bool, l logr.Logger) error {
	// Remove previous oauth-client secrets
	// Check if component is going from state of `Not Installed --> Installed`
	// Assumption: Component is currently set to enabled
	name := "dashboard-oauth-client"
	if !currentComponentExist {
		l.Info("Cleanup any left secret")
		// Delete client secrets from previous installation
		oauthClientSecret := &corev1.Secret{}
		err := cli.Get(ctx, client.ObjectKey{
			Namespace: componentSpec.DSCISpec.ApplicationsNamespace,
			Name:      name,
		}, oauthClientSecret)
		if err != nil {
			if !k8serr.IsNotFound(err) {
				return fmt.Errorf("error getting secret %s: %w", name, err)
			}
		} else {
			if err := cli.Delete(ctx, oauthClientSecret); err != nil {
				return fmt.Errorf("error deleting secret %s: %w", name, err)
			}
			l.Info("successfully deleted secret", "secret", name)
		}
	}
	return nil
}

