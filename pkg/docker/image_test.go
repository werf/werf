package docker

import (
	"github.com/docker/docker/api/types/filters"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
)

var _ = Describe("docker images", func() {
	DescribeTable("mapImagesPruneOptionsToImagesPruneFilters",
		func(opts ImagesPruneOptions, expected filters.Args) {
			actual := mapImagesPruneOptionsToImagesPruneFilters(opts)
			Expect(actual).To(Equal(expected))
		},
		Entry(
			"should work with empty filters",
			ImagesPruneOptions{},
			filters.NewArgs(),
		),
		Entry("should work with 'label' filter",
			ImagesPruneOptions{
				Filters: []util.Pair[string, string]{
					util.NewPair("label", "foo=bar"),
				},
			},
			filters.NewArgs(
				filters.Arg("label", "foo=bar"),
			),
		),
	)
})
