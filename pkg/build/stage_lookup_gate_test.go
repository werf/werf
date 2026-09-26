package build

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/stage"
	imagePkg "github.com/werf/werf/v3/pkg/image"
)

var _ = ginkgo.Describe("Stage lookup strictness gate", func() {
	newPhaseWithImage := func(shouldBeBuiltMode bool) (*BuildPhase, *anchorLookupStorageManager) {
		published := &imagePkg.StageDesc{
			StageID: imagePkg.NewStageID("digest0", 1),
			Info: &imagePkg.Info{
				Name:   "repo:published",
				Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "content"},
			},
		}
		storageManager := &anchorLookupStorageManager{
			primaryStagesStorage:   &anchorPrimaryStagesStorage{},
			secondaryStagesStorage: &fakeStagesStorage{},
			inPrimary:              imagePkg.NewStageDescSet(published),
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
		return phase, storageManager
	}

	ginkgo.It("uses the recent-cache primary lookup in normal build mode", func() {
		_, storageManager := newPhaseWithImage(false)
		gomega.Expect(storageManager.recentPrimaryLookups).To(gomega.BeNumerically(">", 0))
		gomega.Expect(storageManager.strictPrimaryLookups).To(gomega.BeZero())
	})

	ginkgo.It("uses the strict primary lookup in should-be-built mode", func() {
		_, storageManager := newPhaseWithImage(true)
		gomega.Expect(storageManager.strictPrimaryLookups).To(gomega.BeNumerically(">", 0))
		gomega.Expect(storageManager.recentPrimaryLookups).To(gomega.BeZero())
	})
})
