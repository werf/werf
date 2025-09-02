package host_cleaning

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
)

var _ = Describe("LocalBackendCleaner helpers", func() {
	DescribeTable("countImagesToFree",
		func(images image.ImagesList, startIndex int, bytesToFree uint64, expected int) {
			actual := countImagesToFree(images, startIndex, bytesToFree)
			Expect(actual).To(Equal(expected))
		},
		Entry("should work if bytesToFree = 0",
			image.ImagesList{},
			1,
			uint64(0),
			-1,
		),
		Entry("should work if startIndex < 0",
			image.ImagesList{},
			-1,
			uint64(100),
			-1,
		),
		Entry("should work if startIndex >= len(list)",
			image.ImagesList{},
			1,
			uint64(100),
			-1,
		),
		Entry("should work if bytesToFree > 0 AND bytesToFree < sum(img.Size)",
			image.ImagesList{
				{Size: 0},
				{Size: 150},
				{Size: 200},
				{Size: 250},
			},
			1,
			uint64(300),
			2,
		),
		Entry("should work if bytesToFree > sum(img.Size)",
			image.ImagesList{
				{Size: 100},
				{Size: 150},
			},
			0,
			uint64(300),
			1,
		),
		Entry("should work if bytesToFree = sum(img.Size)",
			image.ImagesList{
				{Size: 100},
				{Size: 200},
			},
			0,
			uint64(300),
			1,
		),
		Entry("should work if image list has no items",
			image.ImagesList{},
			0,
			uint64(150),
			-1,
		),
	)
})
