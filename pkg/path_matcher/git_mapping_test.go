package path_matcher

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

type matchPathEntry struct {
	baseBase        string
	includePaths    []string
	excludePaths    []string
	matchedPaths    []string
	notMatchedPaths []string
}

var _ = DescribeTable("GitMapping_MatchPath", func(e matchPathEntry) {
	pathMatcher := NewGitMappingPathMatcher(e.baseBase, e.includePaths, e.excludePaths)

	for _, matchedPath := range e.matchedPaths {
		Ω(pathMatcher.MatchPath(matchedPath)).Should(BeTrue())
	}

	for _, notMatchedPath := range e.notMatchedPaths {
		Ω(pathMatcher.MatchPath(notMatchedPath)).Should(BeFalse())
	}
},
	Entry("basePath is equal to the path (base)", matchPathEntry{
		baseBase:     filepath.Join("a", "b", "c"),
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (includePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (excludePaths)", matchPathEntry{
		baseBase:     filepath.Join("a", "b", "c"),
		excludePaths: []string{"d"},
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (includePaths and excludePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath ''", matchPathEntry{
		baseBase:     filepath.Join("a", "b", "c"),
		includePaths: []string{""},
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, excludePath ''", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{""},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath '', excludePath ''", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{""},
		excludePaths:    []string{""},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is relative to the basePath (base)", matchPathEntry{
		baseBase:     filepath.Join("a", "b", "c"),
		matchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),
	Entry("path is relative to the basePath (includePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
	}),
	Entry("path is relative to the basePath (excludePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),
	Entry("path is relative to the basePath (includePaths and excludePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e")},
	}),

	Entry("path is not relative to the basePath (base)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath(includePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (excludePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (includePaths and excludePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (base)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		notMatchedPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (includePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (excludePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (includePaths and excludePaths)", matchPathEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),

	Entry("glob completion by default (includePaths)", matchPathEntry{
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
	Entry("glob completion by default (excludePaths)", matchPathEntry{
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
	Entry("glob completion by default (includePaths and excludePaths)", matchPathEntry{
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

type processDirOrSubmodulePath struct {
	baseBase               string
	includePaths           []string
	excludePaths           []string
	matchedPaths           []string
	shouldWalkThroughPaths []string
	notMatchedPaths        []string
}

var _ = DescribeTable("GitMapping_ProcessDirOrSubmodulePath", func(e processDirOrSubmodulePath) {
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
	Entry("basePath is equal to the path (base)", processDirOrSubmodulePath{
		baseBase:     filepath.Join("a", "b", "c"),
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (includePaths)", processDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		includePaths:           []string{"d"},
		shouldWalkThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (excludePaths)", processDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		excludePaths:           []string{"d"},
		shouldWalkThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (includePaths and excludePaths)", processDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		includePaths:           []string{"d"},
		excludePaths:           []string{"e"},
		shouldWalkThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath ''", processDirOrSubmodulePath{
		baseBase:     filepath.Join("a", "b", "c"),
		includePaths: []string{""},
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, excludePath ''", processDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{""},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath '', excludePath ''", processDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{""},
		excludePaths:    []string{""},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is relative to the basePath (base)", processDirOrSubmodulePath{
		baseBase:     filepath.Join("a", "b", "c"),
		matchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),
	Entry("path is relative to the basePath (includePaths)", processDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
	}),
	Entry("path is relative to the basePath (excludePaths)", processDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),
	Entry("path is relative to the basePath (includePaths and excludePaths)", processDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e")},
	}),

	Entry("path is not relative to the basePath (base)", processDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath(includePaths)", processDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (excludePaths)", processDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (includePaths and excludePaths)", processDirOrSubmodulePath{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (base)", processDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (includePaths)", processDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		includePaths:           []string{"d"},
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (excludePaths)", processDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		excludePaths:           []string{"d"},
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (includePaths and excludePaths)", processDirOrSubmodulePath{
		baseBase:               filepath.Join("a", "b", "c"),
		includePaths:           []string{"d"},
		excludePaths:           []string{"e"},
		shouldWalkThroughPaths: []string{filepath.Join("a")},
	}),

	Entry("glob completion by default (includePaths)", processDirOrSubmodulePath{
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
	Entry("glob completion by default (excludePaths)", processDirOrSubmodulePath{
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
	Entry("glob completion by default (includePaths and excludePaths)", processDirOrSubmodulePath{
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
