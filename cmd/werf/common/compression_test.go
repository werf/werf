package common

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = ginkgo.DescribeTable("Layer compression options", func(env map[string]string, args []string, expectedFormat string, expectedLevel *int, expectedError string) {
	clearListEnv()
	for name, value := range env {
		ginkgo.GinkgoT().Setenv(name, value)
	}

	opts, err := getCompressionOptions(compressionCmdData(args...))

	if expectedError != "" {
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedError)))
		return
	}

	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(opts.Format).To(gomega.Equal(expectedFormat))
	if expectedLevel == nil {
		gomega.Expect(opts.Level).To(gomega.BeNil())
		return
	}
	gomega.Expect(opts.Level).NotTo(gomega.BeNil())
	gomega.Expect(*opts.Level).To(gomega.Equal(*expectedLevel))
},
	ginkgo.Entry("unspecified options leave the backend defaults alone", nil, nil, "", nil, ""),
	ginkgo.Entry("format flag", nil, []string{"--compression-format", "zstd"}, "zstd", nil, ""),
	ginkgo.Entry("format environment", map[string]string{"WERF_COMPRESSION_FORMAT": "zstd"}, nil, "zstd", nil, ""),
	ginkgo.Entry("format flag wins over environment", map[string]string{"WERF_COMPRESSION_FORMAT": "zstd"}, []string{"--compression-format", "gzip"}, "gzip", nil, ""),
	ginkgo.Entry("zero level is an explicit level", nil, []string{"--compression-level", "0"}, "", lo.ToPtr(0), ""),
	ginkgo.Entry("level environment", map[string]string{"WERF_COMPRESSION_LEVEL": "5"}, nil, "", lo.ToPtr(5), ""),
	ginkgo.Entry("level flag wins over environment", map[string]string{"WERF_COMPRESSION_LEVEL": "5"}, []string{"--compression-level", "1"}, "", lo.ToPtr(1), ""),
	ginkgo.Entry("max zstd level", nil, []string{"--compression-format", "zstd", "--compression-level", "22"}, "zstd", lo.ToPtr(22), ""),
	ginkgo.Entry("none format", nil, []string{"--compression-format", "none"}, "none", nil, ""),
	ginkgo.Entry("unknown format", nil, []string{"--compression-format", "lz4"}, "", nil, `unexpected compression format "lz4"`),
	ginkgo.Entry("none format with a level", nil, []string{"--compression-format", "none", "--compression-level", "0"}, "", nil, "the none compression format has no level"),
	ginkgo.Entry("none format from environment with a level", map[string]string{"WERF_COMPRESSION_FORMAT": "none", "WERF_COMPRESSION_LEVEL": "1"}, nil, "", nil, "the none compression format has no level"),
	ginkgo.Entry("non-integer level", nil, []string{"--compression-level", "fast"}, "", nil, "unexpected compression level specified: fast"),
	ginkgo.Entry("gzip level above the gzip range", nil, []string{"--compression-level", "10"}, "", nil, "expected 0-9 for the gzip compression format"),
	ginkgo.Entry("zstd level below the zstd range", nil, []string{"--compression-format", "zstd", "--compression-level", "0"}, "", nil, "expected 1-22 for the zstd compression format"),
	ginkgo.Entry("gzip level above the gzip range from environment", map[string]string{"WERF_COMPRESSION_LEVEL": "10"}, nil, "", nil, "expected 0-9 for the gzip compression format"),
)

var _ = ginkgo.Describe("Commands without layer compression flags", func() {
	ginkgo.It("do not fail on the missing registration", func() {
		clearListEnv()

		opts, err := getCompressionOptions(&CmdData{})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(opts.Format).To(gomega.BeEmpty())
		gomega.Expect(opts.Level).To(gomega.BeNil())
	})
})
