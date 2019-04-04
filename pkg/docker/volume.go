package docker

import (
	"golang.org/x/net/context"
)

func VolumeRm(volumeName string, force bool) error {
	ctx := context.Background()
	return apiClient.VolumeRemove(ctx, volumeName, force)
}
