package host_cleaning

import (
	"slices"

	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
)

type cleanupReport struct {
	ItemsDeleted []string
}

func (cr cleanupReport) Normalize() cleanupReport {
	if len(cr.ItemsDeleted) > 0 {
		return cleanupReport{
			ItemsDeleted: slices.Clip(cr.ItemsDeleted),
		}
	}
	return cleanupReport{}
}

func mapPruneReportToCleanupReport(report prune.Report) cleanupReport {
	return cleanupReport{
		ItemsDeleted: report.ItemsDeleted,
	}
}

func mapImageListToCleanupReport(list image.ImagesList) cleanupReport {
	itemsDeleted := make([]string, 0, len(list))

	for _, img := range list {
		itemsDeleted = append(itemsDeleted, img.ID)
	}

	report := cleanupReport{
		ItemsDeleted: itemsDeleted,
	}

	return report.Normalize()
}
