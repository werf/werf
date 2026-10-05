package build

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/image"
	"github.com/werf/werf/v3/pkg/build/stage"
	imagePkg "github.com/werf/werf/v3/pkg/image"
)

var _ = ginkgo.Describe("Stage lookup strictness gate", func() {
	newPhaseWithImage := func(shouldBeBuiltMode, publishedInPrimary bool) (*image.Image, *anchorLookupStorageManager) {
		published := &imagePkg.StageDesc{
			StageID: imagePkg.NewStageID("digest0", 1),
			Info: &imagePkg.Info{
				Name:   "repo:published",
				Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "content"},
			},
		}
		inPrimary := imagePkg.NewStageDescSet()
		if publishedInPrimary {
			inPrimary = imagePkg.NewStageDescSet(published)
		}
		storageManager := &anchorLookupStorageManager{
			primaryStagesStorage:   &anchorPrimaryStagesStorage{},
			secondaryStagesStorage: &fakeStagesStorage{},
			inPrimary:              inPrimary,
			inSecondary:            imagePkg.NewStageDescSet(),
		}
		phase := newTestBuildPhase(storageManager, nil)
		phase.ShouldBeBuiltMode = shouldBeBuiltMode

		img := newTestImage("image0", true)
		img.Conveyor = phase.Conveyor
		img.ForceTargetPlatformLogging = true
		img.SetAnchorDigest("anchor0")
		anchor := stage.NewBaseStage(stage.ImageSpec, &stage.BaseStageOptions{ImageName: img.Name})
		anchor.SetContentAnchor(true)
		img.SetStages([]stage.Interface{anchor})
		phase.Conveyor.imagesTree.SetImagesGraphForTests(newTestImagesGraph(img))
		phase.StagesIterator = NewStagesIterator(phase.Conveyor)

		_, err := phase.BeforeImageStages(ginkgo.GinkgoT().Context(), img)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return img, storageManager
	}

	ginkgo.It("reuses the whole-build primary snapshot in normal build mode", func() {
		_, storageManager := newPhaseWithImage(false, true)
		gomega.Expect(storageManager.cachedPrimaryLookups).To(gomega.BeNumerically(">", 0))
		gomega.Expect(storageManager.recentPrimaryLookups).To(gomega.BeZero())
		gomega.Expect(storageManager.strictPrimaryLookups).To(gomega.BeZero())
	})

	ginkgo.It("scans secondary storages on a primary miss in normal build mode", func() {
		img, storageManager := newPhaseWithImage(false, false)
		gomega.Expect(storageManager.cachedSecondaryLookups).To(gomega.BeNumerically(">", 0))
		gomega.Expect(storageManager.secondaryLookups).To(gomega.BeZero())
		gomega.Expect(img.GetContentTagDesc()).To(gomega.BeNil())
	})

	ginkgo.DescribeTable("looks up nothing but the fresh primary storage in should-be-built mode",
		func(publishedInPrimary bool) {
			img, storageManager := newPhaseWithImage(true, publishedInPrimary)
			gomega.Expect(storageManager.strictPrimaryLookups).To(gomega.BeNumerically(">", 0))
			gomega.Expect(storageManager.recentPrimaryLookups).To(gomega.BeZero())
			gomega.Expect(storageManager.cachedPrimaryLookups).To(gomega.BeZero())
			gomega.Expect(storageManager.secondaryLookups).To(gomega.BeZero())
			gomega.Expect(storageManager.cachedSecondaryLookups).To(gomega.BeZero())

			if publishedInPrimary {
				gomega.Expect(img.GetContentTagDesc()).NotTo(gomega.BeNil())
			} else {
				gomega.Expect(img.GetContentTagDesc()).To(gomega.BeNil())
			}
		},
		ginkgo.Entry("the stage is published in the primary storage", true),
		ginkgo.Entry("the stage is missing from the primary storage", false),
	)
})
