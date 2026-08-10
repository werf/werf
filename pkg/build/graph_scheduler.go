package build

import (
	"context"
	"sync"

	"github.com/werf/werf/v2/pkg/build/image"
)

// graphScheduler drives image builds over an image.ImagesGraph dynamically:
// an image becomes eligible to build as soon as all of its own dependencies
// have finished building, instead of waiting for an entire unrelated batch
// ("wave"/level) of images to finish first.
type graphScheduler struct {
	graph *image.ImagesGraph
	nodes []*image.Image
	index map[*image.Image]int

	mu        sync.Mutex
	remaining map[*image.Image]int // number of not-yet-finished dependencies
	ready     chan int             // node indices ready to build
	done      chan struct{}        // closed once every node has completed
	completed int
}

func newGraphScheduler(graph *image.ImagesGraph) *graphScheduler {
	nodes := graph.Nodes()

	s := &graphScheduler{
		graph:     graph,
		nodes:     nodes,
		index:     make(map[*image.Image]int, len(nodes)),
		remaining: make(map[*image.Image]int, len(nodes)),
		ready:     make(chan int, len(nodes)),
		done:      make(chan struct{}),
	}

	for i, img := range nodes {
		s.index[img] = i
	}

	for _, img := range nodes {
		deps := len(graph.Dependencies(img))
		s.remaining[img] = deps
		if deps == 0 {
			s.ready <- s.index[img]
		}
	}

	if len(nodes) == 0 {
		close(s.done)
	}

	return s
}

// next implements parallel.NextTaskFunc: it blocks until either a node
// becomes ready, every node has completed, or ctx is done.
func (s *graphScheduler) next(ctx context.Context) (int, bool, error) {
	select {
	case taskId := <-s.ready:
		return taskId, true, nil
	case <-s.done:
		return 0, false, nil
	case <-ctx.Done():
		return 0, false, ctx.Err()
	}
}

// complete must be called exactly once after a node finishes building
// successfully. It unblocks dependents whose last outstanding dependency was
// this node.
func (s *graphScheduler) complete(img *image.Image) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.completed++

	for _, dependent := range s.graph.Dependents(img) {
		s.remaining[dependent]--
		if s.remaining[dependent] == 0 {
			s.ready <- s.index[dependent]
		}
	}

	if s.completed == len(s.nodes) {
		close(s.done)
	}
}
