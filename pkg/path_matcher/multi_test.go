package path_matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("multi path matcher", func() {
	It("should return false if at least one matcher returns false", func() {
		matcher := NewMultiPathMatcher(
			NewTruePathMatcher(),
			NewFalsePathMatcher(),
		)

		Ω(matcher.IsPathMatched("any")).Should(BeFalse())
		Ω(matcher.ShouldGoThrough("any")).Should(BeFalse())
		Ω(matcher.IsDirOrSubmodulePathMatched("any")).Should(BeFalse())

		matcher = NewMultiPathMatcher(
			NewFalsePathMatcher(),
			NewFalsePathMatcher(),
		)

		Ω(matcher.IsPathMatched("any")).Should(BeFalse())
		Ω(matcher.ShouldGoThrough("any")).Should(BeFalse())
		Ω(matcher.IsDirOrSubmodulePathMatched("any")).Should(BeFalse())
	})

	It("should return true if all matchers return true", func() {
		matcher := NewMultiPathMatcher(
			NewTruePathMatcher(),
			NewTruePathMatcher(),
		)

		Ω(matcher.IsPathMatched("any")).Should(BeTrue())
		Ω(matcher.ShouldGoThrough("any")).Should(BeFalse())
		Ω(matcher.IsDirOrSubmodulePathMatched("any")).Should(BeTrue())
	})

	type entry struct {
		path                        string
		isPathMatched               bool
		shouldGoThrough             bool
		isDirOrSubmodulePathMatched bool
	}

	matcher := NewMultiPathMatcher(
		newBasePathMatcher("dir", nil),
		newBasePathMatcher("dir/sub-dir", nil),
	)

	DescribeTable("ShouldGoThrough", func(e entry) {
		Ω(matcher.IsPathMatched(e.path)).Should(BeEquivalentTo(e.isPathMatched))
		Ω(matcher.ShouldGoThrough(e.path)).Should(BeEquivalentTo(e.shouldGoThrough))
		Ω(matcher.IsDirOrSubmodulePathMatched(e.path)).Should(BeEquivalentTo(e.isDirOrSubmodulePathMatched))
	},
		Entry(`""`, entry{
			path:                        "",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("file", entry{
			path:                        "file",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("dir", entry{
			path:                        "dir",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("dir/sub-dir", entry{
			path:                        "dir/sub-dir",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("dir/sub-dir/file", entry{
			path:                        "dir/sub-dir/file",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)
})
