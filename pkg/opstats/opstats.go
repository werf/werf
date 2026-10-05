package opstats

import (
	"context"
	"io"
	"maps"
	"sort"
	"sync"
	"time"
)

type Operation string

const (
	OperationImagePull               Operation = "image pull"
	OperationImagePush               Operation = "image push"
	OperationStageBuild              Operation = "stage build"
	OperationImageInspect            Operation = "local image inspect"
	OperationImageSaveLoad           Operation = "image save/load"
	OperationStapelContainer         Operation = "stapel container prepare"
	OperationStapelContainerLockWait Operation = "stapel container lock wait"
	OperationGitClone                Operation = "git: clone"
	OperationGitFetch                Operation = "git: fetch"
	OperationGitLsRemote             Operation = "git: ls-remote"
	OperationGitPatch                Operation = "git: patch"
	OperationGitArchive              Operation = "git: archive"
	OperationGitChecksum             Operation = "git: checksum"
	OperationStageDigestLockWait     Operation = "stage lock wait (parallel tasks)"
	OperationContextAddFiles         Operation = "context add files"
	OperationConfigRender            Operation = "config render"
	OperationGiterminismInit         Operation = "giterminism init"

	// OperationStageLockWait measures acquiring a stage lock in the storage lock manager,
	// not holding or releasing it.
	OperationStageLockWait Operation = "sync: lock acquire"

	OperationRegistryTagsList Operation = "registry: tags list"
	OperationDockerImageList  Operation = "docker: image list"
	OperationBuildahImageList Operation = "buildah: image list"
)

// CacheLayer names the actual caching layer a lookup went through. The enum is
// closed: one row per real layer, so an unrecognized value records nothing
// rather than inventing a layer that does not exist.
type CacheLayer string

const (
	CacheLayerMemory CacheLayer = "memory"
	CacheLayerDisk   CacheLayer = "disk"
)

func (l CacheLayer) order() (int, bool) {
	switch l {
	case CacheLayerMemory:
		return 0, true
	case CacheLayerDisk:
		return 1, true
	}
	return 0, false
}

// CacheOutcome classifies a single completed call through a caching layer.
// Exactly one outcome is recorded per call.
type CacheOutcome string

const (
	// CacheOutcomeHit is a usable cached result, including a cached empty one.
	CacheOutcomeHit CacheOutcome = "hit"
	// CacheOutcomeMiss is a lookup the cache could not satisfy, so the underlying call ran.
	CacheOutcomeMiss CacheOutcome = "miss"
	// CacheOutcomeBypass is a call that asked for a fresh result and never consulted the cache.
	CacheOutcomeBypass CacheOutcome = "bypass"
)

type Event string

const (
	EventStageCacheHitLocal       Event = "found in local stages storage"
	EventStageCacheHitRepo        Event = "found in repo stages storage"
	EventStageCacheHitSecondary   Event = "copied from secondary storage"
	EventStageBuilt               Event = "built"
	EventRegistryTagsCacheHit     Event = "registry tags cache hit"
	EventRegistryTagsSharedResult Event = "registry tags shared result"

	// EventStageDiscarded counts stages that were built locally and then thrown away
	// because another publisher had already published a suitable stage. The reused
	// stage is counted as reused too, so this is a subset of the reused stages and
	// not an additional outcome.
	EventStageDiscarded Event = "discarded"

	// EventStageBroken counts stage reads, fetches and mutations that the stages
	// storage rejected as a broken image. A stage that is missing, rejected or
	// unavailable is not broken.
	EventStageBroken Event = "broken stage detections"

	// EventConveyorRestart counts conveyor attempts after the first one, i.e. the
	// restarts caused by an unexpected stages storage state.
	EventConveyorRestart Event = "conveyor restarts"
)

// IsRegistryEvent reports whether the event counts registry API requests rather
// than stages; the build report and the console summary keep the two apart.
func IsRegistryEvent(ctx context.Context, event Event) bool {
	return event == EventRegistryTagsCacheHit || event == EventRegistryTagsSharedResult
}

// IsRecoveryEvent reports whether the event counts recovering from a broken or
// conflicting storage state rather than how a stage or a registry request was
// satisfied; the build report and the console summary keep those apart.
func IsRecoveryEvent(ctx context.Context, event Event) bool {
	return event == EventStageBroken || event == EventConveyorRestart
}

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
// measurement into the collector bound to ctx. The returned function records at
// most once, so it is safe to both call it early and defer it. When no
// collector is bound, it is a no-op. Usage: defer opstats.Observe(ctx, opstats.OperationImagePull)()
func Observe(ctx context.Context, op Operation) func() {
	collector := FromContext(ctx)
	if collector == nil {
		return func() {}
	}

	start := time.Now()
	var once sync.Once
	return func() {
		once.Do(func() {
			collector.add(op, start, time.Now())
		})
	}
}

