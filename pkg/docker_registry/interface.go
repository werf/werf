package docker_registry

import (
	"context"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type DockerRegistryInterface interface {
	GetRepoImageConfigFile(ctx context.Context, reference string) (*v1.ConfigFile, error)
}
