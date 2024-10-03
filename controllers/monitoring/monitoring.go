package monitoring

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	dscservicev1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/services/v1alpha1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
)

type MonitoringReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (m *MonitoringReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dscservicev1alpha1.DSCServices{}).
		Owns(
			&corev1.Service{}).
		// TODO: remove if we are not gonna use customized Prom stack
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(m.watchMonitoringConfigMapResource),
			builder.WithPredicates(CMContentChangedPredicate),
		).
		// TODO: remove if we are not gonna use customized Prom stack
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(m.watchMonitoringSecretResource),
			builder.WithPredicates(SecretContentChangedPredicate),
		).
		Owns(&corev1.Namespace{},
			builder.WithPredicates(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{}), monidtoringNSPredicates),
		).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(m)
}

var SecretContentChangedPredicate = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldSecret, _ := e.ObjectOld.(*corev1.Secret)
		newSecret, _ := e.ObjectNew.(*corev1.Secret)

		return !reflect.DeepEqual(oldSecret.Data, newSecret.Data)
	},
}

var CMContentChangedPredicate = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldCM, _ := e.ObjectOld.(*corev1.ConfigMap)
		newCM, _ := e.ObjectNew.(*corev1.ConfigMap)

		return !reflect.DeepEqual(oldCM.Data, newCM.Data)
	},
}

var monidtoringNSPredicates = predicate.Funcs{
	DeleteFunc: func(e event.DeleteEvent) bool {
		return true
	},
}

func (m *MonitoringReconciler) watchMonitoringConfigMapResource(_ context.Context, a client.Object) []reconcile.Request {
	if a.GetName() == "prometheus" && a.GetNamespace() == "redhat-ods-monitoring" {
		m.Log.Info("Found monitoring configmap has updated, start reconcile")

		return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: "prometheus", Namespace: "redhat-ods-monitoring"}}}
	}
	return []reconcile.Request{}
}

func (m *MonitoringReconciler) watchMonitoringSecretResource(_ context.Context, a client.Object) []reconcile.Request {
	operatorNs, err := cluster.GetOperatorNamespace()
	if err != nil {
		return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: "prometheus", Namespace: "redhat-ods-monitoring"}}}
	}

	if a.GetName() == "addon-managed-odh-parameters" && a.GetNamespace() == operatorNs {
		m.Log.Info("Found monitoring secret has updated, start reconcile")

		return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: "addon-managed-odh-parameters", Namespace: operatorNs}}}
	}
	return []reconcile.Request{}
}

func (m *MonitoringReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// TODO
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
	return ctrl.Result{}, nil
}

// UpdatePrometheusConfig update prometheus-configs.yaml to include/exclude <component>.rules
// parameter enable when set to true to add new rules, when set to false to remove existing rules.
// TODO: remove if we are not gonna use customized Prom stack
func (m *MonitoringReconciler) UpdatePrometheusConfig(_ client.Client, logger logr.Logger, enable bool, component string) error {
	prometheusconfigPath := filepath.Join("/opt/manifests", "monitoring", "prometheus", "apps", "prometheus-configs.yaml")

	// create a struct to mock poremtheus.yml
	type ConfigMap struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name      string `yaml:"name"`
			Namespace string `yaml:"namespace"`
		} `yaml:"metadata"`
		Data struct {
			PrometheusYML          string `yaml:"prometheus.yml"`
			OperatorRules          string `yaml:"operator-recording.rules"`
			DeadManSnitchRules     string `yaml:"deadmanssnitch-alerting.rules"`
			CFRRules               string `yaml:"codeflare-recording.rules"`
			CRARules               string `yaml:"codeflare-alerting.rules"`
			DashboardRRules        string `yaml:"rhods-dashboard-recording.rules"`
			DashboardARules        string `yaml:"rhods-dashboard-alerting.rules"`
			DSPRRules              string `yaml:"data-science-pipelines-operator-recording.rules"`
			DSPARules              string `yaml:"data-science-pipelines-operator-alerting.rules"`
			MMRRules               string `yaml:"model-mesh-recording.rules"`
			MMARules               string `yaml:"model-mesh-alerting.rules"`
			OdhModelRRules         string `yaml:"odh-model-controller-recording.rules"`
			OdhModelARules         string `yaml:"odh-model-controller-alerting.rules"`
			RayARules              string `yaml:"ray-alerting.rules"`
			WorkbenchesRRules      string `yaml:"workbenches-recording.rules"`
			WorkbenchesARules      string `yaml:"workbenches-alerting.rules"`
			KserveRRules           string `yaml:"kserve-recording.rules"`
			KserveARules           string `yaml:"kserve-alerting.rules"`
			TrustyAIRRules         string `yaml:"trustyai-recording.rules"`
			TrustyAIARules         string `yaml:"trustyai-alerting.rules"`
			KueueRRules            string `yaml:"kueue-recording.rules"`
			KueueARules            string `yaml:"kueue-alerting.rules"`
			TrainingOperatorRRules string `yaml:"trainingoperator-recording.rules"`
			TrainingOperatorARules string `yaml:"trainingoperator-alerting.rules"`
			ModelRegistryRRules    string `yaml:"model-registry-operator-recording.rules"`
			ModelRegistryARules    string `yaml:"model-registry-operator-alerting.rules"`
		} `yaml:"data"`
	}
	var configMap ConfigMap
	// prometheusContent will represent content of prometheus.yml due to its dynamic struct
	var prometheusContent map[interface{}]interface{}

	// read prometheus.yml from local disk /opt/mainfests/monitoring/prometheus/apps/
	yamlData, err := os.ReadFile(prometheusconfigPath)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(yamlData, &configMap); err != nil {
		return err
	}

	// get prometheus.yml part from configmap
	if err := yaml.Unmarshal([]byte(configMap.Data.PrometheusYML), &prometheusContent); err != nil {
		return err
	}

	// to add component rules when it is not there yet
	if enable {
		// Check if the rule not yet exists in rule_files
		if !strings.Contains(configMap.Data.PrometheusYML, component+"*.rules") {
			// check if have rule_files
			if ruleFiles, ok := prometheusContent["rule_files"]; ok {
				if ruleList, isList := ruleFiles.([]interface{}); isList {
					// add new component rules back to rule_files
					ruleList = append(ruleList, component+"*.rules")
					prometheusContent["rule_files"] = ruleList
				}
			}
		}
	} else { // to remove component rules if it is there
		logger.Info("Removing prometheus rule: " + component + "*.rules")
		if ruleList, ok := prometheusContent["rule_files"].([]interface{}); ok {
			for i, item := range ruleList {
				if rule, isStr := item.(string); isStr && rule == component+"*.rules" {
					ruleList = append(ruleList[:i], ruleList[i+1:]...)

					break
				}
			}
			prometheusContent["rule_files"] = ruleList
		}
	}

	// Marshal back
	newDataYAML, err := yaml.Marshal(&prometheusContent)
	if err != nil {
		return err
	}
	configMap.Data.PrometheusYML = string(newDataYAML)

	newyamlData, err := yaml.Marshal(&configMap)
	if err != nil {
		return err
	}

	// Write the modified content back to the file
	err = os.WriteFile(prometheusconfigPath, newyamlData, 0)

	return err
}
