package path_matcher

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("false path matcher", func() {
	It("all", func() {
		matcher := NewFalsePathMatcher()

		Ω(matcher.IsPathMatched("any")).Should(BeFalse())
		Ω(matcher.ShouldGoThrough("any")).Should(BeFalse())
		Ω(matcher.IsPathMatched("any")).Should(BeFalse())
	})
})
