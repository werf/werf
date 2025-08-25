package host_cleaning

import (
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
)

type cleanupReport prune.Report

func newCleanupReport() cleanupReport {
	return cleanupReport{
		ItemsDeleted:   []string{},
		SpaceReclaimed: 0,
	}
}

func mapPruneReportToCleanupReport(report prune.Report) cleanupReport {
	cr := newCleanupReport()
	cr.SpaceReclaimed = report.SpaceReclaimed
	cr.ItemsDeleted = append(cr.ItemsDeleted, report.ItemsDeleted...)
	return cr
}

func mapImageListToCleanupReport(list image.ImagesList) cleanupReport {
	report := newCleanupReport()
	report.ItemsDeleted = make([]string, 0, len(list))
	for _, img := range list {
		report.ItemsDeleted = append(report.ItemsDeleted, img.ID)
		report.SpaceReclaimed += uint64(img.Size)
	}
	return report
}
