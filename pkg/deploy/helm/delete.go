package helm

import (
	"context"

	"k8s.io/helm/pkg/proto/hapi/services"
)

type DeleteOptions struct {
	DisableHooks bool
	Purge        bool
	Timeout      int64
}

func Delete(releaseName string, opts DeleteOptions) error {
	_, err := tillerReleaseServer.UninstallRelease(context.Background(), &services.UninstallReleaseRequest{
		Name:         releaseName,
		DisableHooks: opts.DisableHooks,
		Purge:        opts.Purge,
		Timeout:      opts.Timeout,
	})
	return err
}
