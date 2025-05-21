package filter_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend/filter"
)

var _ = Describe("FilterList", func() {
	DescribeTable("Add",
		func(list filter.FilterList, item filter.Filter, expectedList filter.FilterList) {
			list.Add(item)
			Expect(list).To(Equal(expectedList))
		},
		Entry(
			"should modify list in place",
			filter.FilterList{},
			filter.NewFilter("foo", "bar"),
			filter.FilterList{
				filter.NewFilter("foo", "bar"),
			},
		),
		Entry(
			"should prevent duplication",
			filter.FilterList{
				filter.NewFilter("foo", "bar"),
			},
			filter.NewFilter("foo", "bar"),
			filter.FilterList{
				filter.NewFilter("foo", "bar"),
			},
		),
	)

	DescribeTable("Remove",
		func(list filter.FilterList, item filter.Filter, expectedList filter.FilterList) {
			list.Remove(item)
			Expect(list).To(Equal(expectedList))
		},
		Entry(
			"should not modify list if item not found",
			filter.FilterList{
				filter.NewFilter("foo", "bar"),
			},
			filter.NewFilter("foo", "1"),
			filter.FilterList{
				filter.NewFilter("foo", "bar"),
			},
		),
		Entry(
			"should remove item from the list if found",
			filter.FilterList{
				filter.NewFilter("foo", "bar"),
			},
			filter.NewFilter("foo", "bar"),
			filter.FilterList{},
		),
	)

	DescribeTable("ToStringSlice",
		func(input filter.FilterList, expected []string) {
			Expect(input.ToStringSlice()).To(Equal(expected))
		},
		Entry(
			"should work",
			filter.FilterList{
				filter.NewFilter("foo", "bar"),
				filter.NewFilter("baz", "wow"),
			},
			[]string{"foo=bar", "baz=wow"},
		),
	)
})
