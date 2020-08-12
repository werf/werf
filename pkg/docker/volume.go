package docker

import (
	"golang.org/x/net/context"
)

func VolumeRm(ctx context.Context, volumeName string, force bool) error {
	return apiClient.VolumeRemove(ctx, volumeName, force)
}
