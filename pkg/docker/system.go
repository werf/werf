package docker

import (
	"context"

	"github.com/docker/docker/api/types"
)

func Info(ctx context.Context) (types.Info, error) {
	return apiCli(ctx).Info(ctx)
}

// GetRegistryMirrors returns registry mirrors from Docker daemon.
func GetRegistryMirrors(ctx context.Context) ([]string, error) {
	if !IsEnabled() {
		return nil, nil
	}

	var info types.Info
	var err error

	if IsContext(ctx) {
		info, err = apiCli(ctx).Info(ctx)
	} else if defaultCLI != nil {
		info, err = defaultCLI.Client().Info(ctx)
	} else {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if info.RegistryConfig == nil {
		return nil, nil
	}

	return info.RegistryConfig.Mirrors, nil
}
