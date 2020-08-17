package docker

import (
	"golang.org/x/net/context"
)

func VolumeRm(ctx context.Context, volumeName string, force bool) error {
	apiClient, err := apiCli()
	if err != nil {
		return err
	}

	return apiClient.VolumeRemove(ctx, volumeName, force)
}
