package path_matcher

import (
	"path/filepath"

	"github.com/docker/docker/pkg/fileutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type pathMatcherEntry struct {
	pathMatcherName         string
	newOnlyWithBasePathFunc func(basePath string) PathMatcher
}

var _ = Describe("path matcher (basePath)", func() {
	for _, entry := range []pathMatcherEntry{
		{
			pathMatcherName: "dockerfile ignore",
			newOnlyWithBasePathFunc: func(basePath string) PathMatcher {
				pm, _ := fileutils.NewPatternMatcher([]string{})
				return NewDockerfileIgnorePathMatcher(basePath, pm)
			},
		},
		{
			pathMatcherName: "git mapping",
			newOnlyWithBasePathFunc: func(basePath string) PathMatcher {
				return NewGitMappingPathMatcher(basePath, []string{}, []string{})
			},
		},
		{
			pathMatcherName: "simple",
			newOnlyWithBasePathFunc: func(basePath string) PathMatcher {
				return NewSimplePathMatcher(basePath, []string{})
			},
		},
	} {
		Context(entry.pathMatcherName, func() {
			var pathMatcher = entry.newOnlyWithBasePathFunc(filepath.Join("a", "b", "c"))

			It("base path is equal to the path", func() {
				isMatched := pathMatcher.IsPathMatched(filepath.Join("a", "b", "c"))
				Ω(isMatched).Should(BeTrue())

				shouldGoThrough := pathMatcher.ShouldGoThrough(filepath.Join("a", "b", "c"))
				Ω(shouldGoThrough).Should(BeFalse())
			})

			It("path is relative to the basePath", func() {
				isMatched := pathMatcher.IsPathMatched(filepath.Join("a", "b", "c", "d"))
				Ω(isMatched).Should(BeTrue())

				shouldGoThrough := pathMatcher.ShouldGoThrough(filepath.Join("a", "b", "c", "d"))
				Ω(shouldGoThrough).Should(BeFalse())
			})

			It("path is not relative to the basePath", func() {
				for _, path := range []string{filepath.Join("a", "b", "d"), "b"} {
					isMatched := pathMatcher.IsPathMatched(path)
					Ω(isMatched).Should(BeFalse())

					shouldGoThrough := pathMatcher.ShouldGoThrough(path)
					Ω(shouldGoThrough).Should(BeFalse())
				}
			})

			It("basePath is relative to the path", func() {
				pathMatcher := entry.newOnlyWithBasePathFunc(filepath.Join("a"))

				isMatched := pathMatcher.IsPathMatched(filepath.Join("a", "b", "c"))
				Ω(isMatched).Should(BeTrue())

				shouldGoThrough := pathMatcher.ShouldGoThrough(filepath.Join("a", "b", "c"))
				Ω(shouldGoThrough).Should(BeFalse())
			})
		})
	}
})
