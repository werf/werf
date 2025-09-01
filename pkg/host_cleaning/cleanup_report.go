package host_cleaning

import (
	"slices"

	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
)

type cleanupReport prune.Report

func (cr cleanupReport) Normalize() cleanupReport {
	if len(cr.ItemsDeleted) > 0 {
		return cleanupReport{
			ItemsDeleted:   slices.Clip(cr.ItemsDeleted),
			SpaceReclaimed: cr.SpaceReclaimed,
		}
	}
	return cleanupReport{}
}

func mapPruneReportToCleanupReport(report prune.Report) cleanupReport {
	return cleanupReport{
		ItemsDeleted:   report.ItemsDeleted,
		SpaceReclaimed: report.SpaceReclaimed,
	}
}

func mapImageListToCleanupReport(list image.ImagesList) cleanupReport {
	itemsDeleted := make([]string, 0, len(list))
	var spaceReclaimed uint64

	for _, img := range list {
		itemsDeleted = append(itemsDeleted, img.ID)
		spaceReclaimed += uint64(img.Size)
	}

	report := cleanupReport{
		ItemsDeleted:   itemsDeleted,
		SpaceReclaimed: spaceReclaimed,
	}

	return report.Normalize()
}
