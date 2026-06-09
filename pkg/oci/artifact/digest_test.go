package artifact_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/oci/artifact"
)

var _ = Describe("DigestHex", func() {
	DescribeTable("should parse various digest formats",
		func(input, expected string) {
			result, err := artifact.DigestHex(input)
			Expect(err).To(Succeed())
			Expect(result).To(Equal(expected))
		},
		Entry("sha256 digest", "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"),
		Entry("sha512 digest", "sha512:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
	)

	DescribeTable("should error on malformed digests",
		func(input string) {
			_, err := artifact.DigestHex(input)
			Expect(err).To(HaveOccurred())
		},
		Entry("empty string", ""),
		Entry("missing algorithm prefix", "abc123"),
		Entry("invalid hex length", "sha256:abc123"),
		Entry("invalid characters", "sha256:xyz!"),
	)
})

var _ = Describe("FallbackTag", func() {
	DescribeTable("should compute correct fallback tag",
		func(digest, expected string) {
			Expect(artifact.FallbackTag(digest)).To(Equal(expected))
		},
		Entry("sha256 digest", "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "sha256-e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"),
	)

	DescribeTable("should handle malformed digests gracefully (best-effort)",
		func(digest, expected string) {
			Expect(artifact.FallbackTag(digest)).To(Equal(expected))
		},
		Entry("plain hex without prefix", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "sha256-e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"),
		Entry("digest with colons", "sha256:abc:def", "sha256-abc-def"),
		Entry("digest with slashes", "sha256:abc/def", "sha256-abc_def"),
		Entry("digest with at sign", "sha256:abc@def", "sha256-abc-def"),
	)
})
