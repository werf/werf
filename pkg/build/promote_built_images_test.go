package build

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/config"
	imagePkg "github.com/werf/werf/v3/pkg/image"
)

var _ = ginkgo.Describe("Promote built images", func() {
	ginkgo.It("reuses a published stage despite a stale empty stages snapshot", func(ctx ginkgo.SpecContext) {
		server, _ := newPublicationLockServer()
		primary := &publicationStorage{desc: &imagePkg.StageDesc{
			StageID: imagePkg.NewStageID("shared-digest", 100),
			Info:    &imagePkg.Info{Name: "repo:shared-digest-100", Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "published-content"}},
		}}
		storageManager := &publicationStorageManager{primary: primary}
		phase, img, stg := newPublicationPhase(ctx, storageManager, server.URL, false, 10)
		phase.Conveyor.StorageManager = &promotionLookupStorageManager{storageManager}
		phase.PromoteBuiltImages = true
		reported := &reportedStage{buildableStage: &buildableStage{publicationStage: stg.(*publicationStage)}}

		gomega.Expect(phase.onImageStage(ctx, img, reported)).To(gomega.Succeed())
		gomega.Expect(reported.GetStageImage().Image.GetStageDesc()).To(gomega.Equal(primary.desc))
		gomega.Expect(primary.writes).To(gomega.BeZero())
		gomega.Expect(reported.GetStageImage().Builder.(*stageBuilderStub).builds).To(gomega.BeZero())
	})

	ginkgo.DescribeTable("does not build a missing stage in promote mode",
		func(ctx ginkgo.SpecContext, promote bool) {
			server, _ := newPublicationLockServer()
			primary := &publicationStorage{}
			phase, img, stg := newPublicationPhase(ctx, &publicationStorageManager{primary: primary}, server.URL, false, 10)
			phase.PromoteBuiltImages = promote
			img.IsDockerfileImage = true
			img.DockerfileImageConfig = &config.ImageFromDockerfile{}
			reported := &reportedStage{buildableStage: &buildableStage{publicationStage: stg.(*publicationStage)}}
			img.SetStages([]stage.Interface{reported})

			err := phase.onImageStage(ctx, img, reported)
			if promote {
				gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("build images first without --promote-built-images")))
				gomega.Expect(primary.writes).To(gomega.BeZero())
				gomega.Expect(reported.GetStageImage().Builder).To(gomega.BeNil())
			} else {
				gomega.Expect(err).To(gomega.Succeed())
				gomega.Expect(primary.writes).To(gomega.Equal(1))
				gomega.Expect(reported.GetStageImage().Builder.(*stageBuilderStub).builds).To(gomega.Equal(1))
			}
		},
		ginkgo.Entry("promote fails before invoking the builder", true),
		ginkgo.Entry("ordinary build invokes the builder and publishes the stage", false),
	)
})
