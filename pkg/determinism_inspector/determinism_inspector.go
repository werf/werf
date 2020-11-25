package determinism_inspector

import (
	"context"

	"github.com/werf/logboek"
)

var (
	ReportedUncommittedPaths []string
)

func ReportUncommittedFile(ctx context.Context, path string) {
	for _, p := range ReportedUncommittedPaths {
		if p == path {
			return
		}
	}
	ReportedUncommittedPaths = append(ReportedUncommittedPaths, path)

	logboek.Context(ctx).Warn().LogF("WARNING: Uncommitted file %s was not taken into account due to enabled determinism mode (more info https://werf.io/documentation/advanced/configuration/determinism.html)\n", path)
}

func PrintInspectionDebrief(ctx context.Context) {
	if len(ReportedUncommittedPaths) > 0 {
		logboek.Context(ctx).Warn().LogF("\n### Determinism inspection debrief ###\n\n")
		logboek.Context(ctx).Warn().LogF("Following uncommitted files were not taken into account due to enabled determinism mode:\n")
		for _, path := range ReportedUncommittedPaths {
			logboek.Context(ctx).Warn().LogF(" - %s\n", path)
		}
		logboek.Context(ctx).Warn().LogLn()
		logboek.Context(ctx).Warn().LogF("More info about determinism in the werf avaiable on the page: https://werf.io/documentation/advanced/configuration/determinism.html\n")
	}
}
