package log_sanitize

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SanitizeDockerRateLimit", func() {
	It("should hide credentials in docker rate limit error", func() {
		input := "toomanyrequests: You have reached your pull rate limit as 'username': token. You may increase the limit."

		out := SanitizeDockerRateLimit(input)

		Expect(out).To(ContainSubstring("credentials hidden"))
		Expect(out).NotTo(ContainSubstring("username"))
		Expect(out).NotTo(ContainSubstring("token"))
	})

	It("should return original string if no rate limit message present", func() {
		input := "some random error"

		out := SanitizeDockerRateLimit(input)

		Expect(out).To(Equal(input))
	})
})
