package image

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/werf"
)

var _ = Describe("ManifestCache.DeleteImageInfo", func() {
	var (
		ctx         context.Context
		cache       *ManifestCache
		storageName string
	)

	BeforeEach(func() {
		ctx = context.Background()
		storageName = "test-storage"

		tempDir := GinkgoT().TempDir()
		err := werf.Init("", tempDir)
		Expect(err).NotTo(HaveOccurred())

		cache = NewManifestCache(tempDir)
	})

	It("should delete an existing cache entry", func() {
		imageName := "test-image"

		testInfo := &Info{
			Name:       imageName,
			Repository: "test-repo/test-image",
			Tag:        "tag",
		}

		Expect(cache.StoreImageInfo(ctx, storageName, testInfo)).To(Succeed())

		info, err := cache.GetImageInfo(ctx, storageName, imageName)
		Expect(err).NotTo(HaveOccurred())
		Expect(info).NotTo(BeNil())

		Expect(cache.DeleteImageInfo(ctx, storageName, imageName)).To(Succeed())

		info, err = cache.GetImageInfo(ctx, storageName, imageName)
		Expect(err).NotTo(HaveOccurred())
		Expect(info).To(BeNil())
	})

	It("should not error when deleting a non-existent entry", func() {
		Expect(cache.DeleteImageInfo(ctx, storageName, "non-existent-image")).To(Succeed())
	})
})
