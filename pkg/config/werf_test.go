package config

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/samber/lo"
)

var _ = Describe("WerfConfig", func() {
	Describe("validateInfiniteLoopBetweenRelatedImages", func() {
		DescribeTable("detects loops and accepts acyclic graphs",
			func(ctx SpecContext, images []ImageInterface, expectedErr types.GomegaMatcher) {
				werfConfig := NewWerfConfig(nil, images)

				Expect(werfConfig.validateInfiniteLoopBetweenRelatedImages()).To(expectedErr)
			},
			Entry("no related images", []ImageInterface{
				NewImageStub("a", DependsOn{}),
				NewImageStub("b", DependsOn{}),
			}, Succeed()),
			Entry("linear chain a -> b -> c", []ImageInterface{
				NewImageStub("a", DependsOn{From: "b"}),
				NewImageStub("b", DependsOn{Imports: []string{"c"}}),
				NewImageStub("c", DependsOn{}),
			}, Succeed()),
			Entry("self reference a -> a", []ImageInterface{
				NewImageStub("a", DependsOn{From: "a"}),
			}, MatchError("infinite loop detected: a -> a")),
			Entry("loop through from a -> b -> a", []ImageInterface{
				NewImageStub("a", DependsOn{From: "b"}),
				NewImageStub("b", DependsOn{From: "a"}),
			}, MatchError("infinite loop detected: a -> b -> a")),
			Entry("loop through import a -> b -> c -> a", []ImageInterface{
				NewImageStub("a", DependsOn{Imports: []string{"b"}}),
				NewImageStub("b", DependsOn{Imports: []string{"c"}}),
				NewImageStub("c", DependsOn{Imports: []string{"a"}}),
			}, MatchError("infinite loop detected: a -> b -> c -> a")),
			Entry("loop through dependencies b -> c -> b, reported from the first image reaching it", []ImageInterface{
				NewImageStub("a", DependsOn{Dependencies: []string{"b"}}),
				NewImageStub("b", DependsOn{Dependencies: []string{"c"}}),
				NewImageStub("c", DependsOn{Dependencies: []string{"b"}}),
			}, MatchError("infinite loop detected: a -> b -> c -> b")),
			Entry("loop reachable only from the second image, after the first one is fully walked", []ImageInterface{
				NewImageStub("a", DependsOn{From: "shared"}),
				NewImageStub("shared", DependsOn{}),
				NewImageStub("b", DependsOn{From: "shared", Imports: []string{"c"}}),
				NewImageStub("c", DependsOn{Dependencies: []string{"b"}}),
			}, MatchError("infinite loop detected: b -> c -> b")),
			Entry("loop below a shared image, reached through it from two images", []ImageInterface{
				NewImageStub("a", DependsOn{From: "shared"}),
				NewImageStub("b", DependsOn{From: "shared"}),
				NewImageStub("shared", DependsOn{Imports: []string{"x"}}),
				NewImageStub("x", DependsOn{Imports: []string{"y"}}),
				NewImageStub("y", DependsOn{Imports: []string{"x"}}),
			}, MatchError("infinite loop detected: a -> shared -> x -> y -> x")),
			Entry("diamond DAG of 30 layers by 2 images (2^29 root-to-leaf paths) is validated within the timeout",
				SpecTimeout(10*time.Second),
				newDiamondDAGImages(30, 2), Succeed()),
			Entry("diamond DAG of 30 layers by 2 images with an edge from the bottom back to the top is a loop",
				SpecTimeout(10*time.Second),
				newDiamondDAGImagesWithLoop(30, 2),
				MatchError("infinite loop detected: "+diamondDAGColumn(0, 28, 0)+" -> l29-1 -> l0-1 -> l1-0")),
		)
	})
})

// newDiamondDAGImages builds layers*width images named l<layer>-<i>, where every image of layer i
// depends on every image of layer i+1 (the first one via from, the rest via import), so the number
// of root-to-leaf paths is width^(layers-1).
func newDiamondDAGImages(layers, width int) []ImageInterface {
	return lo.Map(newDiamondDAGImageStubs(layers, width), func(image *ImageStub, _ int) ImageInterface { return image })
}

func newDiamondDAGImagesWithLoop(layers, width int) []ImageInterface {
	images := newDiamondDAGImageStubs(layers, width)
	images[len(images)-1].deps = DependsOn{From: fmt.Sprintf("l0-%d", width-1)}
	return lo.Map(images, func(image *ImageStub, _ int) ImageInterface { return image })
}

func newDiamondDAGImageStubs(layers, width int) []*ImageStub {
	var images []*ImageStub
	for layer := 0; layer < layers; layer++ {
		for i := 0; i < width; i++ {
			var deps DependsOn
			if layer+1 < layers {
				deps.From = fmt.Sprintf("l%d-%d", layer+1, 0)
				for j := 1; j < width; j++ {
					deps.Imports = append(deps.Imports, fmt.Sprintf("l%d-%d", layer+1, j))
				}
			}
			images = append(images, NewImageStub(fmt.Sprintf("l%d-%d", layer, i), deps))
		}
	}
	return images
}

func diamondDAGColumn(fromLayer, toLayer, i int) string {
	var names []string
	for layer := fromLayer; layer <= toLayer; layer++ {
		names = append(names, fmt.Sprintf("l%d-%d", layer, i))
	}
	return strings.Join(names, " -> ")
}
