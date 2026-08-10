package build

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/werf"
)

// recordingPhase is a minimal Phase implementation that records, for each
// image it processes, the image's name (after an optional artificial delay)
// into a shared, mutex-protected order slice. It lets a test observe the
// actual build ORDER produced by Conveyor.doImages/doImagesInParallel without
// needing a real container backend.
type recordingPhase struct {
	mu     *sync.Mutex
	order  *[]string
	delays map[string]time.Duration
}

func (p *recordingPhase) Name() string                       { return "recording" }
func (p *recordingPhase) BeforeImages(context.Context) error { return nil }
func (p *recordingPhase) AfterImages(context.Context) error  { return nil }
func (p *recordingPhase) BeforeImageStages(context.Context, *image.Image) (func(), error) {
	return nil, nil
}

func (p *recordingPhase) OnImageStage(context.Context, *image.Image, stage.Interface) error {
	return nil
}

func (p *recordingPhase) AfterImageStages(ctx context.Context, img *image.Image) error {
	if d := p.delays[img.Name]; d > 0 {
		time.Sleep(d)
	}
	p.mu.Lock()
	*p.order = append(*p.order, img.Name)
	p.mu.Unlock()
	return nil
}

func (p *recordingPhase) ImageProcessingShouldBeStopped(context.Context, *image.Image) bool {
	return false
}

func (p *recordingPhase) Clone() Phase          { return p }
func (p *recordingPhase) Report() *ImagesReport { return nil }

// TestDoImagesInParallel_DependentImageDoesNotWaitForUnrelatedSlowImage is
// the end-to-end regression test for replacing wave/level scheduling with a
// real dependency graph: it drives the ACTUAL Conveyor.doImages /
// doImagesInParallel wiring (scheduler construction, parallel.DoTasksDynamic,
// scheduler.complete) rather than calling graphScheduler.next/.complete
// directly, over a hand-built graph with an independent slow image and a
// fast a -> b -> c chain. It asserts b/c build right after their real
// dependency finishes instead of waiting for the unrelated slow image.
func TestDoImagesInParallel_DependentImageDoesNotWaitForUnrelatedSlowImage(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), "")) // tmp_manager (used by parallel.NewWorker) requires werf init

	newImg := func(name string) *image.Image {
		img := &image.Image{Name: name, TargetPlatform: "linux/amd64"}
		img.ForceTargetPlatformLogging = true
		return img
	}

	slow := newImg("slow")
	a := newImg("a")
	b := newImg("b")
	c := newImg("c")
	b.AddDependencyName("a")
	c.AddDependencyName("b")

	graph, err := image.BuildImagesGraph([]*image.Image{slow, a, b, c})
	require.NoError(t, err)

	tree := &image.ImagesTree{}
	tree.SetImagesGraphForTests(graph)

	conveyor := &Conveyor{
		ConveyorOptions: ConveyorOptions{
			Parallel:           true,
			ParallelTasksLimit: -1,
		},
		imagesTree:       tree,
		stageImages:      make(map[string]*stage.StageImage),
		serviceRWMutex:   map[string]*sync.RWMutex{},
		stageDigestMutex: map[string]*sync.Mutex{},
	}

	var mu sync.Mutex
	var order []string
	phase := &recordingPhase{
		mu:    &mu,
		order: &order,
		delays: map[string]time.Duration{
			"slow": 150 * time.Millisecond,
			"a":    10 * time.Millisecond,
			"b":    10 * time.Millisecond,
			"c":    10 * time.Millisecond,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, conveyor.doImages(ctx, []Phase{phase}, false))

	indexOf := func(name string) int {
		mu.Lock()
		defer mu.Unlock()
		for i, n := range order {
			if n == name {
				return i
			}
		}
		t.Fatalf("image %q never completed; order=%v", name, order)
		return -1
	}

	require.Less(t, indexOf("a"), indexOf("b"), "b must build after a")
	require.Less(t, indexOf("b"), indexOf("c"), "c must build after b")

	// The core regression check: b/c must not be gated behind the unrelated,
	// slower "slow" image just because a graph-scheduling bug reintroduced a
	// wave/level barrier. With a 150ms artificial delay on "slow" and only
	// 10ms on a/b/c, both b and c finish well before "slow" does.
	require.Less(t, indexOf("c"), indexOf("slow"),
		"dependent chain a->b->c must not wait for unrelated image \"slow\"; observed order=%v", order)
}

