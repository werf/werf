package docker

import (
	"context"

	"github.com/docker/docker/api/types/system"
)

func Info(ctx context.Context) (system.Info, error) {
	return apiCli(ctx).Info(ctx)
}
