package path_matcher

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

type simpleMatchPathEntry struct {
	baseBase        string
	paths           []string
	matchedPaths    []string
	notMatchedPaths []string
}

var _ = DescribeTable("simple path matcher (MatchPath)", func(e simpleMatchPathEntry) {
	pathMatcher := NewSimplePathMatcher(e.baseBase, e.paths, false)

	for _, matchedPath := range e.matchedPaths {
		Ω(pathMatcher.MatchPath(matchedPath)).Should(BeTrue())
	}

	for _, notMatchedPath := range e.notMatchedPaths {
		Ω(pathMatcher.MatchPath(notMatchedPath)).Should(BeFalse())
	}
},
	Entry("basePath is equal to the path (paths)", simpleMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		paths:           []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath ''", simpleMatchPathEntry{
		baseBase:     filepath.Join("a", "b", "c"),
		paths:        []string{""},
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is relative to the basePath (paths)", simpleMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		paths:           []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
	}),

	Entry("path is not relative to the basePath(paths)", simpleMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		paths:           []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (paths)", simpleMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		paths:           []string{"d"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),

	Entry("glob completion by default (paths)", simpleMatchPathEntry{
		paths: []string{
			"a",
			filepath.Join("b", "*"),
			filepath.Join("c", "**"),
			filepath.Join("d", "**", "*"),
		},
		matchedPaths: []string{
			filepath.Join("a", "b", "c", "d"),
			filepath.Join("b", "b", "c", "d"),
			filepath.Join("c", "b", "c", "d"),
			filepath.Join("d", "b", "c", "d"),
		},
	}),
)

type simpleProcessDirOrSubmodulePath struct {
	baseBase               string
	paths                  []string
	matchedPaths           []string
	shouldWalkThroughPaths []string
	notMatchedPaths        []string
}

var _ = DescribeTable("simple (ProcessDirOrSubmodulePath)", func(e simpleProcessDirOrSubmodulePath) {
	pathMatcher := NewSimplePathMatcher(e.baseBase, e.paths, false)

	for _, matchedPath := range e.matchedPaths {
		isMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(matchedPath)
		Ω(isMatched).Should(BeTrue())
		Ω(shouldWalkThrough).Should(BeFalse())
	}

	for _, shouldWalkThroughPath := range e.shouldWalkThroughPaths {
		isMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(shouldWalkThroughPath)
		Ω(isMatched).Should(BeFalse())
		Ω(shouldWalkThrough).Should(BeTrue())
	}

	for _, notMatchedPath := range e.notMatchedPaths {
		isMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(notMatchedPath)
		Ω(isMatched).Should(BeFalse())
		Ω(shouldWalkThrough).Should(BeFalse())
	}
},
	Entry("basePath is equal to the path (base)", simpleProcessDirOrSubmodulePath{
		baseBase:     filepath.Join("a", "b", "c"),
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (paths)", simpleProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		paths:                  []string{"d"},
		shouldWalkThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is relative to the basePath (base)", simpleProcessDirOrSubmodulePath{
		baseBase:     filepath.Join("a", "b", "c"),
		matchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),
	Entry("path is relative to the basePath (paths)", simpleProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		paths:           []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
	}),

	Entry("path is not relative to the basePath (base)", simpleProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath(paths)", simpleProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		paths:           []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (base)", simpleProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (paths)", simpleProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		paths:                  []string{"d"},
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),

	Entry("glob completion by default (paths)", simpleProcessDirOrSubmodulePath{
		paths: []string{
			"a",
			filepath.Join("b", "*"),
			filepath.Join("c", "**"),
			filepath.Join("d", "**", "*"),
		},
		matchedPaths: []string{
			filepath.Join("a", "b", "c", "d"),
			filepath.Join("b", "b", "c", "d"),
			filepath.Join("c", "b", "c", "d"),
			filepath.Join("d", "b", "c", "d"),
		},
	}),
)
