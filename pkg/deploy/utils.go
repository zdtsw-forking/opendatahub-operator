package deploy

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/conversion"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
)

func isSharedResource(componentCounter []string, componentName string) bool {
	return len(componentCounter) > 1 || (len(componentCounter) == 1 && componentCounter[0] != componentName)
}

func isOwnedByODHCRD(ownerReferences []metav1.OwnerReference) bool {
	for _, owner := range ownerReferences {
		if owner.Kind == "DataScienceCluster" || owner.Kind == "DSCInitialization" {
			return true
		}
	}
	return false
}

func getComponentCounter(foundLabels map[string]string) []string {
	var componentCounter []string
	for label := range foundLabels {
		if strings.Contains(label, labels.ODHAppPrefix) {
			compFound := strings.Split(label, "/")[1]
			componentCounter = append(componentCounter, compFound)
		}
	}
	return componentCounter
}

func Wen(ctx context.Context, cli client.Client, data map[string]string, owner metav1.Object, templatesFS embed.FS, tempFile string) error {

	content, err := os.ReadFile(tempFile)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("templatefile").
		Option("missingkey=error").
		Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	unstrObj, err := conversion.StrToUnstructured(buffer.String())
	if err != nil || len(unstrObj) != 1 {
		return err
	}
	return client.IgnoreAlreadyExists(cli.Create(ctx, unstrObj[0]))
}
