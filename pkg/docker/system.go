package docker

import (
	"context"

	"github.com/docker/docker/api/types"
)

func Info(ctx context.Context) (types.Info, error) {
	return apiCli(ctx).Info(ctx)
}
