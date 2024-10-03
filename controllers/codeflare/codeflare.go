package codeflare

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	operatorv1 "github.com/openshift/api/operator/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
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

type CodeFlareReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

var (
	ComponentName     = "codeflare"
	CodeflarePath     = deploy.DefaultManifestPath + "/" + ComponentName + "/default"
	CodeflareOperator = "codeflare-operator"
	ParamsPath        = deploy.DefaultManifestPath + "/" + ComponentName + "/manager"
)

// SetupWithManager sets up the controller with the Manager.
func (c *CodeFlareReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dsccomponentv1alpha1.CodeFlare{}).
		Owns(&corev1.Secret{}).
		Owns(&admissionregistrationv1.MutatingWebhookConfiguration{}).
		Owns(&admissionregistrationv1.ValidatingWebhookConfiguration{}).
		Owns(
			&appsv1.Deployment{}, builder.WithPredicates(componentDeploymentPredicates)).
		Owns(
			&corev1.Service{}).
		Owns(
			&corev1.ServiceAccount{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(c)
}

// reduce unnecessary reconcile triggered by codeflare's deployment change due to ManagedByODHOperator annotation.
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

func (c *CodeFlareReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {

	// check if component is Managed or not
	// TODO : component.GetManagementState()
	mockManagementStatus := operatorv1.Managed

	// Fetch the CodeFlareComponent instance to know created or deleted
	obj := &dsccomponentv1alpha1.CodeFlare{}
	err := c.Client.Get(ctx, request.NamespacedName, obj)

	if err = c.CreateComponentCR(ctx, c.Client, instance, &dsciInstances.Items[0], mockManagementStatus == operatorv1.Managed); err != nil {
		componentErrors = multierror.Append(componentErrors, err)

	}

	// deletion case
	if err != nil {
		if k8serr.IsNotFound(err) {
			c.Log.Info("CodeFlare CR has been deleted.", "Request.Name", request.Name)
			if err = c.DeployManifests(ctx, c.Client, c.Log, obj, &obj.Spec.ComponentSpec, true); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	c.Log.Info("Dashboard CR has been created/updated.", "Request.Name", request.Name)
	return ctrl.Result{}, c.DeployManifests(ctx, c.Client, c.Log, obj, &obj.Spec.ComponentSpec, true)
}

func (d *CodeFlareReconciler) CreateComponentCR(ctx context.Context, cli client.Client, owner metav1.Object, dsci *dsciv1.DSCInitialization, enabled bool) error {

	// create/delete CodeFlare Component CR
	cfoCR := &dsccomponentv1alpha1.CodeFlare{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CodeFlare",
			APIVersion: "components.opendatahub.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk.DataScienceCluster)},
		},
		Spec: dsccomponentv1alpha1.CodeFlareSpec{
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
		cli.Create(ctx, cfoCR)
	} else {
		cli.Delete(ctx, cfoCR)
	}
	return nil
}
func (c *CodeFlareReconciler) OverrideManifests(ctx context.Context, _ cluster.Platform) error {
	// If devflags are set, update default manifests path
	if len(c.DevFlags.Manifests) != 0 {
		manifestConfig := c.DevFlags.Manifests[0]
		if err := deploy.DownloadManifests(ctx, ComponentName, manifestConfig); err != nil {
			return err
		}
		// If overlay is defined, update paths
		defaultKustomizePath := "default"
		if manifestConfig.SourcePath != "" {
			defaultKustomizePath = manifestConfig.SourcePath
		}
		CodeflarePath = filepath.Join(deploy.DefaultManifestPath, ComponentName, defaultKustomizePath)
	}

	return nil
}

func (c *CodeFlareReconciler) GetComponentName() string {
	return ComponentName
}

func (c *CodeFlareReconciler) DeployManifests(ctx context.Context,
	cli client.Client,
	l logr.Logger,
	owner metav1.Object,
	componentSpec *dsccomponentv1alpha1.ComponentSpec,
	_ bool) error {
	var imageParamMap = map[string]string{
		"codeflare-operator-controller-image": "RELATED_IMAGE_ODH_CODEFLARE_OPERATOR_IMAGE", // no need mcad, embedded in cfo
	}
	obj := (owner).(*dsccomponentv1alpha1.CodeFlare)
	enabled := c.GetManagementState() == operatorv1.Managed

	if enabled {
		// if c.DevFlags != nil {
		// 	// Download manifests and update paths
		// 	if err := c.OverrideManifests(ctx, componentSpec.Platform); err != nil {
		// 		return err
		// 	}
		// }
		// check if the CodeFlare operator is installed: it should not be installed
		// Both ODH and RHOAI should have the same operator name
		dependentOperator := CodeflareOperator

		if found, err := cluster.OperatorExists(ctx, cli, dependentOperator); err != nil {
			return fmt.Errorf("operator exists throws error %w", err)
		} else if found {
			return fmt.Errorf("operator %s is found. Please uninstall the operator before enabling %s component",
				dependentOperator, ComponentName)
		}

		// Update image parameters only when we do not have customized manifests set
		// if c.DevFlags == nil || len(c.DevFlags.Manifests) == 0 {
		// 	if err := deploy.ApplyParams(ParamsPath, imageParamMap, map[string]string{"namespace": componentSpec.ApplicationsNamespace}); err != nil {
		// 		return fmt.Errorf("failed update image from %s : %w", CodeflarePath+"/bases", err)
		// 	}
		// }
	}

	// Deploy Codeflare
	if err := deploy.DeployManifestsFromPath(ctx, cli, owner, //nolint:revive,nolintlint
		CodeflarePath,
		componentSpec.ApplicationsNamespace,
		ComponentName, enabled); err != nil {
		return err
	}
	l.Info("apply manifests done")

	if enabled {
		if err := cluster.WaitForDeploymentAvailable(ctx, cli, ComponentName, componentSpec.ApplicationsNamespace, 20, 2); err != nil {
			return fmt.Errorf("deployment for %s is not ready to server: %w", ComponentName, err)
		}
	}

	return nil
}
