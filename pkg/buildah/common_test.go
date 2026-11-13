package buildah

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"go.podman.io/image/v5/manifest"
	"go.podman.io/image/v5/pkg/compression"
)

var _ = Describe("push compression", func() {
	DescribeTable("newPushCompression",
		func(format string, level *int, expectedAlgorithm, expectedManifestType string, expectedForce bool) {
			compressionOpts, err := newPushCompression(format, level)
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
		Entry("explicit gzip forces gzip with the Docker manifest", CompressionFormatGzip, nil, compression.Gzip.Name(), manifest.DockerV2Schema2MediaType, true),
		Entry("gzip with level forces gzip with the Docker manifest", CompressionFormatGzip, lo.ToPtr(0), compression.Gzip.Name(), manifest.DockerV2Schema2MediaType, true),
		Entry("zstd switches to the OCI manifest", CompressionFormatZstd, nil, compression.Zstd.Name(), "application/vnd.oci.image.manifest.v1+json", true),
		Entry("zstd with level switches to the OCI manifest", CompressionFormatZstd, lo.ToPtr(22), compression.Zstd.Name(), "application/vnd.oci.image.manifest.v1+json", true),
		Entry("none switches to the OCI manifest", CompressionFormatNone, nil, "", "application/vnd.oci.image.manifest.v1+json", false),
	)

	DescribeTable("newPushCompression rejects unsupported values",
		func(format string, level *int, expectedError string) {
			_, err := newPushCompression(format, level)
			Expect(err).To(MatchError(ContainSubstring(expectedError)))
		},
		Entry("unknown format", "lz4", nil, `unsupported compression format "lz4"`),
		Entry("none with a level", CompressionFormatNone, lo.ToPtr(0), `unsupported compression level 0 for compression format "none"`),
		Entry("unknown format with a valid level", "lz4", lo.ToPtr(1), `unsupported compression format "lz4"`),
		Entry("format is case-sensitive", "GZIP", nil, `unsupported compression format "GZIP"`),
		Entry("gzip level above the maximum", CompressionFormatGzip, lo.ToPtr(10), `unsupported compression level 10 for compression format "gzip": expected a value between 0 and 9`),
		Entry("gzip level below the minimum", CompressionFormatGzip, lo.ToPtr(-1), `unsupported compression level -1 for compression format "gzip"`),
		Entry("default format validates the level as gzip", "", lo.ToPtr(10), `unsupported compression level 10 for compression format "gzip"`),
		Entry("zstd level above the maximum", CompressionFormatZstd, lo.ToPtr(23), `unsupported compression level 23 for compression format "zstd": expected a value between 1 and 22`),
		Entry("zstd level below the minimum", CompressionFormatZstd, lo.ToPtr(0), `unsupported compression level 0 for compression format "zstd"`),
	)

	It("rejects compression options before initializing the backend", func() {
		_, err := NewBuildah(ModeNative, BuildahOpts{
			CommonBuildahOpts: CommonBuildahOpts{CompressionFormat: "lz4"},
		})
		Expect(err).To(MatchError(ContainSubstring(`unsupported compression format "lz4"`)))
	})

	It("rejects an out of range compression level before initializing the backend", func() {
		_, err := NewBuildah(ModeNative, BuildahOpts{
			CommonBuildahOpts: CommonBuildahOpts{CompressionFormat: CompressionFormatZstd, CompressionLevel: lo.ToPtr(100)},
		})
		Expect(err).To(MatchError(ContainSubstring(`unsupported compression level 100 for compression format "zstd"`)))
	})
})