// CountEvent increments the event counter in the collector bound to ctx.
// When no collector is bound, it is a no-op.
func CountEvent(ctx context.Context, event Event) {
	collector := FromContext(ctx)
	if collector == nil {
		return
	}

	collector.mu.Lock()
	defer collector.mu.Unlock()
	collector.events[event]++
}

var _ io.ReadCloser = (*observedReadCloser)(nil)

type observedReadCloser struct {
	io.ReadCloser
	done func()
}

func (r *observedReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if err != nil {
		r.done()
	}
	return n, err
}

func (r *observedReadCloser) Close() error {
	defer r.done()
	return r.ReadCloser.Close()
}

// NewObservedReadCloser extends an in-flight observation over the consumption
// of a stream: done fires on first read error (including io.EOF) or once Close
// has completed, whichever comes first.
func NewObservedReadCloser(rc io.ReadCloser, done func()) io.ReadCloser {
	return &observedReadCloser{ReadCloser: rc, done: done}
}

// CountCacheLookup records one completed call through a caching layer: how the
// cache answered it and whether the call joined an already in-flight request
// instead of starting one. Call it exactly once per call, when the layer call
// returns, so that an outcome and its shared flag never end up in different
// reports. A hit answers from the cache without any request to join, so it is
// never shared. Each layer of a multi-layer cache counts its own lookup: a
// memory hit leaves the disk row untouched, while a memory miss answered from
// disk is one lookup on each of the two layers and not two lookups overall.
// No-op when no collector is bound to ctx, when the operation is empty, which
// is how an unrecognized backend avoids an invented row, or when the layer is
// not one of the known ones.
func CountCacheLookup(ctx context.Context, op Operation, layer CacheLayer, outcome CacheOutcome, shared bool) {
	collector := FromContext(ctx)
	if collector == nil || op == "" {
		return
	}
	if _, known := layer.order(); !known {
		return
	}

	collector.mu.Lock()
	defer collector.mu.Unlock()

	key := cacheKey{Operation: op, Layer: layer}
	counters := collector.cacheCounts[key]
	switch outcome {
	case CacheOutcomeHit:
		counters.hit++
		shared = false
	case CacheOutcomeMiss:
		counters.miss++
	case CacheOutcomeBypass:
		counters.bypass++
	default:
		return
	}
	if shared {
		counters.shared++
	}
	collector.cacheCounts[key] = counters
}

// cacheKey identifies one row of the cache summary. The operation and the layer
// stay separate fields so that neither can be confused with the other through
// an encoded separator.
type cacheKey struct {
	Operation Operation
	Layer     CacheLayer
}

type Collector struct {
	mu            sync.Mutex
	intervals     map[Operation][]interval
	events        map[Event]int
	cacheCounts   map[cacheKey]cacheCounters
	flushedOps    map[Operation]int
	flushedEvents map[Event]int
	flushedCache  map[cacheKey]cacheCounters
	// pendingCache is the cache-counter checkpoint captured by the last
	// PendingCacheSummary; CommitFlush advances to it and not to whatever has been
	// counted since, so observations made while the report was written survive.
	// It is keyed per layer too, so flushing one layer never drops the deltas of
	// another layer of the same operation.
	pendingCache map[cacheKey]cacheCounters
}

type cacheCounters struct {
	hit    int
	miss   int
	bypass int
	shared int
}

type interval struct {
	start time.Time
	end   time.Time
}

func NewCollector() *Collector {
	return &Collector{
		intervals:   make(map[Operation][]interval),
		events:      make(map[Event]int),
		cacheCounts: make(map[cacheKey]cacheCounters),
	}
}

func (c *Collector) add(op Operation, start, end time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.intervals[op] = append(c.intervals[op], interval{start: start, end: end})
}

type OperationSummary struct {
	Operation Operation
	Count     int
	TotalTime time.Duration
	WallTime  time.Duration
	AvgTime   time.Duration
	MaxTime   time.Duration
}

// Summary returns per-operation stats sorted by total time in descending
// order. WallTime is the union of possibly overlapping intervals measured
// across parallel workers, so it never exceeds the real elapsed time, while
// TotalTime sums all intervals and may exceed it.
func (c *Collector) Summary() []OperationSummary {
	c.mu.Lock()
	defer c.mu.Unlock()

	return summarizeOperations(c.intervals, nil)
}

// PendingSummary returns the same stats as Summary but only for the intervals
// recorded since the last CommitFlush, without advancing the flush mark. The
// build report uses the pending/commit pair so that a report covers only the
// build that wrote it (e.g. across --follow iterations) and a failed report
// write does not lose the pending observations.
func (c *Collector) PendingSummary(ctx context.Context) []OperationSummary {
	c.mu.Lock()
	defer c.mu.Unlock()

	return summarizeOperations(c.intervals, c.flushedOps)
}

