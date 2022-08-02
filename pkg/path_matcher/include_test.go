package path_matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("include path matcher", func() {
	type entry struct {
		includeGlobs                []string
		testPath                    string
		isPathMatched               bool
		shouldGoThrough             bool
		isDirOrSubmodulePathMatched bool
	}

	itBodyFunc := func(e entry) {
		matcher := newIncludePathMatcher(e.includeGlobs)

		Ω(matcher.IsPathMatched(e.testPath)).Should(BeEquivalentTo(e.isPathMatched))
		Ω(matcher.ShouldGoThrough(e.testPath)).Should(BeEquivalentTo(e.shouldGoThrough))
		Ω(matcher.IsDirOrSubmodulePathMatched(e.testPath)).Should(BeEquivalentTo(e.isDirOrSubmodulePathMatched))
	}

	DescribeTable("empty include globs list", itBodyFunc,
		Entry("empty test path", entry{
			includeGlobs:                nil,
			testPath:                    "",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)

	DescribeTable("non-empty include globs list", itBodyFunc,
		Entry("empty test path (1)", entry{
			includeGlobs:                []string{""},
			testPath:                    "",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("empty test path (2)", entry{
			includeGlobs:                []string{"dir"},
			testPath:                    "",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("matched test path (1)", entry{
			includeGlobs:                []string{"dir"},
			testPath:                    "dir",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("matched test path (2)", entry{
			includeGlobs:                []string{"dir"},
			testPath:                    "dir/any",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("not matched test path (1)", entry{
			includeGlobs:                []string{"dir1"},
			testPath:                    "dir2",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("not matched test path (2)", entry{
			includeGlobs:                []string{"dir/file"},
			testPath:                    "dir",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("glob completion (1)", entry{
			includeGlobs:                []string{"a"},
			testPath:                    "a/b/c/d",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("glob completion (2)", entry{
			includeGlobs:                []string{"a/*/*/**/*"},
			testPath:                    "a/d/c/d",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)
})
