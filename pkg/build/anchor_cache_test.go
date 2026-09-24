package build

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	imagePkg "github.com/werf/werf/v2/pkg/image"
)

var _ = ginkgo.Describe("Content anchor cache prepass", func() {
	ginkgo.DescribeTable("does not refresh misses before parallel image processing",
		func(ctx ginkgo.SpecContext, imageCount int) {
			storageManager := &anchorLookupStorageManager{
				primaryStagesStorage:   &anchorPrimaryStagesStorage{},
				secondaryStagesStorage: &fakeStagesStorage{},
				inPrimary:              imagePkg.NewStageDescSet(),
				inSecondary:            imagePkg.NewStageDescSet(),
			}
			phase := newTestBuildPhase(storageManager, nil)
			var images []*image.Image
			for i := range imageCount {
				img := newTestImage(fmt.Sprintf("image%d", i), true)
				img.Conveyor = phase.Conveyor
				img.ForceTargetPlatformLogging = true
				img.SetAnchorDigest(fmt.Sprintf("anchor%d", i))
				anchor := stage.NewBaseStage(stage.ImageSpec, &stage.BaseStageOptions{ImageName: img.Name})
				anchor.SetContentAnchor(true)
				img.SetStages([]stage.Interface{anchor})
				images = append(images, img)
			}
			phase.Conveyor.imagesTree.SetImagesGraphForTests(newTestImagesGraph(images...))

			gomega.Expect(phase.resolveAvailableContentAnchors(ctx)).To(gomega.Succeed())
			gomega.Expect(storageManager.primaryLookups).To(gomega.BeZero())
			gomega.Expect(storageManager.secondaryLookups).To(gomega.BeZero())
			gomega.Expect(storageManager.cachedPrimaryLookups).To(gomega.Equal(imageCount))
			gomega.Expect(storageManager.cachedSecondaryLookups).To(gomega.Equal(imageCount))

			phase.StagesIterator = NewStagesIterator(phase.Conveyor)
			gomega.Expect(phase.resolveContentAnchor(ctx, images[0], true)).To(gomega.Succeed())
			gomega.Expect(storageManager.primaryLookups).To(gomega.Equal(1))
			gomega.Expect(storageManager.secondaryLookups).To(gomega.Equal(1))
			gomega.Expect(images[0].GetContentTagDesc()).To(gomega.BeNil())

			published := &imagePkg.StageDesc{
				StageID: imagePkg.NewStageID("anchor0", 1),
				Info: &imagePkg.Info{
					Name:   "repo:published-after-prepass",
					Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "content"},
				},
			}
			storageManager.inPrimary = imagePkg.NewStageDescSet(published)
			_, err := phase.BeforeImageStages(ctx, images[0])
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(storageManager.primaryLookups).To(gomega.Equal(2))
			gomega.Expect(images[0].GetContentTagDesc()).To(gomega.Equal(published))
			gomega.Expect(images[0].AnchorReused).To(gomega.BeTrue())
		},
		ginkgo.Entry("one image", 1),
		ginkgo.Entry("six images", 6),
	)
})
