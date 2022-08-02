package path_matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("exclude path matcher", func() {
	type entry struct {
		excludeGlobs                []string
		testPath                    string
		isPathMatched               bool
		shouldGoThrough             bool
		isDirOrSubmodulePathMatched bool
	}

	itBodyFunc := func(e entry) {
		matcher := newExcludePathMatcher(e.excludeGlobs)

		Ω(matcher.IsPathMatched(e.testPath)).Should(BeEquivalentTo(e.isPathMatched))
		Ω(matcher.ShouldGoThrough(e.testPath)).Should(BeEquivalentTo(e.shouldGoThrough))
		Ω(matcher.IsDirOrSubmodulePathMatched(e.testPath)).Should(BeEquivalentTo(e.isDirOrSubmodulePathMatched))
	}

	DescribeTable("empty exclude globs list", itBodyFunc,
		Entry("empty test path", entry{
			excludeGlobs:                nil,
			testPath:                    "",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)

	DescribeTable("non-empty exclude globs list", itBodyFunc,
		Entry("empty test path (1)", entry{
			excludeGlobs:                []string{""},
			testPath:                    "",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("empty test path (2)", entry{
			excludeGlobs:                []string{"dir"},
			testPath:                    "",
			isPathMatched:               true,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("matched test path (1)", entry{
			excludeGlobs:                []string{"dir1"},
			testPath:                    "dir2",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("matched test path (2)", entry{
			excludeGlobs:                []string{"dir/file"},
			testPath:                    "dir",
			isPathMatched:               true,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("not matched test path (1)", entry{
			excludeGlobs:                []string{"dir"},
			testPath:                    "dir",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("not matched test path (2)", entry{
			excludeGlobs:                []string{"dir"},
			testPath:                    "dir/any",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("glob completion (1)", entry{
			excludeGlobs:                []string{"a"},
			testPath:                    "a/b/c/d",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("glob completion (2)", entry{
			excludeGlobs:                []string{"a/*/*/**/*"},
			testPath:                    "a/d/c/d",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
	)
})
