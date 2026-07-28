package image

import "fmt"

// ImagesGraph is the build-time dependency graph over all *Image nodes
// (across all config images, their platforms, and — for staged Dockerfile
// images — their per-stage sub-images). An edge from B to A (A being one of
// B's Dependencies) means B's build cannot start until A has been fully
// built.
type ImagesGraph struct {
	nodes      []*Image
	deps       map[*Image][]*Image
	dependents map[*Image][]*Image
}

// BuildImagesGraph resolves each image's recorded dependency names (see
// Image.GetDependencyNames) into concrete *Image nodes (matched by name and
// target platform) and returns the resulting graph. Nodes() returns images
// in topological order (dependencies before dependents).
//
// A dependency name that cannot be resolved among the given images is
// ignored here: it either refers to an external base image (not built by
// werf) or will surface as a build-time error where the dependency's result
// is actually consumed (e.g. Image.SetupBaseImage), so this is not treated
// as a fatal error during graph construction.
func BuildImagesGraph(images []*Image) (*ImagesGraph, error) {
	byNameAndPlatform := make(map[string]*Image, len(images))
	for _, img := range images {
		byNameAndPlatform[imageGraphKey(img.Name, img.TargetPlatform)] = img
	}

	g := &ImagesGraph{
		deps:       make(map[*Image][]*Image, len(images)),
		dependents: make(map[*Image][]*Image, len(images)),
	}

	for _, img := range images {
		for _, depName := range img.GetDependencyNames() {
			depImg, ok := byNameAndPlatform[imageGraphKey(depName, img.TargetPlatform)]
			if !ok || depImg == img {
				continue
			}

			g.deps[img] = append(g.deps[img], depImg)
			g.dependents[depImg] = append(g.dependents[depImg], img)
		}
	}

	nodes, err := topologicalSort(images, g.deps)
	if err != nil {
		return nil, err
	}
	g.nodes = nodes

	return g, nil
}

func imageGraphKey(name, targetPlatform string) string {
	return targetPlatform + "\x00" + name
}

// Nodes returns all images in topological order (dependencies before dependents).
func (g *ImagesGraph) Nodes() []*Image {
	return g.nodes
}

// Dependencies returns the images that must be fully built before img can start building.
func (g *ImagesGraph) Dependencies(img *Image) []*Image {
	return g.deps[img]
}

// Dependents returns the images that depend on img.
func (g *ImagesGraph) Dependents(img *Image) []*Image {
	return g.dependents[img]
}

// Levels groups nodes into longest-path levels, for display purposes only
// (e.g. logging a "concurrent build plan"); it does not drive scheduling.
func (g *ImagesGraph) Levels() [][]*Image {
	level := make(map[*Image]int, len(g.nodes))
	var maxLevel int

	for _, img := range g.nodes {
		var l int
		for _, dep := range g.deps[img] {
			if level[dep]+1 > l {
				l = level[dep] + 1
			}
		}
		level[img] = l
		if l > maxLevel {
			maxLevel = l
		}
	}

	levels := make([][]*Image, maxLevel+1)
	for _, img := range g.nodes {
		l := level[img]
		levels[l] = append(levels[l], img)
	}

	return levels
}

func topologicalSort(images []*Image, deps map[*Image][]*Image) ([]*Image, error) {
	const (
		white = 0
		gray  = 1
		black = 2
	)

	color := make(map[*Image]int, len(images))
	result := make([]*Image, 0, len(images))

	var visit func(img *Image) error
	visit = func(img *Image) error {
		switch color[img] {
		case black:
			return nil
		case gray:
			return fmt.Errorf("dependency cycle detected involving image %q", img.Name)
		}

		color[img] = gray
		for _, dep := range deps[img] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		color[img] = black
		result = append(result, img)

		return nil
	}

	for _, img := range images {
		if err := visit(img); err != nil {
			return nil, err
		}
	}

	return result, nil
}
