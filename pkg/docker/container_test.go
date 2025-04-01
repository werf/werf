package docker

import (
	"github.com/docker/docker/api/types/filters"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
)

var _ = Describe("docker container", func() {
	DescribeTable("mapContainersPruneOptionsToContainersPruneFilters",
		func(opts ContainersPruneOptions, expected filters.Args) {
			actual := mapContainersPruneOptionsToContainersPruneFilters(opts)
			Expect(actual).To(Equal(expected))
		},
		Entry(
			"should work with empty filters",
			ContainersPruneOptions{},
			filters.NewArgs(),
		),
		Entry("should work with 'until' filter",
			ContainersPruneOptions{
				Filters: []util.Pair[string, string]{
					util.NewPair("until", "1h"),
				},
			},
			filters.NewArgs(
				filters.Arg("until", "1h"),
			),
		),
	)
})
