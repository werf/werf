package units

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("UnitValue", func() {
	DescribeTable("ParseUnitValue",
		func(input string, expected *UnitValue, envErrMatcher types.GomegaMatcher) {
			res, err := ParseUnitValue(input)
			Expect(err).To(envErrMatcher)
			Expect(res).To(Equal(expected))
		},
		Entry("Percentage: valid numeric",
			"10", &UnitValue{Value: 10, IsBytes: false}, BeNil(),
		),
		Entry("Percentage: zero",
			"0", &UnitValue{Value: 0, IsBytes: false}, BeNil(),
		),
		Entry("Percentage: hundred",
			"100", &UnitValue{Value: 100, IsBytes: false}, BeNil(),
		),
		Entry("Percentage: over hundred",
			"101", nil, MatchError("percentage value 101 cannot exceed 100"),
		),
		Entry("Bytes: GB",
			"10GB", &UnitValue{Value: 10 * 1024 * 1024 * 1024, IsBytes: true}, BeNil(),
		),
		Entry("Bytes: GiB",
			"10GiB", &UnitValue{Value: 10 * 1024 * 1024 * 1024, IsBytes: true}, BeNil(),
		),
		Entry("Bytes: MB",
			"500MB", &UnitValue{Value: 500 * 1024 * 1024, IsBytes: true}, BeNil(),
		),
		Entry("Invalid: empty",
			"", nil, MatchError("empty storage value"),
		),
		Entry("Invalid: random string", "foo", nil,
			MatchError(
				And(
					ContainSubstring("specify percentage (0-100"),
					ContainSubstring("or absolute size (e.g. 10GB, 500MiB"),
				),
			),
		),
	)

	DescribeTable("ToBytes",
		func(sv *UnitValue, total uint64, expected uint64) {
			Expect(sv.ToBytes(total)).To(Equal(expected))
		},
		Entry("From percentage 10% of 1000",
			&UnitValue{Value: 10, IsBytes: false},
			uint64(1000),
			uint64(100),
		),
		Entry("From bytes 500 of 1000",
			&UnitValue{Value: 500, IsBytes: true},
			uint64(1000),
			uint64(500),
		),
		Entry("Percentage rounding check 33% of 1000",
			&UnitValue{Value: 33, IsBytes: false},
			uint64(1000),
			uint64(330),
		),
	)

	DescribeTable("String",
		func(sv *UnitValue, expected string) {
			Expect(sv.String()).To(Equal(expected))
		},
		Entry("Format percentage",
			&UnitValue{Value: 70, IsBytes: false},
			"70%",
		),
		Entry("Format bytes (KiB)",
			&UnitValue{Value: 1024, IsBytes: true},
			"1KiB",
		),
		Entry("Format bytes (GB)",
			&UnitValue{Value: 10 * 1024 * 1024 * 1024, IsBytes: true},
			"10GiB",
		),
	)
})
