package deploy

import "github.com/flant/werf/pkg/deploy/helm"

func Init() error {
	if err := helm.ValidateHelmVersion(); err != nil {
		return err
	}

	return nil
}
