package image

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("LayerCompression", func() {
	DescribeTable("Validate accepts supported settings",
		func(compression LayerCompression) {
			Expect(compression.Validate()).To(Succeed())
		},
		Entry("default", LayerCompression{}),
		Entry("gzip level bounds", LayerCompression{Format: LayerCompressionGzip, Level: lo.ToPtr(9)}),
		Entry("level without format is a gzip level", LayerCompression{Level: lo.ToPtr(0)}),
		Entry("zstd level bounds", LayerCompression{Format: LayerCompressionZstd, Level: lo.ToPtr(22)}),
		Entry("none", LayerCompression{Format: LayerCompressionNone}),
	)

	DescribeTable("Validate rejects unsupported settings",
		func(compression LayerCompression, expectedError string) {
			Expect(compression.Validate()).To(MatchError(ContainSubstring(expectedError)))
		},
		Entry("unknown format", LayerCompression{Format: "lz4"}, `unexpected compression format "lz4"`),
		Entry("gzip level above the range", LayerCompression{Level: lo.ToPtr(10)}, "expected 0-9 for the gzip compression format"),
		Entry("zstd level below the range", LayerCompression{Format: LayerCompressionZstd, Level: lo.ToPtr(0)}, "expected 1-22 for the zstd compression format"),
		Entry("none with a level", LayerCompression{Format: LayerCompressionNone, Level: lo.ToPtr(0)}, "the none compression format has no level"),
	)

	DescribeTable("DigestInput keeps the default out of stage digests",
		func(compression LayerCompression, expected string) {
			Expect(compression.DigestInput()).To(Equal(expected))
		},
		Entry("default", LayerCompression{}, ""),
		Entry("explicit gzip", LayerCompression{Format: LayerCompressionGzip}, ""),
		Entry("gzip level", LayerCompression{Level: lo.ToPtr(1)}, "gzip:1"),
		Entry("zstd", LayerCompression{Format: LayerCompressionZstd}, "zstd"),
		Entry("zstd level", LayerCompression{Format: LayerCompressionZstd, Level: lo.ToPtr(19)}, "zstd:19"),
		Entry("none", LayerCompression{Format: LayerCompressionNone}, "none"),
	)
})
