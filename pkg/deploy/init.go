package deploy

import "github.com/werf/werf/pkg/deploy/helm"

type InitOptions struct {
	HelmInitOptions helm.InitOptions
	WithoutHelm     bool
}

func Init(options InitOptions) error {
	if !options.WithoutHelm {
		if err := helm.Init(options.HelmInitOptions); err != nil {
			return err
		}
	}

	return nil
}
