package path_matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("true path matcher", func() {
	It("all", func() {
		matcher := NewTruePathMatcher()

		Expect(matcher.IsPathMatched("any")).Should(BeTrue())
		Expect(matcher.ShouldGoThrough("any")).Should(BeFalse())
		Expect(matcher.IsPathMatched("any")).Should(BeTrue())
	})
})
