package path_matcher

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

type gitMappingMatchPathEntry struct {
	baseBase        string
	includePaths    []string
	excludePaths    []string
	matchedPaths    []string
	notMatchedPaths []string
}

var _ = DescribeTable("git mapping path matcher (MatchPath)", func(e gitMappingMatchPathEntry) {
	pathMatcher := NewGitMappingPathMatcher(e.baseBase, e.includePaths, e.excludePaths)

	for _, matchedPath := range e.matchedPaths {
		Ω(pathMatcher.MatchPath(matchedPath)).Should(BeTrue())
	}

	for _, notMatchedPath := range e.notMatchedPaths {
		Ω(pathMatcher.MatchPath(notMatchedPath)).Should(BeFalse())
	}
},
	Entry("basePath is equal to the path (includePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (excludePaths)", gitMappingMatchPathEntry{
		baseBase:     filepath.Join("a", "b", "c"),
		excludePaths: []string{"d"},
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (includePaths and excludePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath ''", gitMappingMatchPathEntry{
		baseBase:     filepath.Join("a", "b", "c"),
		includePaths: []string{""},
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, excludePath ''", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{""},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath '', excludePath ''", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{""},
		excludePaths:    []string{""},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is relative to the basePath (includePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
	}),
	Entry("path is relative to the basePath (excludePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),
	Entry("path is relative to the basePath (includePaths and excludePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e")},
	}),

	Entry("path is not relative to the basePath(includePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (excludePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (includePaths and excludePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (includePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (excludePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (includePaths and excludePaths)", gitMappingMatchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),

	Entry("glob completion by default (includePaths)", gitMappingMatchPathEntry{
		includePaths: []string{
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
	Entry("glob completion by default (excludePaths)", gitMappingMatchPathEntry{
		excludePaths: []string{
			"a",
			filepath.Join("b", "*"),
			filepath.Join("c", "**"),
			filepath.Join("d", "**", "*"),
		},
		notMatchedPaths: []string{
			filepath.Join("a", "b", "c", "d"),
			filepath.Join("b", "b", "c", "d"),
			filepath.Join("c", "b", "c", "d"),
			filepath.Join("d", "b", "c", "d"),
		},
	}),
	Entry("glob completion by default (includePaths and excludePaths)", gitMappingMatchPathEntry{
		includePaths: []string{
			"a",
			filepath.Join("b", "*"),
			filepath.Join("c", "**"),
			filepath.Join("d", "**", "*"),
		},
		excludePaths: []string{
			"a",
			filepath.Join("b", "*"),
			filepath.Join("c", "**"),
			filepath.Join("d", "**", "*"),
		},
		notMatchedPaths: []string{
			filepath.Join("a", "b", "c", "d"),
			filepath.Join("b", "b", "c", "d"),
			filepath.Join("c", "b", "c", "d"),
			filepath.Join("d", "b", "c", "d"),
		},
	}),
)

type gitMappingProcessDirOrSubmodulePath struct {
	baseBase               string
	includePaths           []string
	excludePaths           []string
	matchedPaths           []string
	shouldWalkThroughPaths []string
	notMatchedPaths        []string
}

var _ = DescribeTable("git mapping path matcher (ProcessDirOrSubmodulePath)", func(e gitMappingProcessDirOrSubmodulePath) {
	pathMatcher := NewGitMappingPathMatcher(e.baseBase, e.includePaths, e.excludePaths)

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
	Entry("basePath is equal to the path (base)", gitMappingProcessDirOrSubmodulePath{
		baseBase:     filepath.Join("a", "b", "c"),
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (includePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		includePaths:           []string{"d"},
		shouldWalkThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (excludePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		excludePaths:           []string{"d"},
		shouldWalkThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (includePaths and excludePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		includePaths:           []string{"d"},
		excludePaths:           []string{"e"},
		shouldWalkThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath ''", gitMappingProcessDirOrSubmodulePath{
		baseBase:     filepath.Join("a", "b", "c"),
		includePaths: []string{""},
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, excludePath ''", gitMappingProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{""},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath '', excludePath ''", gitMappingProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{""},
		excludePaths:    []string{""},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is relative to the basePath (base)", gitMappingProcessDirOrSubmodulePath{
		baseBase:     filepath.Join("a", "b", "c"),
		matchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),
	Entry("path is relative to the basePath (includePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
	}),
	Entry("path is relative to the basePath (excludePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),
	Entry("path is relative to the basePath (includePaths and excludePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e")},
	}),

	Entry("path is not relative to the basePath (base)", gitMappingProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath(includePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (excludePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (includePaths and excludePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (base)", gitMappingProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (includePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		includePaths:           []string{"d"},
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (excludePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		excludePaths:           []string{"d"},
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (includePaths and excludePaths)", gitMappingProcessDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		includePaths:           []string{"d"},
		excludePaths:           []string{"e"},
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),

	Entry("glob completion by default (includePaths)", gitMappingProcessDirOrSubmodulePath{
		includePaths: []string{
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
	Entry("glob completion by default (excludePaths)", gitMappingProcessDirOrSubmodulePath{
		excludePaths: []string{
			"a",
			filepath.Join("b", "*"),
			filepath.Join("c", "**"),
			filepath.Join("d", "**", "*"),
		},
		notMatchedPaths: []string{
			filepath.Join("a", "b", "c", "d"),
			filepath.Join("b", "b", "c", "d"),
			filepath.Join("c", "b", "c", "d"),
			filepath.Join("d", "b", "c", "d"),
		},
	}),
	Entry("glob completion by default (includePaths and excludePaths)", gitMappingProcessDirOrSubmodulePath{
		includePaths: []string{
			"a",
			filepath.Join("b", "*"),
			filepath.Join("c", "**"),
			filepath.Join("d", "**", "*"),
		},
		excludePaths: []string{
			"a",
			filepath.Join("b", "*"),
			filepath.Join("c", "**"),
			filepath.Join("d", "**", "*"),
		},
		notMatchedPaths: []string{
			filepath.Join("a", "b", "c", "d"),
			filepath.Join("b", "b", "c", "d"),
			filepath.Join("c", "b", "c", "d"),
			filepath.Join("d", "b", "c", "d"),
		},
	}),
)
