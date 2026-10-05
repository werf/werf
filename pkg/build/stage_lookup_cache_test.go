package build

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	imagePkg "github.com/werf/werf/v2/pkg/image"
)

var _ = ginkgo.Describe("Build stage snapshot lookup", func() {
	ginkgo.DescribeTable("preserves strict checks while caching normal lookups", func(ctx ginkgo.SpecContext, checkMode bool) {
		phase, img, stg, storageManager := newCachedLookupPhase(ctx)
		phase.ShouldBeBuiltMode = checkMode
		found, unlock, err := phase.calculateStage(ctx, img, stg)
		if unlock != nil {
			defer unlock()
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(found).To(gomega.BeFalse())
		if checkMode {
			gomega.Expect(storageManager.strictLookups).To(gomega.Equal(1))
			gomega.Expect(storageManager.cachedLookups).To(gomega.BeZero())
		} else {
			gomega.Expect(storageManager.cachedLookups).To(gomega.Equal(1))
			gomega.Expect(storageManager.strictLookups).To(gomega.BeZero())
		}
	},
		ginkgo.Entry("normal build", false),
		ginkgo.Entry("check mode", true),
	)

	ginkgo.It("preserves strict secondary lookup in check mode", func(ctx ginkgo.SpecContext) {
		phase, img, stg, storageManager := newCachedLookupPhase(ctx)
		phase.ShouldBeBuiltMode = true
		found, err := phase.findAndFetchStageFromSecondaryStagesStorage(ctx, img, stg)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(found).To(gomega.BeFalse())
		gomega.Expect(storageManager.strictLookups).To(gomega.Equal(1))
		gomega.Expect(storageManager.cachedLookups).To(gomega.BeZero())
	})

	ginkgo.DescribeTable("checks primary storage under the publication lock before copying", func(ctx ginkgo.SpecContext, concurrentWinner bool) {
		phase, img, stg, storageManager := newCachedLookupPhase(ctx)
		stg.SetDigest("digest")
		secondary := cachedLookupDesc("secondary", "secondary-content")
		storageManager.secondary = secondary
		expected := secondary
		if concurrentWinner {
			expected = cachedLookupDesc("winner", "winner-content")
			storageManager.publishOnLock = expected
		}

		found, err := phase.findAndFetchStageFromSecondaryStagesStorage(ctx, img, stg)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(found).To(gomega.BeTrue())
		gomega.Expect(stg.GetStageImage().Image.GetStageDesc()).To(gomega.Equal(expected))
		gomega.Expect(stg.GetContentDigest()).To(gomega.Equal(expected.Info.Labels[imagePkg.WerfStageContentDigestLabel]))
		gomega.Expect(storageManager.freshLookups).To(gomega.Equal(1))
		gomega.Expect(storageManager.cachedLookups).To(gomega.Equal(1))
		gomega.Expect(storageManager.locked).To(gomega.BeFalse())
		if concurrentWinner {
			gomega.Expect(storageManager.copies).To(gomega.BeZero())
		} else {
			gomega.Expect(storageManager.copies).To(gomega.Equal(1))
		}
	},
		ginkgo.Entry("adopts a concurrent primary publisher", true),
		ginkgo.Entry("copies when primary remains empty", false),
	)
})
