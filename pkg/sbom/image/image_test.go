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

	DescribeTable("ImageName",
		func(name, expected string) {
			Expect(ImageName(name)).To(Equal(expected))
		},
		Entry("simple image", "myapp:v1", "myapp:v1-sbom"),
		Entry("registry image", "registry.io/image:tag", "registry.io/image:tag-sbom"),
	)

	DescribeTable("BaseImageSbomName",
		func(repo, tag, expected string) {
			Expect(BaseImageName(repo, tag)).To(Equal(expected))
		},
		Entry("registry image", "registry.io/image", "v1", "registry.io/image:v1-sbom"),
		Entry("localhost image", "localhost:5000/app", "latest", "localhost:5000/app:latest-sbom"),
	)
})
