package build

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	imagePkg "github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/opstats"
)

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
