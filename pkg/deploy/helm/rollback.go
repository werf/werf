package helm

import (
	"context"

	"k8s.io/helm/pkg/proto/hapi/services"
)

type RollbackOptions struct {
	DisableHooks  bool
	Recreate      bool
	Wait          bool
	Force         bool
	CleanupOnFail bool
	Timeout       int64
}

func Rollback(ctx context.Context, releaseName string, revision int32, opts RollbackOptions) error {
	_, err := tillerReleaseServer.RollbackRelease(context.Background(), &services.RollbackReleaseRequest{
		Name:              releaseName,
		Version:           revision,
		DisableHooks:      opts.DisableHooks,
		Recreate:          opts.Recreate,
		Wait:              opts.Wait,
		Force:             opts.Force,
		CleanupOnFail:     opts.CleanupOnFail,
		Timeout:           opts.Timeout,
		ThreeWayMergeMode: services.ThreeWayMergeMode_enabled,
	})
	return err

}
