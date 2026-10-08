//go:build !linux

package host_cleaning

import (
	"context"
	"os"

	"github.com/werf/werf/v3/pkg/container_backend"
)

func removeProjectTmpDir(ctx context.Context, backend container_backend.ContainerBackend, path string) error {
	return os.RemoveAll(path)
}
