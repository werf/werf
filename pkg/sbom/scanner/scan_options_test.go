package scanner

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("ScanOptions", func() {
	Describe("DefaultSyftScanOptions()", func() {
		It("should work", func() {
			Expect(DefaultSyftScanOptions()).To(Equal(ScanOptions{
				Image:      "ghcr.io/anchore/syft:v1.23.1",
				PullPolicy: PullIfMissing,
				Commands: []ScanCommand{
					NewSyftScanCommand(),
				},
			}))
		})
	})

	DescribeTable("Checksum()",
		func(scanOpts ScanOptions, expected types.GomegaMatcher) {
			Expect(scanOpts.Checksum()).To(expected)
		},
		Entry(
			"should work for DefaultSyftScanOptions",
			DefaultSyftScanOptions(),
			Equal("4d3f85d435a14d6152e3542736941e5b46cffaf25c4d1e71fe07474f5b11f790"),
		),
	)
})
