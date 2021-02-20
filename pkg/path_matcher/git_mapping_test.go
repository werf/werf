package path_matcher

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

type gitMappingIsPathMatchedEntry struct {
	baseBase        string
	includePaths    []string
	excludePaths    []string
	matchedPaths    []string
	notMatchedPaths []string
}

var _ = DescribeTable("git mapping path matcher (IsPathMatched)", func(e gitMappingIsPathMatchedEntry) {
	pathMatcher := NewGitMappingPathMatcher(e.baseBase, e.includePaths, e.excludePaths)

	for _, matchedPath := range e.matchedPaths {
		Ω(pathMatcher.IsPathMatched(matchedPath)).Should(BeTrue(), matchedPath)
	}

	for _, notMatchedPath := range e.notMatchedPaths {
		Ω(pathMatcher.IsPathMatched(notMatchedPath)).Should(BeFalse(), notMatchedPath)
	}
},
	Entry("basePath is equal to the path (includePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (excludePaths)", gitMappingIsPathMatchedEntry{
		baseBase:     filepath.Join("a", "b", "c"),
		excludePaths: []string{"d"},
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (includePaths and excludePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath ''", gitMappingIsPathMatchedEntry{
		baseBase:     filepath.Join("a", "b", "c"),
		includePaths: []string{""},
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, excludePath ''", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{""},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath '', excludePath ''", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{""},
		excludePaths:    []string{""},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is relative to the basePath (includePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
	}),
	Entry("path is relative to the basePath (excludePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),
	Entry("path is relative to the basePath (includePaths and excludePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e")},
	}),

	Entry("path is not relative to the basePath(includePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (excludePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (includePaths and excludePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (includePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (excludePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		excludePaths:    []string{"d"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (includePaths and excludePaths)", gitMappingIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		includePaths:    []string{"d"},
		excludePaths:    []string{"e"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),

	Entry("glob completion by default (includePaths)", gitMappingIsPathMatchedEntry{
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
	Entry("glob completion by default (excludePaths)", gitMappingIsPathMatchedEntry{
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
	Entry("glob completion by default (includePaths and excludePaths)", gitMappingIsPathMatchedEntry{
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

type gitMappingShouldGoThrough struct {
	baseBase                string
	includePaths            []string
	excludePaths            []string
	shouldGoThroughPaths    []string
	shouldNotGoThroughPaths []string
}

var _ = DescribeTable("git mapping path matcher (ShouldGoThrough)", func(e gitMappingShouldGoThrough) {
	pathMatcher := NewGitMappingPathMatcher(e.baseBase, e.includePaths, e.excludePaths)

	for _, shouldGoThroughPath := range e.shouldGoThroughPaths {
		shouldGoThrough := pathMatcher.ShouldGoThrough(shouldGoThroughPath)
		Ω(shouldGoThrough).Should(BeTrue(), fmt.Sprintln(pathMatcher, "==", shouldGoThroughPath))
	}

	for _, shouldNotGoThroughPath := range e.shouldNotGoThroughPaths {
		shouldGoThrough := pathMatcher.ShouldGoThrough(shouldNotGoThroughPath)
		Ω(shouldGoThrough).Should(BeFalse(), fmt.Sprintln(pathMatcher, "!=", shouldNotGoThroughPath))
	}
},
	Entry("basePath is equal to the path (includePaths)", gitMappingShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		includePaths:         []string{"d"},
		shouldGoThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (excludePaths)", gitMappingShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		excludePaths:         []string{"d"},
		shouldGoThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path (includePaths and excludePaths)", gitMappingShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		includePaths:         []string{"d"},
		excludePaths:         []string{"e"},
		shouldGoThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, excludePath ''", gitMappingShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		excludePaths:            []string{""},
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath '', excludePath ''", gitMappingShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		includePaths:            []string{""},
		excludePaths:            []string{""},
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is relative to the basePath (includePaths)", gitMappingShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		includePaths:            []string{"d"},
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
	}),
	Entry("path is relative to the basePath (excludePaths)", gitMappingShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		excludePaths:            []string{"d"},
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "c", "d")},
	}),
	Entry("path is relative to the basePath (includePaths and excludePaths)", gitMappingShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		includePaths:            []string{"d"},
		excludePaths:            []string{"e"},
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "c", "e")},
	}),

	Entry("path is not relative to the basePath (base)", gitMappingShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath(includePaths)", gitMappingShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		includePaths:            []string{"d"},
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (excludePaths)", gitMappingShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		excludePaths:            []string{"d"},
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath (includePaths and excludePaths)", gitMappingShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		includePaths:            []string{"d"},
		excludePaths:            []string{"e"},
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (base)", gitMappingShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		shouldGoThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (includePaths)", gitMappingShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		includePaths:         []string{"d"},
		shouldGoThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (excludePaths)", gitMappingShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		excludePaths:         []string{"d"},
		shouldGoThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (includePaths and excludePaths)", gitMappingShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		includePaths:         []string{"d"},
		excludePaths:         []string{"e"},
		shouldGoThroughPaths: []string{filepath.Join("a")},
	}),

	Entry("glob completion by default (excludePaths)", gitMappingShouldGoThrough{
		excludePaths: []string{
			"a",
			filepath.Join("b", "*"),
			filepath.Join("c", "**"),
			filepath.Join("d", "**", "*"),
		},
		shouldNotGoThroughPaths: []string{
			filepath.Join("a", "b", "c", "d"),
			filepath.Join("b", "b", "c", "d"),
			filepath.Join("c", "b", "c", "d"),
			filepath.Join("d", "b", "c", "d"),
		},
	}),
	Entry("glob completion by default (includePaths and excludePaths)", gitMappingShouldGoThrough{
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
		shouldNotGoThroughPaths: []string{
			filepath.Join("a", "b", "c", "d"),
			filepath.Join("b", "b", "c", "d"),
			filepath.Join("c", "b", "c", "d"),
			filepath.Join("d", "b", "c", "d"),
		},
	}),
)
