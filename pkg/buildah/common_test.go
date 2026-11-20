package buildah

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"go.podman.io/image/v5/manifest"
	"go.podman.io/image/v5/pkg/compression"

	"github.com/werf/werf/v3/pkg/image"
)

var _ = Describe("push compression", func() {
	DescribeTable("newPushCompression",
		func(format string, level *int, expectedAlgorithm, expectedManifestType string, expectedForce bool) {
			compressionOpts, err := newPushCompression(image.LayerCompression{Format: format, Level: level})
			Expect(err).NotTo(HaveOccurred())
			if expectedAlgorithm == "" {
				Expect(compressionOpts.Algorithm).To(BeNil())
			} else {
				Expect(compressionOpts.Algorithm).NotTo(BeNil())
				Expect(compressionOpts.Algorithm.Name()).To(Equal(expectedAlgorithm))
			}
			Expect(compressionOpts.Level).To(Equal(level))
			Expect(compressionOpts.ManifestType).To(Equal(expectedManifestType))
			Expect(compressionOpts.Force).To(Equal(expectedForce))
		},
		Entry("empty format forces gzip with the Docker manifest", "", nil, compression.Gzip.Name(), manifest.DockerV2Schema2MediaType, true),
		Entry("empty format with level forces gzip with the Docker manifest", "", lo.ToPtr(9), compression.Gzip.Name(), manifest.DockerV2Schema2MediaType, true),
		Entry("explicit gzip forces gzip with the Docker manifest", image.LayerCompressionGzip, nil, compression.Gzip.Name(), manifest.DockerV2Schema2MediaType, true),
		Entry("gzip with level forces gzip with the Docker manifest", image.LayerCompressionGzip, lo.ToPtr(0), compression.Gzip.Name(), manifest.DockerV2Schema2MediaType, true),
		Entry("zstd switches to the OCI manifest", image.LayerCompressionZstd, nil, compression.Zstd.Name(), "application/vnd.oci.image.manifest.v1+json", true),
		Entry("zstd with level switches to the OCI manifest", image.LayerCompressionZstd, lo.ToPtr(22), compression.Zstd.Name(), "application/vnd.oci.image.manifest.v1+json", true),
		Entry("none switches to the OCI manifest", image.LayerCompressionNone, nil, "", "application/vnd.oci.image.manifest.v1+json", false),
	)

	DescribeTable("newPushCompression rejects what LayerCompression.Validate rejects",
		func(layerCompression image.LayerCompression) {
			_, err := newPushCompression(layerCompression)
			Expect(err).To(MatchError(layerCompression.Validate().Error()))
		},
		Entry("unknown format", image.LayerCompression{Format: "lz4"}),
		Entry("format is case-sensitive", image.LayerCompression{Format: "GZIP"}),
		Entry("none with a level", image.LayerCompression{Format: image.LayerCompressionNone, Level: lo.ToPtr(0)}),
		Entry("default format validates the level as gzip", image.LayerCompression{Level: lo.ToPtr(10)}),
		Entry("zstd level below the minimum", image.LayerCompression{Format: image.LayerCompressionZstd, Level: lo.ToPtr(0)}),
	)
})
