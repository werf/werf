package deploy

import (
	"context"

	"github.com/werf/werf/pkg/deploy/helm"
)

type InitOptions struct {
	HelmInitOptions helm.InitOptions
	WithoutHelm     bool
}

func Init(ctx context.Context, options InitOptions) error {
	if !options.WithoutHelm {
		if err := helm.Init(ctx, options.HelmInitOptions); err != nil {
			return err
		}
	}

	return nil
}
