package docker

import (
	"archive/tar"
	"bytes"
	"io"

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
	// emptyTarArchive provides the rootfs body sent to the Docker Engine import
	// API for `from: scratch` images. A malformed/empty body was the root cause
	// of unreadable scratch images on Linux, so the body must be a valid tar.
	Describe("emptyTarArchive", func() {
		It("returns a valid, non-nil, empty tar archive", func() {
			r, err := emptyTarArchive()
			Expect(err).NotTo(HaveOccurred())
			Expect(r).NotTo(BeNil())

			data, err := io.ReadAll(r)
			Expect(err).NotTo(HaveOccurred())
			// A valid empty tar is not a zero-byte stream: it contains the
			// end-of-archive marker (two 512-byte zero blocks).
			Expect(len(data)).To(BeNumerically(">", 0))

			tr := tar.NewReader(bytes.NewReader(data))
			_, err = tr.Next()
			Expect(err).To(Equal(io.EOF), "expected a well-formed empty tar with no entries")
		})
	})
})
