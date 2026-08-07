package build

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/build/image"
)

func drainReadyNames(s *graphScheduler, nodes []*image.Image) map[string]bool {
	names := map[string]bool{}
	for {
		select {
		case idx := <-s.ready:
			names[nodes[idx].Name] = true
		default:
			return names
		}
	}
}

// TestGraphScheduler_FastChainDoesNotWaitForUnrelatedSlowImage is the core
// regression test for replacing wave/level barrier scheduling with a real
// dependency graph: a fast chain (a -> b -> c) must become ready to build
// step by step as soon as ITS OWN predecessor finishes, without waiting for
// an unrelated "slow" image (which has no dependency relation to the chain)
// to complete first.
func TestGraphScheduler_FastChainDoesNotWaitForUnrelatedSlowImage(t *testing.T) {
	a := &image.Image{Name: "a", TargetPlatform: "linux/amd64"}
	b := &image.Image{Name: "b", TargetPlatform: "linux/amd64"}
	c := &image.Image{Name: "c", TargetPlatform: "linux/amd64"}
	slow := &image.Image{Name: "slow", TargetPlatform: "linux/amd64"}

	b.AddDependencyName("a")
	c.AddDependencyName("b")

	graph, err := image.BuildImagesGraph([]*image.Image{a, b, c, slow})
	require.NoError(t, err)

	nodes := graph.Nodes()
	byName := make(map[string]*image.Image, len(nodes))
	for _, img := range nodes {
		byName[img.Name] = img
	}

	s := newGraphScheduler(graph)

	// Only images with zero dependencies are ready up front.
	require.Equal(t, map[string]bool{"a": true, "slow": true}, drainReadyNames(s, nodes))

	// Completing "a" must ready "b" immediately, regardless of whether the
	// unrelated "slow" image (also still running) has finished.
	s.complete(byName["a"])
	require.Equal(t, map[string]bool{"b": true}, drainReadyNames(s, nodes))

	s.complete(byName["b"])
	require.Equal(t, map[string]bool{"c": true}, drainReadyNames(s, nodes))

	select {
	case <-s.done:
		t.Fatal("scheduler must not be done: \"slow\" and \"c\" have not completed yet")
	default:
	}

	s.complete(byName["c"])
	select {
	case <-s.done:
		t.Fatal("scheduler must not be done: \"slow\" has not completed yet")
	default:
	}

	s.complete(byName["slow"])
	select {
	case <-s.done:
	default:
		t.Fatal("scheduler must be done once every node has completed")
	}
}

func TestGraphScheduler_DiamondDependencyReadyOnlyAfterBothParents(t *testing.T) {
	base := &image.Image{Name: "base", TargetPlatform: "linux/amd64"}
	left := &image.Image{Name: "left", TargetPlatform: "linux/amd64"}
	right := &image.Image{Name: "right", TargetPlatform: "linux/amd64"}
	top := &image.Image{Name: "top", TargetPlatform: "linux/amd64"}

	left.AddDependencyName("base")
	right.AddDependencyName("base")
	top.AddDependencyName("left")
	top.AddDependencyName("right")

	graph, err := image.BuildImagesGraph([]*image.Image{base, left, right, top})
	require.NoError(t, err)

	nodes := graph.Nodes()
	byName := make(map[string]*image.Image, len(nodes))
	for _, img := range nodes {
		byName[img.Name] = img
	}

	s := newGraphScheduler(graph)
	require.Equal(t, map[string]bool{"base": true}, drainReadyNames(s, nodes))

	s.complete(byName["base"])
	require.Equal(t, map[string]bool{"left": true, "right": true}, drainReadyNames(s, nodes))

	s.complete(byName["left"])
	require.Empty(t, drainReadyNames(s, nodes), "top must not be ready until both left and right complete")

	s.complete(byName["right"])
	require.Equal(t, map[string]bool{"top": true}, drainReadyNames(s, nodes))

	s.complete(byName["top"])
	select {
	case <-s.done:
	default:
		t.Fatal("scheduler must be done once every node has completed")
	}
}

func TestGraphScheduler_IndependentSetHandsOutEveryNodeExactlyOnceThenTerminates(t *testing.T) {
	a := &image.Image{Name: "a", TargetPlatform: "linux/amd64"}
	b := &image.Image{Name: "b", TargetPlatform: "linux/amd64"}
	c := &image.Image{Name: "c", TargetPlatform: "linux/amd64"}

	graph, err := image.BuildImagesGraph([]*image.Image{a, b, c})
	require.NoError(t, err)

	nodes := graph.Nodes()
	s := newGraphScheduler(graph)

	handedOut := map[string]int{}
	for {
		taskId, ok, err := s.next(context.Background())
		require.NoError(t, err)
		if !ok {
			break
		}
		handedOut[nodes[taskId].Name]++
		s.complete(nodes[taskId])
	}

	require.Equal(t, map[string]int{"a": 1, "b": 1, "c": 1}, handedOut)

	_, ok, err := s.next(context.Background())
	require.NoError(t, err)
	require.False(t, ok, "next must stay terminal after every node has completed")
}

func TestGraphScheduler_ZeroNodesTerminatesImmediately(t *testing.T) {
	graph, err := image.BuildImagesGraph(nil)
	require.NoError(t, err)

	s := newGraphScheduler(graph)

	_, ok, err := s.next(context.Background())
	require.NoError(t, err)
	require.False(t, ok)
}
