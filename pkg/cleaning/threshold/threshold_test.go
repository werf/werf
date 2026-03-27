package threshold

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("threshold", func() {
	DescribeTable("Parse",
		func(value string, expected Threshold, expectsError bool) {
			th, err := Parse(value)
			if expectsError {
				Expect(err).To(HaveOccurred())
				return
			}

			Expect(err).NotTo(HaveOccurred())
			Expect(th).To(Equal(expected))
		},
		Entry("percentage 70", "70", NewPercentage(70), false),
		Entry("percentage 0", "0", NewPercentage(0), false),
		Entry("percentage 100", "100", NewPercentage(100), false),
		Entry("bytes gigabytes", "10GB", NewBytes(10_000_000_000), false),
		Entry("bytes megabytes short b", "10Mb", NewBytes(10_000_000), false),
		Entry("invalid text", "wat", Threshold{}, true),
		Entry("invalid percent sign", "101%", Threshold{}, true),
	)

	Describe("Resolve", func() {
		It("uses default percentage margin", func() {
			threshold, margin, err := Resolve(ptr(NewPercentage(70)), nil, NewPercentage(70), NewPercentage(5), false, "--foo", "--bar")
			Expect(err).NotTo(HaveOccurred())
			Expect(threshold).To(Equal(NewPercentage(70)))
			Expect(margin).To(Equal(NewPercentage(5)))
		})

		It("uses zero-bytes implicit margin for bytes thresholds", func() {
			threshold, margin, err := Resolve(ptr(NewBytes(10_000_000_000)), nil, NewPercentage(70), NewPercentage(5), false, "--foo", "--bar")
			Expect(err).NotTo(HaveOccurred())
			Expect(threshold).To(Equal(NewBytes(10_000_000_000)))
			Expect(margin).To(Equal(NewBytes(0)))
		})

		It("returns an error for explicitly mixed formats", func() {
			_, _, err := Resolve(ptr(NewBytes(10_000_000_000)), ptr(NewPercentage(5)), NewPercentage(70), NewPercentage(5), true, "--foo", "--bar")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must use the same format"))
		})
	})
})

func ptr[T any](v T) *T { return &v }
