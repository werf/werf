package docker

import (
	"github.com/docker/docker/api/types/filters"
	"golang.org/x/net/context"

	"github.com/werf/werf/v2/pkg/container_backend/prune"
)

func VolumeRm(ctx context.Context, volumeName string, force bool) error {
	return apiCli(ctx).VolumeRemove(ctx, volumeName, force)
}

type (
	VolumesPruneOptions prune.Options
	VolumesPruneReport  prune.Report
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
