package build

import (
	"errors"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	buildImage "github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	imagePkg "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/opstats"
)

var _ = ginkgo.Describe("Stage reuse after a failed local build", func() {
	ginkgo.BeforeEach(func() { ginkgo.GinkgoT().Setenv("WERF_DISABLE_PUBLISH_TAG_CACHE_SYNC", "") })

	winnerDesc := &imagePkg.StageDesc{
		StageID: imagePkg.NewStageID("shared-digest", 100),
		Info:    &imagePkg.Info{Name: "repo:shared-digest-100", Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "winner-content"}},
	}

	newFailingBuild := func(ctx ginkgo.SpecContext, published bool) (*BuildPhase, *publicationStorage, <-chan struct{}, *buildImage.Image, stage.Interface) {
		srv, attempts := newPublicationLockServer()
		primary := &publicationStorage{}
		if published {
			primary.desc = winnerDesc
		}
		phase, img, stg := newPublicationPhase(ctx, &publicationStorageManager{primary: primary}, srv.URL)
		buildable := &buildableStage{publicationStage: stg.(*publicationStage)}
		buildable.GetStageImage().Builder = &stageBuilderStub{buildErr: errors.New("builder exploded")}

		return phase, primary, attempts, img, buildable
	}

	ginkgo.It("reuses a stage published by another process", func(ctx ginkgo.SpecContext) {
		phase, primary, attempts, img, stg := newFailingBuild(ctx, true)
		collector := opstats.NewCollector()

		reused, err := phase.atomicBuildStageImage(opstats.NewContext(ctx, collector), img, stg)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(reused).To(gomega.BeTrue())

		gomega.Expect(stg.GetStageImage().Image.GetStageDesc()).To(gomega.Equal(winnerDesc))
		gomega.Expect(stg.GetContentDigest()).To(gomega.Equal("winner-content"))
		gomega.Expect(eventCounts(collector)).To(gomega.Equal(map[opstats.Event]int{opstats.EventStageCacheHitRepo: 1}), "a failed build must be counted as a reused stage, not as a built or discarded one")
		gomega.Expect(primary.writes).To(gomega.BeZero())
		gomega.Eventually(attempts, 5*time.Second).Should(gomega.Receive(), "the fresh lookup must happen under the stage lock")
	})

	ginkgo.DescribeTable("reports a reused stage as not rebuilt",
		func(ctx ginkgo.SpecContext, buildErr error, rebuilt bool) {
			srv, _ := newPublicationLockServer()
			primary := &publicationStorage{}
			if buildErr != nil {
				primary.desc = winnerDesc
			}
			phase, img, stg := newPublicationPhase(ctx, &publicationStorageManager{primary: primary}, srv.URL)
			img.IsDockerfileImage = true
			img.DockerfileImageConfig = &config.ImageFromDockerfile{}
			reported := &reportedStage{buildableStage: &buildableStage{publicationStage: stg.(*publicationStage)}, buildErr: buildErr}
			img.SetStages([]stage.Interface{reported})

			gomega.Expect(phase.onImageStage(ctx, img, reported)).To(gomega.Succeed())

			gomega.Expect(getStagesReport(img, false)).To(gomega.HaveLen(1))
			gomega.Expect(getStagesReport(img, false)[0].Rebuilt).To(gomega.Equal(rebuilt))
		},
		ginkgo.Entry("stage published by another process while the local build was failing", errors.New("builder exploded"), false),
		ginkgo.Entry("stage built by this process", nil, true),
	)

	ginkgo.It("returns the build error when nothing has been published", func(ctx ginkgo.SpecContext) {
		phase, primary, _, img, stg := newFailingBuild(ctx, false)
		collector := opstats.NewCollector()

		_, err := phase.atomicBuildStageImage(opstats.NewContext(ctx, collector), img, stg)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("builder exploded")))

		gomega.Expect(eventCounts(collector)).To(gomega.BeEmpty())
		gomega.Expect(primary.writes).To(gomega.BeZero())
	})
})
