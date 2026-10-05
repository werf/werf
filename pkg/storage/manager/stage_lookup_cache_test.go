package manager

import (
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
)

var _ = ginkgo.Describe("Storage manager cached stage discovery", func() {
	ginkgo.DescribeTable("controls refresh after a cached miss", func(ctx ginkgo.SpecContext, mode string, missing bool) {
		digest := strings.Repeat("a", 56)
		backend := &lookupBackend{}
		stagesStorage := storage.NewLocalStagesStorage(backend)
		manager := &StorageManager{ProjectName: "project", StagesStorage: stagesStorage}

		stages, err := manager.GetStageDescSetByDigestFromStagesStorageCached(ctx, "stage", digest, 0, stagesStorage)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stages.IsEmpty()).To(gomega.BeTrue())

		backend.info = &image.Info{Name: "project:" + digest + "-1700000000001"}
		backend.images = image.ImagesList{{RepoTags: []string{backend.info.Name}}}
		switch mode {
		case "cached":
			stages, err = manager.GetStageDescSetByDigestFromStagesStorageCached(ctx, "stage", digest, 0, stagesStorage)
		case "fresh":
			stages, err = manager.GetStageDescSetByDigestFromStagesStorage(ctx, "stage", digest, 0, stagesStorage)
		case "refresh-on-miss":
			stages, err = manager.GetStageDescSetByDigestFromStagesStorageWithCache(ctx, "stage", digest, 0, stagesStorage)
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stages.IsEmpty()).To(gomega.Equal(missing))
	},
		ginkgo.Entry("cached lookup preserves a miss", "cached", true),
		ginkgo.Entry("fresh lookup sees a new stage", "fresh", false),
		ginkgo.Entry("refresh-on-miss lookup sees a new stage", "refresh-on-miss", false),
	)
})
