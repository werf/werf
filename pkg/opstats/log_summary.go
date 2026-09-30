package opstats

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/werf/logboek"
	stylePkg "github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
)

// LogSummary prints the operations, stage cache and registry cache summary blocks. The timeLabel
// names the elapsed scope (e.g. "build time", "command time"). No-op when the collector is nil.
func LogSummary(ctx context.Context, collector *Collector, timeLabel string, elapsed time.Duration) {
	if collector == nil {
		return
	}

	summary := collector.Summary()
	if len(summary) > 0 {
		logboek.Context(ctx).LogBlock("Operations summary").
			Options(func(options types.LogBlockOptionsInterface) {
				options.Style(stylePkg.Highlight())
			}).
			Do(func() {
				for _, s := range summary {
					var parallelism string
					if s.WallTime > 0 && s.TotalTime > s.WallTime {
						parallelism = fmt.Sprintf("   ×%.1f", float64(s.TotalTime)/float64(s.WallTime))
					}
					logboek.Context(ctx).LogFHighlight("- %-32s %5d op   total %9.2fs   wall %9.2fs   avg %8.3fs   max %8.3fs%s\n",
						s.Operation, s.Count, s.TotalTime.Seconds(), s.WallTime.Seconds(), s.AvgTime.Seconds(), s.MaxTime.Seconds(), parallelism)
				}
				logboek.Context(ctx).LogFHighlight("%s: %.2fs (wall must not exceed it; total may)\n", timeLabel, elapsed.Seconds())
			})
	}

	registryEvents, stageEvents := lo.FilterReject(collector.EventSummary(), func(e EventSummary, _ int) bool {
		return IsRegistryEvent(ctx, e.Event)
	})
	logEventsBlock(ctx, "Stage cache summary", "stage(s)", stageEvents)
	logEventsBlock(ctx, "Registry cache summary", "request(s)", registryEvents)
}

func logEventsBlock(ctx context.Context, title, unit string, events []EventSummary) {
	if len(events) == 0 {
		return
	}

	logboek.Context(ctx).LogBlock(title).
		Options(func(options types.LogBlockOptionsInterface) {
			options.Style(stylePkg.Highlight())
		}).
		Do(func() {
			var total int
			for _, e := range events {
				total += e.Count
				logboek.Context(ctx).LogFHighlight("- %-30s %5d %s\n", e.Event, e.Count, unit)
			}
			logboek.Context(ctx).LogFHighlight("total: %d %s\n", total, unit)
		})
}
