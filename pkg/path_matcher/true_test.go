package path_matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("true path matcher", func() {
	It("all", func() {
		matcher := NewTruePathMatcher()

		Ω(matcher.IsPathMatched("any")).Should(BeTrue())
		Ω(matcher.ShouldGoThrough("any")).Should(BeFalse())
		Ω(matcher.IsPathMatched("any")).Should(BeTrue())
	})
})
