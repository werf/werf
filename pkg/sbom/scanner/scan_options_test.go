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
				Image:      "anchore/syft:v1.23.1",
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
			Equal("0c15bc4e5bd8541138b5b6b7065eb8f641284b4913878d953be46419f50e8ebc"),
		),
	)
})
