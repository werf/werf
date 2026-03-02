package image

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = Describe("ManifestCache.DeleteImageInfo", func() {
	var (
		cache       *ManifestCache
		storageName string
	)

	BeforeEach(func() {
		storageName = "test-storage"

		tempDir := GinkgoT().TempDir()
		Expect(werf.Init("", tempDir)).To(Succeed())

		cache = NewManifestCache(tempDir)
	})

	It("should delete an existing cache entry", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)

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

	It("should not error when deleting a non-existent entry", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)

		Expect(cache.DeleteImageInfo(ctx, storageName, "non-existent-image")).To(Succeed())
	})
})
