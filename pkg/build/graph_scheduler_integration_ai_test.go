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

	// startOrder, when non-nil, records the order in which images actually
	// begin building (BeforeImageStages runs before any artificial delay),
	// as opposed to order, which records completion order.
	startOrder *[]string
}

func (p *recordingPhase) Name() string                       { return "recording" }
func (p *recordingPhase) BeforeImages(context.Context) error { return nil }
func (p *recordingPhase) AfterImages(context.Context) error  { return nil }
func (p *recordingPhase) BeforeImageStages(_ context.Context, img *image.Image) (func(), error) {
	if p.startOrder != nil {
		p.mu.Lock()
		*p.startOrder = append(*p.startOrder, img.Name)
		p.mu.Unlock()
	}
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

// TestAI_DoImagesInParallel_DependentImageDoesNotWaitForUnrelatedSlowImage is
// the end-to-end regression test for replacing wave/level scheduling with a
// real dependency graph: it drives the ACTUAL Conveyor.doImages /
// doImagesInParallel wiring (scheduler construction, parallel.DoTasksDynamic,
// scheduler.complete) rather than calling graphScheduler.next/.complete
// directly, over a hand-built graph with an independent slow image and a
// fast a -> b -> c chain. It asserts b/c build right after their real
// dependency finishes instead of waiting for the unrelated slow image.
func TestAI_DoImagesInParallel_DependentImageDoesNotWaitForUnrelatedSlowImage(t *testing.T) {
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

// TestAI_DoImagesInParallel_AssignsBuildOrderIndexMatchingActualStartOrder is
// the regression test for the build-log progress-numbering fix: images used
// to get their log progress index assigned once, statically, from their
// position in the dependency graph's topological order (ImagesTree.Calculate)
// — a position that has no relation to the real, concurrent, dependency-
// driven order images actually start building in, producing confusingly
// out-of-order "(N/Total)" numbers in the build log. doImagesInParallel must
// instead assign each image's index at the moment it is actually dequeued for
// building, so the numbers form a 0..N-1 permutation matching real start
// order.
func TestAI_DoImagesInParallel_AssignsBuildOrderIndexMatchingActualStartOrder(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), ""))

	newImg := func(name string) *image.Image {
		img := &image.Image{Name: name, TargetPlatform: "linux/amd64"}
		img.ForceTargetPlatformLogging = true
		return img
	}

	// Same shape as the test above: an independent slow image plus a fast
	// a -> b -> c chain, so the real dequeue order is very unlikely to match
	// whatever static topological order ImagesTree.Calculate would assign.
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
	var order, startOrder []string
	phase := &recordingPhase{
		mu:         &mu,
		order:      &order,
		startOrder: &startOrder,
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
	require.Len(t, startOrder, 4)

	byName := map[string]*image.Image{"slow": slow, "a": a, "b": b, "c": c}
	seenIndex := map[int]string{}
	for wantIdx, name := range startOrder {
		img := byName[name]
		gotIdx := img.GetBuildOrderIndex()
		require.Equal(t, wantIdx, gotIdx, "image %q build-order index must match its real start position; startOrder=%v", name, startOrder)

		if other, dup := seenIndex[gotIdx]; dup {
			t.Fatalf("build-order index %d assigned to both %q and %q", gotIdx, other, name)
		}
		seenIndex[gotIdx] = name
	}
}
