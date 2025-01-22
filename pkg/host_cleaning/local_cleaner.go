package host_cleaning

import (
	"context"
	"errors"

	"github.com/werf/werf/v2/pkg/container_backend"
)

var ErrUnsupportedContainerBackend = errors.New("unsupported container backend")

type LocalCleaner interface {
	RunGC(ctx context.Context, options RunGCOptions) error
	ShouldRunAutoGC(ctx context.Context, options RunAutoGCOptions) (bool, error)
	Name() string
}

type RunGCOptions struct {
	AllowedStorageVolumeUsagePercentage       float64
	AllowedStorageVolumeUsageMarginPercentage float64
	StoragePath                               string
	force                                     bool
	dryRun                                    bool
}

type RunAutoGCOptions struct {
	AllowedStorageVolumeUsagePercentage float64
	StoragePath                         string
}

func CreateCleaner(containerBackend container_backend.ContainerBackend) (LocalCleaner, error) {
	switch containerBackend.(type) {
	case *container_backend.DockerServerBackend:
		return NewLocalCleanerDocker(containerBackend), nil

	// TODO: implement buildah cleaner

	default:
		return nil, ErrUnsupportedContainerBackend
	}
}
