package docker

import (
	"github.com/docker/docker/api/types/filters"
	"golang.org/x/net/context"
)

func VolumeRm(ctx context.Context, volumeName string, force bool) error {
	return apiCli(ctx).VolumeRemove(ctx, volumeName, force)
}

type (
	VolumesPruneOptions BuildCachePruneOptions
	VolumesPruneReport  BuildCachePruneReport
)

func VolumesPrune(ctx context.Context, _ VolumesPruneOptions) (VolumesPruneReport, error) {
	report, err := apiCli(ctx).VolumesPrune(ctx, filters.NewArgs())
	if err != nil {
		return VolumesPruneReport{}, err
	}
	return VolumesPruneReport{
		ItemsDeleted:   report.VolumesDeleted,
		SpaceReclaimed: report.SpaceReclaimed,
	}, err
}
