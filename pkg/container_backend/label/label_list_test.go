package label_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend/label"
)

var _ = Describe("LabelList", func() {
	DescribeTable("Add",
		func(list label.LabelList, item label.Label, expectedList label.LabelList) {
			list.Add(item)
			Expect(list).To(Equal(expectedList))
		},
		Entry(
			"should modify list in place",
			label.LabelList{},
			label.NewLabel("foo", "bar"),
			label.LabelList{
				label.NewLabel("foo", "bar"),
			},
		),
		Entry(
			"should prevent duplication",
			label.LabelList{
				label.NewLabel("foo", "bar"),
			},
			label.NewLabel("foo", "bar"),
			label.LabelList{
				label.NewLabel("foo", "bar"),
			},
		),
	)

	DescribeTable("Remove",
		func(list label.LabelList, item label.Label, expectedList label.LabelList) {
			list.Remove(item)
			Expect(list).To(Equal(expectedList))
		},
		Entry(
			"should not modify list if item not found",
			label.LabelList{
				label.NewLabel("foo", "bar"),
			},
			label.NewLabel("foo", "1"),
			label.LabelList{
				label.NewLabel("foo", "bar"),
			},
		),
		Entry(
			"should remove item from the list if found",
			label.LabelList{
				label.NewLabel("foo", "bar"),
			},
			label.NewLabel("foo", "bar"),
			label.LabelList{},
		),
	)

	DescribeTable("ToStringSlice",
		func(input label.LabelList, expected []string) {
			Expect(input.ToStringSlice()).To(Equal(expected))
		},
		Entry(
			"should work",
			label.LabelList{
				label.NewLabel("foo", "bar"),
				label.NewLabel("baz", "wow"),
			},
			[]string{"foo=bar", "baz=wow"},
		),
	)

	DescribeTable("NewLabelListFromMap",
		func(input map[string]string, expected label.LabelList) {
			Expect(label.NewLabelListFromMap(input).ToStringSlice()).Should(ConsistOf(expected.ToStringSlice()))
		},
		Entry(
			"should work",
			map[string]string{
				"foo":  "bar",
				"test": "some",
			},
			label.LabelList{
				label.NewLabel("foo", "bar"),
				label.NewLabel("test", "some"),
			},
		),
	)
})