func summarizeOperations(intervalsByOp map[Operation][]interval, skipByOp map[Operation]int) []OperationSummary {
	res := make([]OperationSummary, 0, len(intervalsByOp))
	for op, intervals := range intervalsByOp {
		intervals = intervals[skipByOp[op]:]
		if len(intervals) == 0 {
			continue
		}

		var total, max time.Duration
		for _, iv := range intervals {
			d := iv.end.Sub(iv.start)
			total += d
			if d > max {
				max = d
			}
		}

		res = append(res, OperationSummary{
			Operation: op,
			Count:     len(intervals),
			TotalTime: total,
			WallTime:  unionDuration(intervals),
			AvgTime:   total / time.Duration(len(intervals)),
			MaxTime:   max,
		})
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].TotalTime == res[j].TotalTime {
			return res[i].Operation < res[j].Operation
		}
		return res[i].TotalTime > res[j].TotalTime
	})

	return res
}

type EventSummary struct {
	Event Event
	Count int
}

// EventSummary returns per-event counters sorted by count in descending order.
func (c *Collector) EventSummary() []EventSummary {
	c.mu.Lock()
	defer c.mu.Unlock()

	return summarizeEvents(c.events, nil)
}

// PendingEventSummary returns the counters accumulated since the last
// CommitFlush without advancing the flush mark, mirroring PendingSummary.
func (c *Collector) PendingEventSummary(ctx context.Context) []EventSummary {
	c.mu.Lock()
	defer c.mu.Unlock()

	return summarizeEvents(c.events, c.flushedEvents)
}

type CacheSummary struct {
	Operation Operation
	Layer     CacheLayer
	Hit       int
	Miss      int
	Bypass    int
	Shared    int
}

func (s CacheSummary) lookups() int {
	return s.Hit + s.Miss + s.Bypass
}

// CacheSummary returns the cache counters of every operation and layer that
// actually received a lookup, sorted by operation and then by layer in lookup
// order, so that the layers of one operation stay together and read as the
// progression of a single lookup.
func (c *Collector) CacheSummary(ctx context.Context) []CacheSummary {
	c.mu.Lock()
	defer c.mu.Unlock()

	return summarizeCache(c.cacheCounts, nil)
}

// PendingCacheSummary returns the cache counters accumulated since the last
// CommitFlush and records the exact checkpoint that CommitFlush will advance to.
func (c *Collector) PendingCacheSummary(ctx context.Context) []CacheSummary {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pendingCache = make(map[cacheKey]cacheCounters, len(c.cacheCounts))
	maps.Copy(c.pendingCache, c.cacheCounts)

	return summarizeCache(c.cacheCounts, c.flushedCache)
}

func summarizeCache(counts, skipCounts map[cacheKey]cacheCounters) []CacheSummary {
	res := make([]CacheSummary, 0, len(counts))
	for key, counters := range counts {
		skip := skipCounts[key]
		s := CacheSummary{
			Operation: key.Operation,
			Layer:     key.Layer,
			Hit:       counters.hit - skip.hit,
			Miss:      counters.miss - skip.miss,
			Bypass:    counters.bypass - skip.bypass,
			Shared:    counters.shared - skip.shared,
		}
		if s.lookups() == 0 {
			continue
		}
		res = append(res, s)
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Operation != res[j].Operation {
			return res[i].Operation < res[j].Operation
		}
		li, _ := res[i].Layer.order()
		lj, _ := res[j].Layer.order()
		return li < lj
	})

	return res
}

// CommitFlush advances the flush mark past everything recorded so far, so the
// next Pending* calls return only later observations. Call it after the report
// consuming the pending summaries has been successfully delivered.
func (c *Collector) CommitFlush(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pendingCache != nil {
		c.flushedCache, c.pendingCache = c.pendingCache, nil
	}

	if c.flushedOps == nil {
		c.flushedOps = make(map[Operation]int)
	}
	for op, intervals := range c.intervals {
		c.flushedOps[op] = len(intervals)
	}

	if c.flushedEvents == nil {
		c.flushedEvents = make(map[Event]int)
	}
	for event, count := range c.events {
		c.flushedEvents[event] = count
	}
}

func summarizeEvents(counts, skipCounts map[Event]int) []EventSummary {
	res := make([]EventSummary, 0, len(counts))
	for event, count := range counts {
		count -= skipCounts[event]
		if count == 0 {
			continue
		}
		res = append(res, EventSummary{Event: event, Count: count})
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Count == res[j].Count {
			return res[i].Event < res[j].Event
		}
		return res[i].Count > res[j].Count
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
