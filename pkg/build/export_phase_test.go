package build

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	buildImage "github.com/werf/werf/v3/pkg/build/image"
	"github.com/werf/werf/v3/pkg/build/stage"
	imagePkg "github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/werf"
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
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
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
			stagesStorage:      newExportStagesStorageStub(storage.LocalStorageAddress),
			finalStageDesc:     finalDesc,
			finalStagesStorage: newExportStagesStorageStub("registry.example.com/project/final"),
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

var _ = ginkgo.Describe("Exporter.Run for a multiplatform image", func() {
	ginkgo.BeforeEach(func() {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("picks the storage that actually holds the multiplatform image",
		func(ctx ginkgo.SpecContext, withFinalRepo, publishedToFinalRepo bool) {
			primaryStagesStorage := newExportStagesStorageStub("registry.example.com/project")
			finalStagesStorage := newExportStagesStorageStub("registry.example.com/project/final")

			storageManager := &exportStorageManager{stagesStorage: primaryStagesStorage}
			if withFinalRepo {
				storageManager.finalStagesStorage = finalStagesStorage
			}

			primaryDesc := &imagePkg.StageDesc{
				StageID: imagePkg.NewStageID("index-digest", 0),
				Info:    &imagePkg.Info{Name: "registry.example.com/project:index-digest", IsIndex: true},
			}
			finalDesc := primaryDesc.GetCopy()
			finalDesc.Info.Name = "registry.example.com/project/final:index-digest"

			images := []*buildImage.Image{
				newTestImageForPlatform("linux/amd64", "app", true),
				newTestImageForPlatform("linux/arm64", "app", true),
			}
			for _, img := range images {
				img.SetContentTagDesc(&imagePkg.StageDesc{
					StageID: imagePkg.NewStageID(img.TargetPlatform, 42),
					Info:    &imagePkg.Info{Name: "registry.example.com/project:" + img.TargetPlatform},
				})
			}

			multiplatformImg := buildImage.NewMultiplatformImage("app", images, 0, 1)
			multiplatformImg.SetStageDesc(primaryDesc)
			if publishedToFinalRepo {
				multiplatformImg.SetFinalStageDesc(finalDesc)
			}

			phase := newTestBuildPhase(storageManager, []string{"app"})
			phase.Conveyor.imagesTree.SetImagesGraphForTests(newTestImagesGraph(images...))
			phase.Conveyor.imagesTree.SetMultiplatformImage(multiplatformImg)

			exporter := NewExporter(phase.Conveyor, ExportOptions{
				ExportImageNameList: []string{"app"},
				ExportTagFuncList: []imagePkg.ExportTagFunc{
					func(name, stageID string) string { return fmt.Sprintf("registry/%s:%s", name, stageID) },
				},
			})

			gomega.Expect(exporter.Run(ctx)).To(gomega.Succeed())

			exportedFrom, notExportedFrom, expectedDesc := finalStagesStorage, primaryStagesStorage, finalDesc
			if !publishedToFinalRepo {
				exportedFrom, notExportedFrom, expectedDesc = primaryStagesStorage, finalStagesStorage, primaryDesc
			}

			gomega.Expect(notExportedFrom.exported).To(gomega.BeEmpty())
			gomega.Expect(exportedFrom.exported).To(gomega.HaveKeyWithValue(
				fmt.Sprintf("registry/app:%s", multiplatformImg.GetStageID().String()), expectedDesc))
		},
		ginkgo.Entry("image published to the final repo", true, true),
		ginkgo.Entry("final repo configured, image not published there", true, false),
		ginkgo.Entry("no final repo configured", false, false),
	)
})

var _ = ginkgo.Describe("Exporter.RunFromReport", func() {
	const (
		primaryRepo = "registry.example.com/project"
		finalRepo   = "registry.example.com/project/final"
		stageTag    = "9d3a6a4a1c8f4b2e8c1d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b-1700000000000"
	)

	ginkgo.BeforeEach(func() {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("validates local reports without a repository digest",
		func(ctx ginkgo.SpecContext, imageID string, valid bool) {
			record := newTestReportImageRecord("frontend", true)
			record.DockerImageID = imageID
			record.DockerImageDigest = ""

			report, err := LoadBuildReportFromFile(ctx, writeBuildReport(record))
			if !valid {
				gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(`image "frontend" has empty DockerImageID`)))
				return
			}
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(report.Images["frontend"].DockerImageID).To(gomega.Equal(imageID))
		},
		ginkgo.Entry("requires an image ID", "", false),
		ginkgo.Entry("does not require a repository digest", "sha256:5f1993108ca9000000000000000000000000000000000000000000000000000000", true),
	)

	ginkgo.DescribeTable("picks the storage that actually holds the reported image",
		func(ctx ginkgo.SpecContext, primaryAddress, finalAddress, recordRepo string, expectExportedFromFinalRepo bool) {
			primaryStagesStorage := newExportStagesStorageStub(primaryAddress)
			finalStagesStorage := newExportStagesStorageStub(finalAddress)

			storageManager := &exportStorageManager{stagesStorage: primaryStagesStorage}
			if finalAddress != "" {
				storageManager.finalStagesStorage = finalStagesStorage
			}

			reportPath := writeBuildReport(ReportImageRecord{
				WerfImageName:     "app",
				DockerRepo:        recordRepo,
				DockerTag:         stageTag,
				DockerImageID:     "sha256:imageid",
				DockerImageDigest: "sha256:digest",
				DockerImageName:   fmt.Sprintf("%s:%s", recordRepo, stageTag),
				TargetPlatform:    "linux/amd64",
				Final:             true,
			})

			exporter := NewExporter(newTestBuildPhase(storageManager, []string{"app"}).Conveyor, ExportOptions{
				ExportImageNameList: []string{"app"},
				ExportTagFuncList: []imagePkg.ExportTagFunc{
					func(name, stageID string) string { return fmt.Sprintf("registry/%s:%s", name, stageID) },
				},
			})

			gomega.Expect(exporter.RunFromReport(ctx, reportPath)).To(gomega.Succeed())

			exportedFrom, notExportedFrom := finalStagesStorage, primaryStagesStorage
			if !expectExportedFromFinalRepo {
				exportedFrom, notExportedFrom = primaryStagesStorage, finalStagesStorage
			}

			gomega.Expect(notExportedFrom.exported).To(gomega.BeEmpty())
			gomega.Expect(exportedFrom.exported).To(gomega.HaveKey(fmt.Sprintf("registry/app:%s", stageTag)))
			gomega.Expect(exportedFrom.exported[fmt.Sprintf("registry/app:%s", stageTag)].Info.Name).
				To(gomega.Equal(fmt.Sprintf("%s:%s", recordRepo, stageTag)))
		},
		ginkgo.Entry("local primary storage with a record published to the final repo",
			storage.LocalStorageAddress, finalRepo, finalRepo, true),
		ginkgo.Entry("remote primary storage with a record published to the final repo",
			primaryRepo, finalRepo, finalRepo, true),
		ginkgo.Entry("final repo address spelled differently than in the record",
			storage.LocalStorageAddress, "docker.io/library/app-final", "index.docker.io/library/app-final", true),
		ginkgo.Entry("remote primary storage with a record left in the primary repo",
			primaryRepo, finalRepo, primaryRepo, false),
		ginkgo.Entry("no final repo configured",
			primaryRepo, "", primaryRepo, false),
	)
})
