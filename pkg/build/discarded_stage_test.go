package build

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/container_backend/stage_builder"
	imagePkg "github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/opstats"
)

type buildableStage struct{ *publicationStage }

var _ stage.Interface = (*buildableStage)(nil)

func (s *buildableStage) IsBuildable() bool { return true }

type stageBuilderStub struct {
	stage_builder.StageBuilderInterface
	builds int
}

func (b *stageBuilderStub) Build(_ context.Context, _ container_backend.BuildOptions) error {
	b.builds++
	return nil
}

var _ = ginkgo.Describe("Discarded stage counting", func() {
	ginkgo.BeforeEach(func() { ginkgo.GinkgoT().Setenv("WERF_DISABLE_PUBLISH_TAG_CACHE_SYNC", "") })

	ginkgo.DescribeTable("counts only a locally built image thrown away for an already published one",
		func(ctx ginkgo.SpecContext, buildable, published bool, expected map[opstats.Event]int) {
			srv, _ := newPublicationLockServer()
			primary := &publicationStorage{}
			if published {
				primary.desc = &imagePkg.StageDesc{
					StageID: imagePkg.NewStageID("shared-digest", 100),
					Info:    &imagePkg.Info{Name: "repo:shared-digest-100", Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "winner-content"}},
				}
			}
			phase, img, stg := newPublicationPhase(ctx, &publicationStorageManager{primary: primary}, srv.URL, false, 10)

			builder := &stageBuilderStub{}
			if buildable {
				stg = &buildableStage{publicationStage: stg.(*publicationStage)}
				stg.GetStageImage().Builder = builder
			}

			collector := opstats.NewCollector()
			gomega.Expect(phase.atomicBuildStageImage(opstats.NewContext(ctx, collector), img, stg)).To(gomega.Succeed())

			gomega.Expect(eventCounts(collector)).To(gomega.Equal(expected))
			if buildable {
				gomega.Expect(builder.builds).To(gomega.Equal(1))
			}
		},
		ginkgo.Entry("buildable stage that lost the publication race", true, true,
			map[opstats.Event]int{opstats.EventStageDiscarded: 1, opstats.EventStageCacheHitRepo: 1}),
		ginkgo.Entry("mutable stage that has not been mutated yet", false, true,
			map[opstats.Event]int{opstats.EventStageCacheHitRepo: 1}),
		ginkgo.Entry("buildable stage that published its own image", true, false,
			map[opstats.Event]int{opstats.EventStageBuilt: 1}),
	)
})

func eventCounts(collector *opstats.Collector) map[opstats.Event]int {
	counts := make(map[opstats.Event]int)
	for _, e := range collector.EventSummary() {
		counts[e.Event] = e.Count
	}
	return counts
}
