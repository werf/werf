package opstats

import (
	"context"
	"sort"
	"sync"
	"time"
)

type Operation string

const (
	OperationImagePull      Operation = "image pull"
	OperationImagePush      Operation = "image push"
	OperationImageBuild     Operation = "image build"
	OperationImageInfo      Operation = "image info"
	OperationImageSaveLoad  Operation = "image save/load"
	OperationImportChecksum Operation = "import checksum"
	OperationGitClone       Operation = "git clone"
	OperationGitFetch       Operation = "git fetch"
	OperationGitPatch       Operation = "git patch"
	OperationGitArchive     Operation = "git archive"
	OperationGitChecksum    Operation = "git checksum"
	OperationRegistryAPI    Operation = "registry API"
	OperationStageLockWait  Operation = "stage lock wait"
	OperationBuildContext   Operation = "build context"
)

type ctxKeyType struct{}

var ctxKey ctxKeyType

func NewContext(ctx context.Context, collector *Collector) context.Context {
	return context.WithValue(ctx, ctxKey, collector)
}

func FromContext(ctx context.Context) *Collector {
	collector, _ := ctx.Value(ctxKey).(*Collector)
	return collector
}

// Observe starts measuring an operation and returns a function that records the
// measurement into the collector bound to ctx. When no collector is bound,
// it is a no-op. Usage: defer opstats.Observe(ctx, opstats.OperationImagePull)()
func Observe(ctx context.Context, op Operation) func() {
	collector := FromContext(ctx)
	if collector == nil {
		return func() {}
	}

	start := time.Now()
	return func() {
		collector.add(op, start, time.Now())
	}
}

type Collector struct {
	mu        sync.Mutex
	intervals map[Operation][]interval
}

type interval struct {
	start time.Time
	end   time.Time
}

func NewCollector() *Collector {
	return &Collector{intervals: make(map[Operation][]interval)}
}

func (c *Collector) add(op Operation, start, end time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.intervals[op] = append(c.intervals[op], interval{start: start, end: end})
}

type OperationSummary struct {
	Operation Operation
	Count     int
	WallTime  time.Duration
}

// Summary returns per-operation stats sorted by wall time in descending order.
// WallTime is the union of possibly overlapping intervals measured across
// parallel workers, so it never exceeds the real elapsed time.
func (c *Collector) Summary() []OperationSummary {
	c.mu.Lock()
	defer c.mu.Unlock()

	res := make([]OperationSummary, 0, len(c.intervals))
	for op, intervals := range c.intervals {
		res = append(res, OperationSummary{
			Operation: op,
			Count:     len(intervals),
			WallTime:  unionDuration(intervals),
		})
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].WallTime == res[j].WallTime {
			return res[i].Operation < res[j].Operation
		}
		return res[i].WallTime > res[j].WallTime
	})

	return res
}

func unionDuration(intervals []interval) time.Duration {
	if len(intervals) == 0 {
		return 0
	}

	sorted := make([]interval, len(intervals))
	copy(sorted, intervals)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].start.Before(sorted[j].start) })

	var total time.Duration
	cur := sorted[0]
	for _, iv := range sorted[1:] {
		if iv.start.After(cur.end) {
			total += cur.end.Sub(cur.start)
			cur = iv
			continue
		}
		if iv.end.After(cur.end) {
			cur.end = iv.end
		}
	}
	total += cur.end.Sub(cur.start)

	return total
}
