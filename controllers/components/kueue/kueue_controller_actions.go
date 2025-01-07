package kueue

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	componentApi "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1alpha1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	odhtypes "github.com/opendatahub-io/opendatahub-operator/v2/pkg/controller/types"
	odhdeploy "github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
)

func initialize(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	rr.Manifests = append(rr.Manifests, manifestsPath())
	// Add OCP 4.17+ specific manifests if Minor > 16 {
	ov, err := cluster.GetOCPVersion(ctx, rr.Client)
	if err != nil {
		return fmt.Errorf("failed to get OCP version: %w", err)
	}
	ctrl.Log.Info("OCP: ", "version", ov.String())
	if ov.Minor > 16 { // it is safe to have to check only on minor, v3 has EOL last version is 11
		ctrl.Log.Info("OCP version is over 4.16")
		rr.Manifests = append(rr.Manifests, extramanifestsPath())
	}

	return nil
}

func devFlags(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	kueue, ok := rr.Instance.(*componentApi.Kueue)
	if !ok {
		return fmt.Errorf("resource instance %v is not a componentApi.Kueue)", rr.Instance)
	}

	if kueue.Spec.DevFlags == nil {
		return nil
	}

	// Implement devflags support logic
	// If dev flags are set, update default manifests path
	if len(kueue.Spec.DevFlags.Manifests) != 0 {
		manifestConfig := kueue.Spec.DevFlags.Manifests[0]
		if err := odhdeploy.DownloadManifests(ctx, ComponentName, manifestConfig); err != nil {
			return err
		}

		if manifestConfig.SourcePath != "" {
			rr.Manifests[0].Path = odhdeploy.DefaultManifestPath
			rr.Manifests[0].ContextDir = ComponentName
			rr.Manifests[0].SourcePath = manifestConfig.SourcePath
		}
	}

	return nil
}
