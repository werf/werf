package build

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	buildImage "github.com/werf/werf/v3/pkg/build/image"
	"github.com/werf/werf/v3/pkg/build/stage"
	imagePkg "github.com/werf/werf/v3/pkg/image"
)

var _ = ginkgo.Describe("Exporter", func() {
	var (
		img            *buildImage.Image
		anchor         *stage.BaseStage
		primaryDesc    *imagePkg.StageDesc
		finalDesc      *imagePkg.StageDesc
		storageManager *exportStorageManager
		phase          *BuildPhase
	)

	ginkgo.BeforeEach(func() {
		img = newImage("app", buildImage.NoBaseImage, buildImage.ImageOptions{IsFinal: true})
		primaryDesc = &imagePkg.StageDesc{
			StageID: imagePkg.NewStageID("anchor-digest", 42),
			Info:    &imagePkg.Info{Name: "repo:anchor-digest-42", Tag: "anchor-digest-42"},
		}
		finalDesc = primaryDesc.GetCopy()
		finalDesc.Info.Name = "final-repo:anchor-digest-42"
		anchor = stage.NewBaseStage(stage.ImageSpec, &stage.BaseStageOptions{ImageName: img.Name})
		anchor.SetContentAnchor(true)
		anchor.SetStageImage(&stage.StageImage{Image: &exportLegacyImageStub{stageDesc: primaryDesc}})
		img.SetStages([]stage.Interface{anchor})
		img.SetContentTagDesc(primaryDesc)
		img.AnchorReused = true
		img.ForceTargetPlatformLogging = true
		gomega.Expect(img.GetLastNonEmptyStage()).To(gomega.BeNil())

		storageManager = &exportStorageManager{
			stagesStorage:      &exportStagesStorageStub{exported: map[string]*imagePkg.StageDesc{}},
			finalStageDesc:     finalDesc,
			finalStagesStorage: &exportStagesStorageStub{},
		}
		phase = newTestBuildPhase(storageManager, []string{"app"})
		phase.Conveyor.imagesTree.SetImagesGraphForTests(newTestImagesGraph(img))
	})

	ginkgo.DescribeTable("exports the primary anchor without a last non-empty stage",
		func(ctx ginkgo.SpecContext, publishedToFinalRepo bool) {
			if publishedToFinalRepo {
				img.SetContentTagDesc(finalDesc)
			}
			exporter := NewExporter(phase.Conveyor, ExportOptions{
				ExportImageNameList: []string{"app"},
				ExportTagFuncList: []imagePkg.ExportTagFunc{
					func(name, stageID string) string { return fmt.Sprintf("registry/%s:%s", name, stageID) },
				},
			})

			gomega.Expect(exporter.Run(ctx)).To(gomega.Succeed())
			gomega.Expect(storageManager.stagesStorage.(*exportStagesStorageStub).exported).
				To(gomega.HaveKeyWithValue("registry/app:anchor-digest-42", primaryDesc))
		},
		ginkgo.Entry("without a final repo", false),
		ginkgo.Entry("with a final repo", true),
	)

	ginkgo.It("passes the initialized anchor when copying into the final repo", func(ctx ginkgo.SpecContext) {
		gomega.Expect(phase.publishFinalImage(ctx, img.Name, img, storageManager.finalStagesStorage)).To(gomega.Succeed())
		gomega.Expect(storageManager.copyOptions.FetchStage).To(gomega.BeIdenticalTo(anchor))
		gomega.Expect(img.GetContentTagDesc()).To(gomega.BeIdenticalTo(finalDesc))
		gomega.Expect(anchor.GetStageImage().Image.GetStageDesc()).To(gomega.BeIdenticalTo(primaryDesc))
	})
})
