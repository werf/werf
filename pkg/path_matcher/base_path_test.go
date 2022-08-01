package path_matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("base path matcher", func() {
	type entry struct {
		basePath                    string
		testPath                    string
		isPathMatched               bool
		shouldGoThrough             bool
		isDirOrSubmodulePathMatched bool
	}

	itBodyFunc := func(e entry) {
		matcher := newBasePathMatcher(e.basePath, nil)

		Ω(matcher.IsPathMatched(e.testPath)).Should(BeEquivalentTo(e.isPathMatched))
		Ω(matcher.ShouldGoThrough(e.testPath)).Should(BeEquivalentTo(e.shouldGoThrough))
		Ω(matcher.IsDirOrSubmodulePathMatched(e.testPath)).Should(BeEquivalentTo(e.isDirOrSubmodulePathMatched))
	}

	DescribeTable("empty base path", itBodyFunc,
		Entry("equal path", entry{
			basePath:                    "",
			testPath:                    "",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("subpath", entry{
			basePath:                    "",
			testPath:                    "path",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)

	DescribeTable("non-empty base path", itBodyFunc,
		Entry("non-crossing paths", entry{
			testPath:                    "path",
			basePath:                    "dir",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("parent directory test path", entry{
			testPath:                    "",
			basePath:                    "path",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("equal paths", entry{
			testPath:                    "path",
			basePath:                    "path",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("subpath", entry{
			testPath:                    "path/sub-path",
			basePath:                    "path",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)

	It("true/false matcher", func() {
		matcher := newBasePathMatcher("", NewTruePathMatcher())

		Ω(matcher.IsPathMatched("any")).Should(BeTrue())
		Ω(matcher.ShouldGoThrough("any")).Should(BeFalse())
		Ω(matcher.IsDirOrSubmodulePathMatched("any")).Should(BeTrue())

		matcher = newBasePathMatcher("", NewFalsePathMatcher())

		Ω(matcher.IsPathMatched("any")).Should(BeFalse())
		Ω(matcher.ShouldGoThrough("any")).Should(BeFalse())
		Ω(matcher.IsDirOrSubmodulePathMatched("any")).Should(BeFalse())
	})
})
