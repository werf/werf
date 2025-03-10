package host_cleaning

import "github.com/werf/werf/v2/pkg/container_backend/prune"

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
