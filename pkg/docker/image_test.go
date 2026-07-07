package docker

import (
	"github.com/docker/docker/api/types/filters"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend/filter"
)

var _ = Describe("docker images", func() {
	DescribeTable("mapBackendFiltersToImagesPruneFilters",
		func(opts ImagesPruneOptions, expected filters.Args) {
			actual := mapBackendFiltersToImagesPruneFilters(opts.Filters)
			Expect(actual).To(Equal(expected))
		},
		Entry(
			"should work with empty filters",
			ImagesPruneOptions{},
			filters.NewArgs(),
		),
		Entry("should work with 'label' filter",
			ImagesPruneOptions{
				Filters: filter.FilterList{
					filter.NewFilter("label", "foo=bar"),
				},
			},
			filters.NewArgs(
				filters.Arg("label", "foo=bar"),
			),
		),
	)
})
