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

type dockerfileIgnorePathMatchEntry struct {
	baseBase        string
	patternMatcher  *fileutils.PatternMatcher
	matchedPaths    []string
	notMatchedPaths []string
}

var _ = DescribeTable("dockerfile ignore path matcher (MatchPath)", func(e dockerfileIgnorePathMatchEntry) {
	pathMatcher := NewDockerfileIgnorePathMatcher(e.baseBase, e.patternMatcher)

	for _, matchedPath := range e.matchedPaths {
		Ω(pathMatcher.MatchPath(matchedPath)).Should(BeTrue())
	}

	for _, notMatchedPath := range e.notMatchedPaths {
		Ω(pathMatcher.MatchPath(notMatchedPath)).Should(BeFalse())
	}
},
	Entry("path is relative to the basePath (exclude)", dockerfileIgnorePathMatchEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d"}),
		matchedPaths:    []string{filepath.Join("a", "b", "c", "de")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "d"), filepath.Join("a", "b", "c", "d", "e")},
	}),
	Entry("path is relative to the basePath (exclude with exclusion)", dockerfileIgnorePathMatchEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d", "!d/e"}),
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d", "e")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),

	Entry("path is not relative to the basePath(exclude)", dockerfileIgnorePathMatchEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d"}),
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (exclude with exclusion)", dockerfileIgnorePathMatchEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d", "!d/e"}),
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (exclude)", dockerfileIgnorePathMatchEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d"}),
		notMatchedPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (exclude with exclusion)", dockerfileIgnorePathMatchEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d", "!d/e"}),
		notMatchedPaths: []string{filepath.Join("a")},
	}),
)

type dockerfileIgnoreProcessDirOrSubmodulePath struct {
	baseBase               string
	patternMatcher         *fileutils.PatternMatcher
	matchedPaths           []string
	shouldWalkThroughPaths []string
	notMatchedPaths        []string
}

var _ = DescribeTable("dockerfile ignore path matcher (ProcessDirOrSubmodulePath)", func(e dockerfileIgnoreProcessDirOrSubmodulePath) {
	pathMatcher := NewDockerfileIgnorePathMatcher(e.baseBase, e.patternMatcher)

	for _, matchedPath := range e.matchedPaths {
		isMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(matchedPath)
		Ω(isMatched).Should(BeTrue(), matchedPath)
		Ω(shouldWalkThrough).Should(BeFalse(), matchedPath)
	}

	for _, shouldWalkThroughPath := range e.shouldWalkThroughPaths {
		isMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(shouldWalkThroughPath)
		Ω(isMatched).Should(BeFalse(), shouldWalkThroughPath)
		Ω(shouldWalkThrough).Should(BeTrue(), shouldWalkThroughPath)
	}

	for _, notMatchedPath := range e.notMatchedPaths {
		isMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(notMatchedPath)
		Ω(isMatched).Should(BeFalse(), notMatchedPath)
		Ω(shouldWalkThrough).Should(BeFalse(), notMatchedPath)
	}
},
	Entry("path is relative to the basePath (exclude)", dockerfileIgnoreProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d"}),
		matchedPaths:    []string{filepath.Join("a", "b", "c", "de")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "d"), filepath.Join("a", "b", "c", "d", "e")},
	}),
	Entry("path is relative to the basePath (exclude with exclusion)", dockerfileIgnoreProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		patternMatcher:         newPatternMatcher([]string{"d", "!d/e"}),
		matchedPaths:           []string{filepath.Join("a", "b", "c", "d", "e")},
		shouldWalkThroughPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),

	Entry("basePath is equal to the path (exclude)", dockerfileIgnoreProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		patternMatcher:         newPatternMatcher([]string{"d"}),
		shouldWalkThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (exclude with exclusion)", dockerfileIgnoreProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		patternMatcher:         newPatternMatcher([]string{"d", "!d/e"}),
		shouldWalkThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is not relative to the basePath(exclude)", dockerfileIgnoreProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d"}),
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (exclude with exclusion)", dockerfileIgnoreProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		patternMatcher:  newPatternMatcher([]string{"d", "!d/e"}),
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (exclude)", dockerfileIgnoreProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		patternMatcher:         newPatternMatcher([]string{"d"}),
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (exclude with exclusion)", dockerfileIgnoreProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		patternMatcher:         newPatternMatcher([]string{"d", "!d/e"}),
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),
)
