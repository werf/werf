package image

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/opstats"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = ginkgo.Describe("ManifestCache lookup counters", func() {
	const storageName = "registry.example.com/project"

	var (
		cache     *ManifestCache
		collector *opstats.Collector
		rows      func(ctx ginkgo.SpecContext) map[string]opstats.CacheSummary
	)

	ginkgo.BeforeEach(func() {
		cacheDir := ginkgo.GinkgoT().TempDir()
		gomega.Expect(werf.Init("", ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		cache = NewManifestCache(cacheDir)
		collector = opstats.NewCollector()
		rows = func(ctx ginkgo.SpecContext) map[string]opstats.CacheSummary {
			res := make(map[string]opstats.CacheSummary)
			for _, summary := range collector.CacheSummary(ctx) {
				res[string(summary.Operation)+"/"+string(summary.Layer)] = summary
			}
			return res
		}
	})

	ginkgo.It("counts a stored manifest as a disk hit and an absent one as a miss", func(ctx ginkgo.SpecContext) {
		countedCtx := opstats.NewContext(logging.WithLogger(ctx), collector)

		info, err := cache.GetImageInfo(countedCtx, storageName, "absent")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(info).To(gomega.BeNil())

		gomega.Expect(cache.StoreImageInfo(countedCtx, storageName, &Info{Name: "stored", Tag: "tag"})).To(gomega.Succeed())
		info, err = cache.GetImageInfo(countedCtx, storageName, "stored")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(info).NotTo(gomega.BeNil())

		// The cache is kept on disk only, so there is no memory row to report.
		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"registry: image get/disk": {Operation: opstats.Operation("registry: image get"), Layer: opstats.CacheLayerDisk, Hit: 1, Miss: 1},
		}))
	})

	ginkgo.It("counts a reset invalid record as a miss without an error", func(ctx ginkgo.SpecContext) {
		countedCtx := opstats.NewContext(logging.WithLogger(ctx), collector)
		recordPath := cache.constructFilePathForImage(storageName, "corrupt")
		gomega.Expect(os.MkdirAll(filepath.Dir(recordPath), 0o755)).To(gomega.Succeed())
		gomega.Expect(os.WriteFile(recordPath, []byte("{not json"), 0o644)).To(gomega.Succeed())

		info, err := cache.GetImageInfo(countedCtx, storageName, "corrupt")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(info).To(gomega.BeNil())

		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"registry: image get/disk": {Operation: opstats.Operation("registry: image get"), Layer: opstats.CacheLayerDisk, Miss: 1},
		}))
	})

	ginkgo.It("counts an unreadable record as a miss and propagates the error", func(ctx ginkgo.SpecContext) {
		countedCtx := opstats.NewContext(logging.WithLogger(ctx), collector)
		recordPath := cache.constructFilePathForImage(storageName, "unreadable")
		gomega.Expect(os.MkdirAll(recordPath, 0o755)).To(gomega.Succeed())

		_, err := cache.GetImageInfo(countedCtx, storageName, "unreadable")
		gomega.Expect(err).To(gomega.HaveOccurred())

		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"registry: image get/disk": {Operation: opstats.Operation("registry: image get"), Layer: opstats.CacheLayerDisk, Miss: 1},
		}))
	})

	ginkgo.It("records nothing when no collector is bound", func(ctx ginkgo.SpecContext) {
		_, err := cache.GetImageInfo(logging.WithLogger(ctx), storageName, "absent")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(collector.CacheSummary(ctx)).To(gomega.BeEmpty())
	})
})
