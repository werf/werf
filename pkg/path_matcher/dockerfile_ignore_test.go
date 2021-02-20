package path_matcher

import (
	"path/filepath"

	"github.com/docker/docker/pkg/fileutils"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func newPatternMatcher(patterns []string) *fileutils.PatternMatcher {
	m, err := fileutils.NewPatternMatcher(patterns)
	if err != nil {
		panic(err)
	}

	return m
}

type dockerfileIgnoreIsPathMatchedEntry struct {
	baseBase        string
	patternMatcher  *fileutils.PatternMatcher
	matchedPaths    []string
	notMatchedPaths []string
}

var _ = DescribeTable("dockerfile ignore path matcher (IsPathMatched)", func(e dockerfileIgnoreIsPathMatchedEntry) {
	pathMatcher := NewDockerfileIgnorePathMatcher(e.baseBase, e.patternMatcher)

	for _, matchedPath := range e.matchedPaths {
		Ω(pathMatcher.IsPathMatched(matchedPath)).Should(BeTrue(), matchedPath)
	}

	for _, notMatchedPath := range e.notMatchedPaths {
		Ω(pathMatcher.IsPathMatched(notMatchedPath)).Should(BeFalse(), notMatchedPath)
	}
},
	Entry("path is relative to the basePath (exclude)", dockerfileIgnoreIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d"}),
		matchedPaths:    []string{filepath.Join("a", "b", "c", "de")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "d"), filepath.Join("a", "b", "c", "d", "e")},
	}),
	Entry("path is relative to the basePath (exclude with exclusion)", dockerfileIgnoreIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d", "!d/e"}),
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d", "e")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),

	Entry("path is not relative to the basePath(exclude)", dockerfileIgnoreIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d"}),
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (exclude with exclusion)", dockerfileIgnoreIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d", "!d/e"}),
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (exclude)", dockerfileIgnoreIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d"}),
		notMatchedPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (exclude with exclusion)", dockerfileIgnoreIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d", "!d/e"}),
		notMatchedPaths: []string{filepath.Join("a")},
	}),
)

type dockerfileIgnoreShouldGoThrough struct {
	baseBase               string
	patternMatcher         *fileutils.PatternMatcher
	shouldGoThroughPaths   []string
	shouldNotGoThroughPath []string
}

var _ = DescribeTable("dockerfile ignore path matcher (ShouldGoThrough)", func(e dockerfileIgnoreShouldGoThrough) {
	pathMatcher := NewDockerfileIgnorePathMatcher(e.baseBase, e.patternMatcher)

	for _, shouldGoThroughPath := range e.shouldGoThroughPaths {
		shouldGoThrough := pathMatcher.ShouldGoThrough(shouldGoThroughPath)
		Ω(shouldGoThrough).Should(BeTrue(), shouldGoThroughPath)
	}

	for _, shouldNotGoThroughPath := range e.shouldNotGoThroughPath {
		shouldGoThrough := pathMatcher.ShouldGoThrough(shouldNotGoThroughPath)
		Ω(shouldGoThrough).Should(BeFalse(), shouldNotGoThroughPath)
	}
},
	Entry("path is relative to the basePath (exclude)", dockerfileIgnoreShouldGoThrough{
		baseBase:               filepath.Join("a", "b", "c"),
		patternMatcher:         newPatternMatcher([]string{"d"}),
		shouldNotGoThroughPath: []string{filepath.Join("a", "b", "c", "d"), filepath.Join("a", "b", "c", "d", "e")},
	}),
	Entry("path is relative to the basePath (exclude with exclusion)", dockerfileIgnoreShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		patternMatcher:       newPatternMatcher([]string{"d", "!d/e"}),
		shouldGoThroughPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),

	Entry("basePath is equal to the path (exclude)", dockerfileIgnoreShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		patternMatcher:       newPatternMatcher([]string{"d"}),
		shouldGoThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (exclude with exclusion)", dockerfileIgnoreShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		patternMatcher:       newPatternMatcher([]string{"d", "!d/e"}),
		shouldGoThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is not relative to the basePath(exclude)", dockerfileIgnoreShouldGoThrough{
		baseBase:               filepath.Join("a", "b", "c"),
		patternMatcher:         newPatternMatcher([]string{"d"}),
		shouldNotGoThroughPath: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (exclude with exclusion)", dockerfileIgnoreShouldGoThrough{
		baseBase:               filepath.Join("a", "b", "c"),
		patternMatcher:         newPatternMatcher([]string{"d", "!d/e"}),
		shouldNotGoThroughPath: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (exclude)", dockerfileIgnoreShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		patternMatcher:       newPatternMatcher([]string{"d"}),
		shouldGoThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (exclude with exclusion)", dockerfileIgnoreShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		patternMatcher:       newPatternMatcher([]string{"d", "!d/e"}),
		shouldGoThroughPaths: []string{filepath.Join("a")},
	}),
)
