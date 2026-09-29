package build

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/image"
	"github.com/werf/werf/v3/pkg/build/stage"
	imagePkg "github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Introspecting a stage of an image reused by its content anchor", func() {
	var record *introspectPipelineRecord

	ginkgo.BeforeEach(func() {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
	})

	newPhase := func(targets ...IntrospectTarget) (*BuildPhase, *image.Image) {
		record = &introspectPipelineRecord{}
		storageManager := &anchorLookupStorageManager{
			primaryStagesStorage:   &anchorPrimaryStagesStorage{},
			secondaryStagesStorage: &fakeStagesStorage{},
			inPrimary:              imagePkg.NewStageDescSet(newIntrospectStageDesc("repo:cached")),
			inSecondary:            imagePkg.NewStageDescSet(),
		}
		phase := newTestBuildPhase(storageManager, nil, targets...)
		phase.StagesIterator = NewStagesIterator(phase.Conveyor)
		phase.Conveyor.SetStageImage(&stage.StageImage{Image: &introspectImageStub{name: "repo:cached", record: record}})

		img := newTestImage("app", true)
		img.Conveyor = phase.Conveyor
		img.ForceTargetPlatformLogging = true
		img.SetAnchorDigest("app-anchor")
		img.SetStages([]stage.Interface{
			newIntrospectStageStub(stage.Install, "app", false, record),
			newIntrospectStageStub(stage.ImageSpec, "app", true, record),
		})
		phase.Conveyor.imagesTree.SetImagesGraphForTests(newTestImagesGraph(img))

		return phase, img
	}

	ginkgo.DescribeTable("processes the stages instead of reusing the anchor",
		func(ctx ginkgo.SpecContext, target IntrospectTarget) {
			phase, img := newPhase(target)

			gomega.Expect(phase.resolveAvailableContentAnchors(ctx)).To(gomega.Succeed())
			gomega.Expect(img.AnchorReused).To(gomega.BeFalse())

			gomega.Expect(phase.Conveyor.doImage(ctx, img, []Phase{phase})).To(gomega.Succeed())
			gomega.Expect(record.processedStages).To(gomega.Equal([]string{"install", "imageSpec"}))
			gomega.Expect(record.introspected).To(gomega.Equal([]string{"repo:cached"}))
			gomega.Expect(img.GetContentTagDesc()).NotTo(gomega.BeNil())
			gomega.Expect(img.GetLastNonEmptyStage()).NotTo(gomega.BeNil())
		},
		ginkgo.Entry("requested by image name", IntrospectTarget{ImageName: "app", StageName: "install"}),
		ginkgo.Entry("requested for every image", IntrospectTarget{ImageName: "*", StageName: "install"}),
	)

	ginkgo.DescribeTable("keeps reusing the anchor",
		func(ctx ginkgo.SpecContext, targets []IntrospectTarget) {
			phase, img := newPhase(targets...)

			gomega.Expect(phase.resolveAvailableContentAnchors(ctx)).To(gomega.Succeed())
			gomega.Expect(img.AnchorReused).To(gomega.BeTrue())

			gomega.Expect(phase.Conveyor.doImage(ctx, img, []Phase{phase})).To(gomega.Succeed())
			gomega.Expect(record.processedStages).To(gomega.BeEmpty())
			gomega.Expect(record.introspected).To(gomega.BeEmpty())
			gomega.Expect(img.GetContentTagDesc()).NotTo(gomega.BeNil())
		},
		ginkgo.Entry("nothing is introspected", []IntrospectTarget(nil)),
		ginkgo.Entry("another image is introspected", []IntrospectTarget{{ImageName: "other", StageName: "install"}}),
		ginkgo.Entry("a stage the image does not have is introspected", []IntrospectTarget{{ImageName: "*", StageName: "beforeInstall"}}),
	)
})
