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
			"10", &UnitValue{Value: 10, isAbsolute: false, isSI: false}, BeNil(),
		),
		Entry("Percentage: zero",
			"0", &UnitValue{Value: 0, isAbsolute: false, isSI: false}, BeNil(),
		),
		Entry("Percentage: hundred",
			"100", &UnitValue{Value: 100, isAbsolute: false, isSI: false}, BeNil(),
		),
		Entry("Percentage: over hundred",
			"101", nil, MatchError("percentage value 101 cannot exceed 100"),
		),
		Entry("Bytes: GB (SI)",
			"10GB", &UnitValue{Value: 10 * 1000 * 1000 * 1000, isAbsolute: true, isSI: true}, BeNil(),
		),
		Entry("Bytes: GiB (Binary)",
			"10GiB", &UnitValue{Value: 10 * 1024 * 1024 * 1024, isAbsolute: true, isSI: false}, BeNil(),
		),
		Entry("Bytes: MB (SI)",
			"500MB", &UnitValue{Value: 500 * 1000 * 1000, isAbsolute: true, isSI: true}, BeNil(),
		),
		Entry("Bytes: MiB (Binary)",
			"500MiB", &UnitValue{Value: 500 * 1024 * 1024, isAbsolute: true, isSI: false}, BeNil(),
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
		func(sv *UnitValue, total, expected uint64) {
			Expect(sv.ToBytes(total)).To(Equal(expected))
		},
		Entry("From percentage 10% of 1000",
			&UnitValue{Value: 10, isAbsolute: false},
			uint64(1000),
			uint64(100),
		),
		Entry("From absolute bytes 500 of 1000",
			&UnitValue{Value: 500, isAbsolute: true},
			uint64(1000),
			uint64(500),
		),
	)

	DescribeTable("String",
		func(sv *UnitValue, expected string) {
			Expect(sv.String()).To(Equal(expected))
		},
		Entry("Format percentage",
			&UnitValue{Value: 70, isAbsolute: false},
			"70",
		),
		Entry("Format absolute SI (GB)",
			&UnitValue{Value: 10 * 1000 * 1000 * 1000, isAbsolute: true, isSI: true},
			"10 GB",
		),
		Entry("Format absolute Binary (GiB)",
			&UnitValue{Value: 10 * 1024 * 1024 * 1024, isAbsolute: true, isSI: false},
			"10 GiB",
		),
		Entry("Format absolute SI (MB)",
			&UnitValue{Value: 500 * 1000 * 1000, isAbsolute: true, isSI: true},
			"500 MB",
		),
	)

	Describe("IsAbsolute", func() {
		It("should return true for absolute values", func() {
			v := &UnitValue{isAbsolute: true}
			Expect(v.IsAbsolute()).To(BeTrue())
		})

		It("should return false for percentages", func() {
			v := &UnitValue{isAbsolute: false}
			Expect(v.IsAbsolute()).To(BeFalse())
		})
	})
})
