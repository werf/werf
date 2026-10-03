package opstats

import (
	"context"
	"fmt"
	"time"

	"github.com/werf/logboek"
	stylePkg "github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
)

// Column widths of both tables. The operation column is shared so that the two
// tables line up with each other.
const (
	opColumnFormat    = "%-32s"
	timeColumnFormat  = "%10s"
	countColumnFormat = "%7s"
)

// Cache summary row: operation, cache layer and the six counters, separated by a
// single space, 72 visible characters wide. A longer operation name or a counter
// wider than its column widens the row instead of being clipped.
const cacheRowFormat = "%-24s %-6s %7s %5s %5s %6s %6s %6s"

// LogSummary prints the operations table, the cache summary table and the
// stages line. The timeLabel and elapsed parameters are kept for call-site
// compatibility and are not rendered: the command already reports its own
// running time. No-op when the collector is nil.
func LogSummary(ctx context.Context, collector *Collector, _ string, _ time.Duration) {
	if collector == nil {
		return
	}

	logOperationsTable(ctx, collector.Summary())
	logCacheTable(ctx, collector.CacheSummary(ctx))
	logStagesAndRecoveryLines(ctx, collector.EventSummary())
}

func logOperationsTable(ctx context.Context, summary []OperationSummary) {
	if len(summary) == 0 {
		return
	}

	logBlock(ctx, "Operations", func() {
		logRow(ctx, opColumnFormat+timeColumnFormat+timeColumnFormat+timeColumnFormat+timeColumnFormat+timeColumnFormat,
			"operation", "count", "sum", "union", "avg", "max")
		for _, s := range summary {
			var parallelism string
			if s.WallTime > 0 && s.TotalTime > s.WallTime {
				parallelism = fmt.Sprintf("   ×%.1f (sum/union)", float64(s.TotalTime)/float64(s.WallTime))
			}
			logRow(ctx, opColumnFormat+timeColumnFormat+timeColumnFormat+timeColumnFormat+timeColumnFormat+timeColumnFormat+"%s",
				string(s.Operation), fmt.Sprintf("%d", s.Count), seconds(s.TotalTime), seconds(s.WallTime),
				seconds(s.AvgTime), seconds(s.MaxTime), parallelism)
		}
		logboek.Context(ctx).LogFHighlight("sum: durations added; parallel calls counted separately\n")
		logboek.Context(ctx).LogFHighlight("union: time with ≥1 active call; overlapping intervals counted once\n")
	})
}

func logCacheTable(ctx context.Context, summary []CacheSummary) {
	if len(summary) == 0 {
		return
	}

	logBlock(ctx, "Cache summary", func() {
		logRow(ctx, cacheRowFormat,
			"operation", "cache", "lookups", "hit", "miss", "bypass", "shared", "hit%")
		for _, s := range summary {
			logRow(ctx, cacheRowFormat,
				string(s.Operation), string(s.Layer), fmt.Sprintf("%d", s.lookups()), fmt.Sprintf("%d", s.Hit),
				fmt.Sprintf("%d", s.Miss), fmt.Sprintf("%d", s.Bypass), fmt.Sprintf("%d", s.Shared), hitRate(s))
		}
		logboek.Context(ctx).LogFHighlight("lookups = hit + miss + bypass\n")
		logboek.Context(ctx).LogFHighlight("hit%% = hit / (hit + miss); bypass excluded\n")
		logboek.Context(ctx).LogFHighlight("shared: calls joining an in-flight request; included in miss or bypass\n")
	})
}

func logStagesAndRecoveryLines(ctx context.Context, events []EventSummary) {
	var reused, built, discarded, broken, restarts int
	var observed bool
	for _, e := range events {
		switch e.Event {
		case EventStageCacheHitLocal, EventStageCacheHitRepo, EventStageCacheHitSecondary:
			reused += e.Count
			observed = true
		case EventStageBuilt:
			built += e.Count
			observed = true
		case EventStageDiscarded:
			discarded += e.Count
			observed = true
		case EventStageBroken:
			broken += e.Count
		case EventConveyorRestart:
			restarts += e.Count
		}
	}

	// A command that did no stage work shows no line rather than an invented zero.
	if observed {
		logboek.Context(ctx).LogFHighlight("Stages: %d reused, %d built, %d discarded\n", reused, built, discarded)
		logboek.Context(ctx).LogFHighlight("discarded: built locally but another published stage was reused; included in reused\n")
	}

	// Recovery is about what went wrong, so a run that hit nothing shows no line.
	if broken > 0 || restarts > 0 {
		logboek.Context(ctx).LogFHighlight("Recovery: %d broken stage detections, %d conveyor restarts\n", broken, restarts)
	}
}

func hitRate(s CacheSummary) string {
	// Bypasses never consulted the cache, so with no hits and no misses the rate is
	// undefined rather than zero.
	if s.Hit+s.Miss == 0 {
		return "—"
	}
	return fmt.Sprintf("%.0f%%", 100*float64(s.Hit)/float64(s.Hit+s.Miss))
}

func seconds(d time.Duration) string {
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func logRow(ctx context.Context, format string, values ...string) {
	cells := make([]interface{}, 0, len(values))
	for _, v := range values {
		cells = append(cells, v)
	}
	logboek.Context(ctx).LogFHighlight(format+"\n", cells...)
}

func logBlock(ctx context.Context, title string, do func()) {
	logboek.Context(ctx).LogBlock(title).
		Options(func(options types.LogBlockOptionsInterface) {
			options.Style(stylePkg.Highlight())
		}).
		Do(do)
}
