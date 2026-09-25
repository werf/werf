package common

import (
	"context"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v2/pkg/opstats"
)

// InitOperationsStatistics installs the operations statistics collector into the context for the
// whole command run when enabled by --build-report-operations or debug logging. The returned
// function prints the summary blocks and must be deferred by the command. It restores the
// accepted log level captured here: nelm actions lower it globally (e.g. render sets Error),
// which would otherwise silently drop the summary printed after them.
func InitOperationsStatistics(ctx context.Context, cmdData *CmdData) (context.Context, func()) {
	if !GetBuildReportOperations(cmdData) && !logboek.Context(ctx).IsAcceptedLevel(level.Debug) {
		return ctx, func() {}
	}

	collector := opstats.NewCollector()
	ctx = opstats.NewContext(ctx, collector)
	startedAt := time.Now()
	acceptedLevel := logboek.Context(ctx).AcceptedLevel()

	return ctx, func() {
		logboek.Context(ctx).SetAcceptedLevel(acceptedLevel)
		opstats.LogSummary(ctx, collector, "command time", time.Since(startedAt))
	}
}