// TestDoImagesInParallel_AssignsBuildOrderIndexByRealDequeueNotStaticTopology
// is the regression test for the build-log progress-numbering fix: images
// used to get their log progress index assigned once, statically, from their
// position in the dependency graph's topological order (ImagesTree.Calculate)
// — a position with no relation to the real, concurrent, dependency-driven
// order images actually start building in, producing confusingly
// out-of-order "(N/Total)" numbers in the build log. doImagesInParallel must
// instead assign each image's index at the moment it is actually dequeued for
// building.
//
// The graph is built with "slow" listed LAST (images passed to
// BuildImagesGraph as [a, b, c, slow]), which — because "slow" has no
// dependencies and is a DFS leaf reachable only after a/b/c in that input
// order — places it LAST (index 3) in the graph's static topological order
// (graph.Nodes()). If build-order numbering regressed back to that static
// position, "slow" would get index 3. But "slow" has no dependencies, so the
// real scheduler makes it ready to build from the very start, alongside "a"
// — it must get an early real build-order index (0 or 1), not the late
// static one. This is what actually distinguishes "assigned by real dequeue
// order" from "assigned by static topological position": a graph where a
// dependency-free image's real-time readiness and its topological-sort
// position disagree.
//
// The a < b < c assertion is a separate, always-true invariant (the shared
// build-order counter only increases, and graphScheduler guarantees b/c
// can't be dequeued before a/b's build has fully completed) kept here as a
// basic sanity check, not the core regression signal — a purely topological
// assignment would also happen to satisfy it for a simple chain.
func TestDoImagesInParallel_AssignsBuildOrderIndexByRealDequeueNotStaticTopology(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), ""))

	newImg := func(name string) *image.Image {
		img := &image.Image{Name: name, TargetPlatform: "linux/amd64"}
		img.ForceTargetPlatformLogging = true
		return img
	}

	slow := newImg("slow")
	a := newImg("a")
	b := newImg("b")
	c := newImg("c")
	b.AddDependencyName("a")
	c.AddDependencyName("b")

	// "slow" listed last on purpose — see the doc comment above.
	graph, err := image.BuildImagesGraph([]*image.Image{a, b, c, slow})
	require.NoError(t, err)

	tree := &image.ImagesTree{}
	tree.SetImagesGraphForTests(graph)

	conveyor := &Conveyor{
		ConveyorOptions: ConveyorOptions{
			Parallel:           true,
			ParallelTasksLimit: -1,
		},
		imagesTree:       tree,
		stageImages:      make(map[string]*stage.StageImage),
		serviceRWMutex:   map[string]*sync.RWMutex{},
		stageDigestMutex: map[string]*sync.Mutex{},
	}

	var mu sync.Mutex
	var order []string
	phase := &recordingPhase{
		mu:    &mu,
		order: &order,
		delays: map[string]time.Duration{
			"slow": 10 * time.Millisecond,
			"a":    10 * time.Millisecond,
			"b":    10 * time.Millisecond,
			"c":    10 * time.Millisecond,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, conveyor.doImages(ctx, []Phase{phase}, false))

	images := []*image.Image{slow, a, b, c}
	seen := make(map[int]string, len(images))
	for _, img := range images {
		idx := img.GetBuildOrderIndex()
		require.GreaterOrEqual(t, idx, 0)
		require.Less(t, idx, len(images))

		if other, dup := seen[idx]; dup {
			t.Fatalf("build-order index %d assigned to both %q and %q", idx, other, img.Name)
		}
		seen[idx] = img.Name
	}

	require.Less(t, a.GetBuildOrderIndex(), b.GetBuildOrderIndex(), "a must be assigned a build-order index before its dependent b")
	require.Less(t, b.GetBuildOrderIndex(), c.GetBuildOrderIndex(), "b must be assigned a build-order index before its dependent c")

	// The core regression signal: "slow" has no dependencies and is ready to
	// build from the very start (same as "a"), so it must get an early
	// build-order index. Under the old, reverted-to bug (static topological
	// index), "slow" would instead get index 3 — the last position, because
	// it was listed last in the input passed to BuildImagesGraph above.
	require.LessOrEqual(t, slow.GetBuildOrderIndex(), 1,
		"\"slow\" has no dependencies and must be dequeued near the start (index 0 or 1), not assigned the static topological tail position (3) it would get under the old bug; got %d", slow.GetBuildOrderIndex())
}

// TestDoImagesInParallel_AnnotatesEachImageWithARealWorkerID is the
// end-to-end test for the worker-ID log annotation: it exercises the real
// path — parallel.DoTasksDynamic stashing the worker ID into the task
// context, Conveyor.doImagesInParallel reading it back out via
// ctx.Value(parallel.CtxBackgroundTaskIDKey) and calling
// Image.SetWorkerID — rather than only unit-testing the string formatting
// in isolation (pkg/logging/image_ai_test.go). With 4 independent images
// and exactly 2 workers, both worker IDs must actually get used (not just
// a default/zero value), and every image must end up with a worker ID set.
func TestDoImagesInParallel_AnnotatesEachImageWithARealWorkerID(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), ""))

	newImg := func(name string) *image.Image {
		img := &image.Image{Name: name, TargetPlatform: "linux/amd64"}
		img.ForceTargetPlatformLogging = true
		return img
	}

	images := []*image.Image{newImg("w1"), newImg("w2"), newImg("w3"), newImg("w4")}

	graph, err := image.BuildImagesGraph([]*image.Image{images[0], images[1], images[2], images[3]})
	require.NoError(t, err)

	tree := &image.ImagesTree{}
	tree.SetImagesGraphForTests(graph)

	conveyor := &Conveyor{
		ConveyorOptions: ConveyorOptions{
			Parallel:           true,
			ParallelTasksLimit: 2,
		},
		imagesTree:       tree,
		stageImages:      make(map[string]*stage.StageImage),
		serviceRWMutex:   map[string]*sync.RWMutex{},
		stageDigestMutex: map[string]*sync.Mutex{},
	}

	var mu sync.Mutex
	var order []string
	phase := &recordingPhase{
		mu:    &mu,
		order: &order,
		delays: map[string]time.Duration{
			"w1": 20 * time.Millisecond,
			"w2": 20 * time.Millisecond,
			"w3": 20 * time.Millisecond,
			"w4": 20 * time.Millisecond,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, conveyor.doImages(ctx, []Phase{phase}, false))

	seenWorkers := map[int]bool{}
	for _, img := range images {
		workerID, ok := img.GetWorkerID()
		require.True(t, ok, "image %q must have a worker ID set by the real doImagesInParallel path", img.Name)
		require.True(t, workerID == 0 || workerID == 1, "worker ID for %q must be 0 or 1 (ParallelTasksLimit=2), got %d", img.Name, workerID)
		seenWorkers[workerID] = true
	}

	require.Len(t, seenWorkers, 2, "both workers must actually have been used across 4 images with ParallelTasksLimit=2, not just one")
}
