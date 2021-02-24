package path_matcher

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

type simpleIsPathMatchedEntry struct {
	baseBase        string
	paths           []string
	matchedPaths    []string
	notMatchedPaths []string
}

var _ = DescribeTable("simple path matcher (IsPathMatched)", func(e simpleIsPathMatchedEntry) {
	pathMatcher := NewSimplePathMatcher(e.baseBase, e.paths)

	for _, matchedPath := range e.matchedPaths {
		Ω(pathMatcher.IsPathMatched(matchedPath)).Should(BeTrue(), fmt.Sprintln(pathMatcher, "==", matchedPath))
	}

	for _, notMatchedPath := range e.notMatchedPaths {
		Ω(pathMatcher.IsPathMatched(notMatchedPath)).Should(BeFalse(), fmt.Sprintln(pathMatcher, "!=", notMatchedPath))
	}
},
	Entry("basePath is equal to the path (paths)", simpleIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		paths:           []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "c")},
	}),
	Entry("basePath is equal to the path, includePath ''", simpleIsPathMatchedEntry{
		baseBase:     filepath.Join("a", "b", "c"),
		paths:        []string{""},
		matchedPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is relative to the basePath (paths)", simpleIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		paths:           []string{"d"},
		matchedPaths:    []string{filepath.Join("a", "b", "c", "d")},
		notMatchedPaths: []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
	}),

	Entry("path is not relative to the basePath(paths)", simpleIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		paths:           []string{"d"},
		notMatchedPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (paths)", simpleIsPathMatchedEntry{
		baseBase:        filepath.Join("a", "b", "c"),
		paths:           []string{"d"},
		notMatchedPaths: []string{filepath.Join("a")},
	}),

	Entry("glob completion by default (paths)", simpleIsPathMatchedEntry{
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

type simpleShouldGoThrough struct {
	baseBase                string
	paths                   []string
	shouldGoThroughPaths    []string
	shouldNotGoThroughPaths []string
}

var _ = DescribeTable("simple (ShouldGoThrough)", func(e simpleShouldGoThrough) {
	pathMatcher := NewSimplePathMatcher(e.baseBase, e.paths)

	for _, shouldGoThroughPath := range e.shouldGoThroughPaths {
		shouldGoThrough := pathMatcher.ShouldGoThrough(shouldGoThroughPath)
		Ω(shouldGoThrough).Should(BeTrue())
	}

	for _, shouldNotGoThroughPath := range e.shouldNotGoThroughPaths {
		shouldGoThrough := pathMatcher.ShouldGoThrough(shouldNotGoThroughPath)
		Ω(shouldGoThrough).Should(BeFalse())
	}
},
	Entry("basePath is equal to the path (paths)", simpleShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		paths:                []string{"d"},
		shouldGoThroughPaths: []string{filepath.Join("a", "b", "c")},
	}),

	Entry("path is relative to the basePath (paths)", simpleShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		paths:                   []string{"d"},
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "c", "e"), filepath.Join("a", "b", "c", "de")},
	}),

	Entry("path is not relative to the basePath (base)", simpleShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),
	Entry("path is not relative to the basePath(paths)", simpleShouldGoThrough{
		baseBase:                filepath.Join("a", "b", "c"),
		paths:                   []string{"d"},
		shouldNotGoThroughPaths: []string{filepath.Join("a", "b", "d"), "b"},
	}),

	Entry("basePath is relative to the path (base)", simpleShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		shouldGoThroughPaths: []string{filepath.Join("a")},
	}),
	Entry("basePath is relative to the path (paths)", simpleShouldGoThrough{
		baseBase:             filepath.Join("a", "b", "c"),
		paths:                []string{"d"},
		shouldGoThroughPaths: []string{filepath.Join("a")},
	}),
)
