package host_cleaning

import (
	"slices"

	"github.com/werf/werf/v2/pkg/cleanup_report"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
)

type backendPruneReport prune.Report

func (r backendPruneReport) Normalize() backendPruneReport {
	if len(r.ItemsDeleted) > 0 {
		return backendPruneReport{
			ItemsDeleted:   slices.Clip(r.ItemsDeleted),
			SpaceReclaimed: r.SpaceReclaimed,
		}
	}
	return backendPruneReport{}
}

func newBackendPruneReport(report prune.Report) backendPruneReport {
	return backendPruneReport{
		ItemsDeleted:   report.ItemsDeleted,
		SpaceReclaimed: report.SpaceReclaimed,
	}
}

func newBackendPruneReportFromImageList(list image.ImagesList) backendPruneReport {
	itemsDeleted := make([]string, 0, len(list))
	var spaceReclaimed uint64

	for _, img := range list {
		itemsDeleted = append(itemsDeleted, img.ID)
		spaceReclaimed += uint64(img.Size)
	}

	report := backendPruneReport{
		ItemsDeleted:   itemsDeleted,
		SpaceReclaimed: spaceReclaimed,
	}

	return report.Normalize()
}

func recordBackendPruneReport(hostReport *cleanup_report.HostReport, itemType cleanup_report.ItemType, report backendPruneReport) {
	items := make([]cleanup_report.Item, 0, len(report.ItemsDeleted))
	for _, id := range report.ItemsDeleted {
		items = append(items, cleanup_report.Item{Type: itemType, ID: id})
	}

	hostReport.AddDeleted(items...)
	hostReport.AddSpaceReclaimed(report.SpaceReclaimed)
}
