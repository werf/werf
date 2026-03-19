package buildah

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
)

var _ = Describe("buildah", func() {
	DescribeTable("mapBackendOldFiltersToBuildahImageFilters",
		func(oldFilters []util.Pair[string, string], expectedFilters []string) {
			actual := mapBackendOldFiltersToBuildahImageFilters(oldFilters)
			Expect(actual).To(Equal(expectedFilters))
		},
		Entry(
			"should work with empty input",
			[]util.Pair[string, string]{},
			[]string{},
		),
		Entry(
			"should work with non-empty input",
			[]util.Pair[string, string]{
				util.NewPair("foo", "bar"),
				util.NewPair("key", "value"),
			},
			[]string{"foo=bar", "key=value"},
		),
	)
})
