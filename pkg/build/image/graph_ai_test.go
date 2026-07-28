package image

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAI_BuildImagesGraph_ResolvesEdgesAndTopologicalOrder(t *testing.T) {
	a := newTestImage(t, "linux/amd64", "a")
	b := newTestImage(t, "linux/amd64", "b")
	c := newTestImage(t, "linux/amd64", "c")

	// c depends on b, b depends on a: a -> b -> c
	b.AddDependencyName("a")
	c.AddDependencyName("b")

	graph, err := BuildImagesGraph([]*Image{c, b, a})
	require.NoError(t, err)

	order := make(map[string]int, len(graph.Nodes()))
	for i, img := range graph.Nodes() {
		order[img.Name] = i
	}

	require.Less(t, order["a"], order["b"])
	require.Less(t, order["b"], order["c"])

	require.ElementsMatch(t, []*Image{a}, graph.Dependencies(b))
	require.ElementsMatch(t, []*Image{b}, graph.Dependencies(c))
	require.Empty(t, graph.Dependencies(a))

	require.ElementsMatch(t, []*Image{b}, graph.Dependents(a))
	require.ElementsMatch(t, []*Image{c}, graph.Dependents(b))
}

func TestAI_BuildImagesGraph_MatchesEdgesByTargetPlatform(t *testing.T) {
	// two platforms of the same top-level images, each platform's edges must
	// resolve to the same-platform counterpart, not cross-platform.
	aAmd := newTestImage(t, "linux/amd64", "a")
	bAmd := newTestImage(t, "linux/amd64", "b")
	bAmd.AddDependencyName("a")

	aArm := newTestImage(t, "linux/arm64", "a")
	bArm := newTestImage(t, "linux/arm64", "b")
	bArm.AddDependencyName("a")

	graph, err := BuildImagesGraph([]*Image{aAmd, bAmd, aArm, bArm})
	require.NoError(t, err)

	require.ElementsMatch(t, []*Image{aAmd}, graph.Dependencies(bAmd))
	require.ElementsMatch(t, []*Image{aArm}, graph.Dependencies(bArm))
}

func TestAI_BuildImagesGraph_UnresolvableDependencyNameIsIgnoredNotFatal(t *testing.T) {
	a := newTestImage(t, "linux/amd64", "a")
	a.AddDependencyName("external-base-image-not-in-config")

	graph, err := BuildImagesGraph([]*Image{a})
	require.NoError(t, err)
	require.Empty(t, graph.Dependencies(a))
	require.Equal(t, []*Image{a}, graph.Nodes())
}

func TestAI_BuildImagesGraph_DetectsCycle(t *testing.T) {
	a := newTestImage(t, "linux/amd64", "a")
	b := newTestImage(t, "linux/amd64", "b")

	a.AddDependencyName("b")
	b.AddDependencyName("a")

	_, err := BuildImagesGraph([]*Image{a, b})
	require.Error(t, err)
}

func TestAI_BuildImagesGraph_DiamondDependencyOrder(t *testing.T) {
	// base <- {left, right} <- top (a diamond, e.g. a Dockerfile stage
	// referenced via COPY --from by two other stages, both feeding a final one)
	base := newTestImage(t, "linux/amd64", "base")
	left := newTestImage(t, "linux/amd64", "left")
	right := newTestImage(t, "linux/amd64", "right")
	top := newTestImage(t, "linux/amd64", "top")

	left.AddDependencyName("base")
	right.AddDependencyName("base")
	top.AddDependencyName("left")
	top.AddDependencyName("right")

	graph, err := BuildImagesGraph([]*Image{top, right, left, base})
	require.NoError(t, err)

	order := make(map[string]int, len(graph.Nodes()))
	for i, img := range graph.Nodes() {
		order[img.Name] = i
	}

	require.Less(t, order["base"], order["left"])
	require.Less(t, order["base"], order["right"])
	require.Less(t, order["left"], order["top"])
	require.Less(t, order["right"], order["top"])
}
