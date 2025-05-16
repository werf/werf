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
			Equal("f2b172aa9b952cfba7ae9914e7e5a9760ff0d2c7d5da69d09195c63a2577da79"),
		),
	)
})
