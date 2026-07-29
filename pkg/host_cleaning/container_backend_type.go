package host_cleaning

import "github.com/werf/werf/v2/pkg/container_backend"

//go:generate enumer -type=containerBackendType -trimprefix=containerBackend

type containerBackendType uint8

const (
	containerBackendDocker containerBackendType = iota
	containerBackendBuildah
	containerBackendTest
)

func resolveContainerBackendType(backend container_backend.ContainerBackend) (containerBackendType, error) {
	switch backend.(type) {
	case *container_backend.DockerServerBackend:
		return containerBackendDocker, nil
	case *container_backend.BuildahBackend:
		return containerBackendBuildah, nil
	default:
		// returns test type for testing with mock
		return containerBackendTest, ErrUnsupportedContainerBackend
	}
}
