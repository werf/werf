package docker

import (
	"context"

	"github.com/docker/docker/api/types"

	"github.com/werf/werf/v2/pkg/container_backend/prune"
)

type (
	BuildCachePruneOptions prune.Options
	BuildCachePruneReport  prune.Report
)

func BuildCachePrune(ctx context.Context, _ BuildCachePruneOptions) (BuildCachePruneReport, error) {
	report, err := apiCli(ctx).BuildCachePrune(ctx, types.BuildCachePruneOptions{})
	if err != nil {
		return BuildCachePruneReport{}, err
	}
	return BuildCachePruneReport{
		ItemsDeleted:   report.CachesDeleted,
		SpaceReclaimed: report.SpaceReclaimed,
	}, err
}
