package build

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
	"github.com/werf/werf/v2/pkg/sbom/scanner"
)

var _ = Describe("SbomStep Checksum", func() {
	DescribeTable("calculateStableChecksum",
		func(scanOpts scanner.ScanOptions, mergeOpts cyclonedxutil.MergeOpts, expectedChecksum string) {
			step := &sbomStep{}
			checksum := step.calculateStableChecksum(scanOpts, mergeOpts)
			Expect(checksum).To(Equal(expectedChecksum))
		},

		Entry("empty options",
			scanner.ScanOptions{},
			cyclonedxutil.MergeOpts{},
			"aa969eabe2faad149265a94e60b173e527e0bc27898afcd0ec4e85a06b28f29b",
		),

		Entry("empty options with GOST configuration (should be invariant)",
			scanner.ScanOptions{},
			cyclonedxutil.MergeOpts{
				Gost: gost.Config{
					AttackSurface:    gost.GostValueYes,
					SecurityFunction: gost.GostValueIndirect,
				},
			},
			"aa969eabe2faad149265a94e60b173e527e0bc27898afcd0ec4e85a06b28f29b",
		),
	)
})
