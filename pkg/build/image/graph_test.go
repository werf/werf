package image

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildImagesGraph_ResolvesEdgesAndTopologicalOrder(t *testing.T) {
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

func TestBuildImagesGraph_MatchesEdgesByTargetPlatform(t *testing.T) {
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

func TestBuildImagesGraph_UnresolvableDependencyNameIsIgnoredNotFatal(t *testing.T) {
	a := newTestImage(t, "linux/amd64", "a")
	a.AddDependencyName("external-base-image-not-in-config")

	graph, err := BuildImagesGraph([]*Image{a})
	require.NoError(t, err)
	require.Empty(t, graph.Dependencies(a))
	require.Equal(t, []*Image{a}, graph.Nodes())
}

func TestBuildImagesGraph_ErrorsOnNameCollisionBetweenDistinctImages(t *testing.T) {
	// Simulates two distinct images that happen to produce the same
	// (name, platform) graph key — e.g. a plain image literally named
	// "app/stage/builder" colliding with a synthesized Dockerfile stage name.
	// Silently letting the last one win in the name->*Image map would
	// resolve dependency edges against the wrong node.
	first := newTestImage(t, "linux/amd64", "app/stage/builder")
	second := newTestImage(t, "linux/amd64", "app/stage/builder")

	_, err := BuildImagesGraph([]*Image{first, second})
	require.Error(t, err)
	require.ErrorContains(t, err, "app/stage/builder")
}

func TestBuildImagesGraph_SameImagePointerListedTwiceIsNotACollision(t *testing.T) {
	a := newTestImage(t, "linux/amd64", "a")
	b := newTestImage(t, "linux/amd64", "b")
	b.AddDependencyName("a")

	graph, err := BuildImagesGraph([]*Image{a, a, b, b})
	require.NoError(t, err)
	require.Equal(t, []*Image{a, b}, graph.Nodes())
	require.Equal(t, []*Image{a}, graph.Dependencies(b))
	require.Equal(t, []*Image{b}, graph.Dependents(a))
}

func TestBuildImagesGraph_ErrorsOnDependencyBuiltOnlyForOtherPlatforms(t *testing.T) {
	dep := newTestImage(t, "linux/arm64", "dep")
	app := newTestImage(t, "linux/amd64", "app")
	app.AddDependencyName("dep")

	_, err := BuildImagesGraph([]*Image{dep, app})
	require.Error(t, err)
	require.ErrorContains(t, err, `image "app" (platform "linux/amd64") depends on image "dep", which is not built for this platform (built for: linux/arm64)`)
}

func TestBuildImagesGraph_DetectsCycle(t *testing.T) {
	a := newTestImage(t, "linux/amd64", "a")
	b := newTestImage(t, "linux/amd64", "b")

	a.AddDependencyName("b")
	b.AddDependencyName("a")

	_, err := BuildImagesGraph([]*Image{a, b})
	require.Error(t, err)
}

func TestBuildImagesGraph_DiamondDependencyOrder(t *testing.T) {
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

func TestImagesGraph_LevelsGroupsDiamondByLongestPath(t *testing.T) {
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

	levels := graph.Levels()
	require.Len(t, levels, 3)
	require.Equal(t, []*Image{base}, levels[0])
	require.ElementsMatch(t, []*Image{left, right}, levels[1])
	require.Equal(t, []*Image{top}, levels[2])
}
