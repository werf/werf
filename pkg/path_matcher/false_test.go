package path_matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("false path matcher", func() {
	It("all", func() {
		matcher := NewFalsePathMatcher()

		Expect(matcher.IsPathMatched("any")).Should(BeFalse())
		Expect(matcher.ShouldGoThrough("any")).Should(BeFalse())
		Expect(matcher.IsPathMatched("any")).Should(BeFalse())
	})
})
