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

package webhook

import (
	"context"
	"fmt"
	"net/http"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	// "github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster".
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
)

var log = ctrl.Log.WithName("odh-controller-webhook")
var endpoint = "/validate-opendatahub-io-v1"

//+kubebuilder:webhook:path=/validate-opendatahub-io-v1,mutating=false,failurePolicy=fail,sideEffects=None,groups=datasciencecluster.opendatahub.io;dscinitialization.opendatahub.io,resources=datascienceclusters;dscinitializations,verbs=create;update;delete,versions=v1,name=operator.opendatahub.io,admissionReviewVersions=v1
//nolint:lll

type OpenDataHubWebhook struct {
	client  client.Client
	decoder *admission.Decoder
}

func (w *OpenDataHubWebhook) SetupWithManager(mgr ctrl.Manager) {
	hookServer := mgr.GetWebhookServer()
	odhWebhook := &webhook.Admission{
		Handler: w,
	}
	hookServer.Register(endpoint, odhWebhook)

	// extra endpoint for webhookreadyz
	hookServer.Register("/webhookreadyz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func (w *OpenDataHubWebhook) InjectDecoder(d *admission.Decoder) error {
	w.decoder = d
	return nil
}

func (w *OpenDataHubWebhook) InjectClient(c client.Client) error {
	w.client = c
	return nil
}

func countObjects(ctx context.Context, cli client.Client, gvk schema.GroupVersionKind) (int, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(gvk)

	if err := cli.List(ctx, list); err != nil {
		return 0, err
	}

	return len(list.Items), nil
}

func denyCountGtZero(ctx context.Context, cli client.Client, gvk schema.GroupVersionKind, msg string) admission.Response {
	count, err := countObjects(ctx, cli, gvk)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if count > 0 {
		return admission.Denied(msg)
	}

	return admission.Allowed("")
}

func (w *OpenDataHubWebhook) checkDupCreation(ctx context.Context, req admission.Request) admission.Response {
	switch req.Kind.Kind {
	case "DataScienceCluster", "DSCInitialization":
	default:
		log.Info("Got wrong kind", "kind", req.Kind.Kind)
		return admission.Errored(http.StatusBadRequest, nil)
	}

	gvk := schema.GroupVersionKind{
		Group:   req.Kind.Group,
		Version: req.Kind.Version,
		Kind:    req.Kind.Kind,
	}

	// if count == 1 now creation of #2 is being handled
	return denyCountGtZero(ctx, w.client, gvk,
		fmt.Sprintf("Only one instance of %s object is allowed", req.Kind.Kind))
}

func (w *OpenDataHubWebhook) checkDeletion(ctx context.Context, req admission.Request) admission.Response {
	if req.Kind.Kind == "DataScienceCluster" {
		return admission.Allowed("")
	}

	// Restrict deletion of DSCI if DSC exists
	return denyCountGtZero(ctx, w.client, gvk.DataScienceCluster,
		fmt.Sprintln("Cannot delete DSCI object when DSC object still exists"))
}

func (w *OpenDataHubWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	var resp admission.Response

	switch req.Operation {
	case admissionv1.Create:
		resp = w.checkDupCreation(ctx, req)
	case admissionv1.Delete:
		resp = w.checkDeletion(ctx, req)
	default:
		msg := fmt.Sprintf("No logic check by webhook is applied on %v request", req.Operation)
		log.Info(msg)
		resp = admission.Allowed("")
	}

	if !resp.Allowed {
		return resp
	}

	return admission.Allowed(fmt.Sprintf("Operation %s on %s allowed", req.Operation, req.Kind.Kind))
}

func WaitForWebhookReady(ctx context.Context, interval int, timeout int) error {
	checkInterval := time.Duration(interval) * time.Second
	checkDuration := time.Duration(timeout) * time.Second
	webhookURL := "http://localhost:9443/webhookreadyz" // TODO: http: TLS handshake error from [::1]:48116: remote error: tls: bad certificate
	// operatorNs, err := cluster.GetOperatorNamespace()
	// if err != nil {
	// 	return err
	// }
	// webhookURL := "https://opendatahub-operator-controller-manager-service." + operatorNs + ".svc.cluster.local:9443" + endpoint

	err := wait.PollUntilContextTimeout(ctx, checkInterval, checkDuration, false, func(ctx context.Context) (bool, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, webhookURL, nil)
		if err != nil {
			return false, nil // Retry on error
		}
		fmt.Println("DEBUG WEN1")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("DEBUG WEN2")
			return false, nil // Retry on error
		}

		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			fmt.Println("DEBUG WEN3")
			return true, nil
		}

		fmt.Println("DEBUG WEN4 Status Code:", resp.StatusCode)
		return false, nil
	})
	return err
}

// hookServer := mgr.GetWebhookServer()
// hookServer.StartedChecker()(req)

// opendatahub-operator-webhook-service.openshift-operators.svc.cluster.local:443
// https://opendatahub-operator-controller-manager-service.openshift-operators.svc.cluster.local:443/validate-opendatahub-io-v1
// opendatahub-operator-controller-manager-service.openshift-operators.svc:443//validate-opendatahub-io-v1
