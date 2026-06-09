package image

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Image", func() {
	DescribeTable("IsScratchRef",
		func(imageRef string, expected bool) {
			Expect(IsScratchRef(imageRef)).To(Equal(expected))
		},
		Entry("bare scratch", "scratch", true),

		Entry("registry scratch with tag", "registry.werf.io/werf/scratch:latest", true),
		Entry("registry scratch with custom tag", "registry.werf.io/werf/scratch:some-valid-tag", true),

		Entry("registry scratch with digest", "registry.werf.io/werf/scratch@sha256:5d68d4300015200b8797ddf93a5dee3491fd2f6c0211d70a6ab8127ea053375a", true),

		Entry("localhost scratch with tag", "localhost:5000/scratch:latest", true),
		Entry("localhost scratch with digest", "localhost:5000/scratch@sha256:abc123def456abc123def456abc123def456abc123def456abc123def456abc1", true),

		Entry("ubuntu", "ubuntu:22.04", false),
		Entry("registry image", "registry.werf.io/base/ubuntu:22.04", false),
		Entry("image with scratch in name", "my-scratch-image:latest", false),
		Entry("scratch prefix", "scratchpad:latest", false),
		Entry("scratch in path", "registry.werf.io/scratch-images/app:latest", false),

		Entry("empty string", "", false),
	)

	DescribeTable("FallbackTag",
		func(digest, expected string) {
			Expect(FallbackTag(digest)).To(Equal(expected))
		},
		Entry("standard digest", "sha256:5d68d4300015200b8797ddf93a5dee3491fd2f6c0211d70a6ab8127ea053375a", "sha256-5d68d4300015200b8797ddf93a5dee3491fd2f6c0211d70a6ab8127ea053375a"),
		Entry("alternate digest", "sha256:abc123def456abc123def456abc123def456abc123def456abc123def456abc1", "sha256-abc123def456abc123def456abc123def456abc123def456abc123def456abc1"),
	)
})
