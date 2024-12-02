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

package modelcontroller

import (
	"context"
	"fmt"
	"strings"

	operatorv1 "github.com/openshift/api/operator/v1"

	componentsv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1"
	odhtypes "github.com/opendatahub-io/opendatahub-operator/v2/pkg/controller/types"
	odhdeploy "github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
)

func initialize(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	// early exist
	_, ok := rr.Instance.(*componentsv1.ModelController)
	if !ok {
		return fmt.Errorf("resource instance %v is not a componentsv1.ModelController)", rr.Instance)
	}
	rr.Manifests = append(rr.Manifests, odhtypes.ManifestInfo{
		Path:       odhdeploy.DefaultManifestPath,
		ContextDir: ComponentName,
		SourcePath: "base",
	})
	return nil
}

func devFlags(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	_, ok := rr.Instance.(*componentsv1.ModelController)
	if !ok {
		return fmt.Errorf("resource instance %v is not a componentsv1.ModelController)", rr.Instance)
	}
	// Get Kserve which can override Kserve devflags
	k := rr.DSC.Spec.Components.Kserve
	if k.ManagementSpec.ManagementState != operatorv1.Managed || k.DevFlags == nil || len(k.DevFlags.Manifests) == 0 {
		// Get ModelMeshServing if it is enabled and has devlfags
		mm := rr.DSC.Spec.Components.ModelMeshServing
		if mm.ManagementSpec.ManagementState != operatorv1.Managed || mm.DevFlags == nil || len(mm.DevFlags.Manifests) == 0 {
			return nil
		}

		for _, subcomponent := range rr.DSC.Spec.Components.ModelMeshServing.DevFlags.Manifests {
			if strings.Contains(subcomponent.URI, ComponentName) {
				// Download odh-model-controller
				if err := odhdeploy.DownloadManifests(ctx, ComponentName, subcomponent); err != nil {
					return err
				}
				// If overlay is defined, update paths
				if subcomponent.SourcePath != "" {
					rr.Manifests[0].SourcePath = subcomponent.SourcePath
				}
			}
		}
		return nil
	}

	for _, subcomponent := range rr.DSC.Spec.Components.Kserve.DevFlags.Manifests {
		if strings.Contains(subcomponent.URI, ComponentName) {
			// Download odh-model-controller
			if err := odhdeploy.DownloadManifests(ctx, ComponentName, subcomponent); err != nil {
				return err
			}
			// If overlay is defined, update paths
			if subcomponent.SourcePath != "" {
				rr.Manifests[0].SourcePath = subcomponent.SourcePath
			}
		}
	}
	return nil
}
