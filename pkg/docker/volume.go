package docker

import (
	"github.com/moby/moby/client"
	"golang.org/x/net/context"

	"github.com/werf/werf/v3/pkg/container_backend/prune"
)

func VolumeRm(ctx context.Context, volumeName string, force bool) error {
	_, err := apiCli(ctx).VolumeRemove(ctx, volumeName, client.VolumeRemoveOptions{Force: force})
	return err
}

type (
	VolumesPruneOptions prune.Options
	VolumesPruneReport  prune.Report
)

func VolumesPrune(ctx context.Context, _ VolumesPruneOptions) (VolumesPruneReport, error) {
	result, err := apiCli(ctx).VolumePrune(ctx, client.VolumePruneOptions{})
	if err != nil {
		return VolumesPruneReport{}, err
	}
	return VolumesPruneReport{
		ItemsDeleted:   result.Report.VolumesDeleted,
		SpaceReclaimed: result.Report.SpaceReclaimed,
	}, err
}
