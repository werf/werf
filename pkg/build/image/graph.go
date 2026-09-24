package image

import (
	"fmt"
	"sort"
	"strings"
)

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
// A dependency name that resolves to no given image on any platform is
// ignored here: it refers to an external base image (not built by werf). A
// dependency name that resolves to a given image only on OTHER platforms is
// a configuration error (the dependency's result could never be consumed on
// the dependent's platform) and fails graph construction. Duplicate *Image
// pointers in the input are deduplicated.
func BuildImagesGraph(images []*Image) (*ImagesGraph, error) {
	seen := make(map[*Image]struct{}, len(images))
	dedupedImages := make([]*Image, 0, len(images))
	for _, img := range images {
		if _, ok := seen[img]; ok {
			continue
		}
		seen[img] = struct{}{}
		dedupedImages = append(dedupedImages, img)
	}

	byNameAndPlatform := make(map[string]*Image, len(dedupedImages))
	platformsByName := make(map[string][]string, len(dedupedImages))
	for _, img := range dedupedImages {
		key := imageGraphKey(img.Name, img.TargetPlatform)
		if _, ok := byNameAndPlatform[key]; ok {
			return nil, fmt.Errorf("build graph name conflict: two distinct images both resolve to name %q on platform %q — image names must not collide with synthesized Dockerfile stage names (\"<image>/stage/<name>\")", img.Name, img.TargetPlatform)
		}
		byNameAndPlatform[key] = img
		platformsByName[img.Name] = append(platformsByName[img.Name], img.TargetPlatform)
	}

	g := &ImagesGraph{
		deps:       make(map[*Image][]*Image, len(dedupedImages)),
		dependents: make(map[*Image][]*Image, len(dedupedImages)),
	}

	for _, img := range dedupedImages {
		for _, depName := range img.GetDependencyNames() {
			depImg, ok := byNameAndPlatform[imageGraphKey(depName, img.TargetPlatform)]
			if !ok {
				if platforms, isNode := platformsByName[depName]; isNode {
					sort.Strings(platforms)
					return nil, fmt.Errorf("image %q (platform %q) depends on image %q, which is not built for this platform (built for: %s)", img.Name, img.TargetPlatform, depName, strings.Join(platforms, ", "))
				}
				continue
			}
			if depImg == img {
				continue
			}

			g.deps[img] = append(g.deps[img], depImg)
			g.dependents[depImg] = append(g.dependents[depImg], img)
		}
	}

	nodes, err := topologicalSort(dedupedImages, g.deps)
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
