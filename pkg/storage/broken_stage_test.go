package storage

import (
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/opstats"
)

var _ = ginkgo.Describe("Broken stage counting", func() {
	brokenErr := fmt.Errorf("%s: repo/image", transport.BlobUnknownErrorCode)

	ginkgo.DescribeTable("counts a stage read that the registry answered with a broken image",
		func(ctx ginkgo.SpecContext, registryErr error, expectedErr types.GomegaMatcher, expectedCount int) {
			collector := opstats.NewCollector()
			stages := newRepoStorage(&brokenStageRegistry{markerRegistry: newMarkerRegistry(), err: registryErr}, stagesRepo)

			_, err := stages.GetStageDesc(opstats.NewContext(ctx, collector), proj, *image.NewStageID("digest", 100))

			gomega.Expect(err).To(expectedErr)
			gomega.Expect(brokenCount(collector)).To(gomega.Equal(expectedCount))
		},
		ginkgo.Entry("broken image", fmt.Errorf("%s: repo/image", transport.BlobUnknownErrorCode), gomega.MatchError(ErrBrokenImage), 1),
		ginkgo.Entry("missing image", fmt.Errorf("%s: repo/image", transport.ManifestUnknownErrorCode), gomega.MatchError(ErrStageNotFound), 0),
		ginkgo.Entry("authorization failure", errors.New("UNAUTHORIZED: authentication required"), gomega.MatchError(gomega.ContainSubstring("authentication required")), 0),
		ginkgo.Entry("rejected stage lookup timeout", errors.New("context deadline exceeded"), gomega.MatchError(gomega.ContainSubstring("context deadline exceeded")), 0),
	)

	ginkgo.It("counts a stage fetch that pulled an image with an unknown blob", func(ctx ginkgo.SpecContext) {
		collector := opstats.NewCollector()
		stages := newRepoStorage(newMarkerRegistry(), stagesRepo)
		stages.ContainerBackend = &brokenStageBackend{err: errors.New("pull failed: unknown blob")}

		err := stages.FetchImage(opstats.NewContext(ctx, collector), container_backend.NewLegacyStageImage(nil, "repo:image", stages.ContainerBackend, ""))

		gomega.Expect(err).To(gomega.MatchError(ErrBrokenImage))
		gomega.Expect(brokenCount(collector)).To(gomega.Equal(1))
	})

	ginkgo.It("counts no broken stage when the fetch failed for another reason", func(ctx ginkgo.SpecContext) {
		collector := opstats.NewCollector()
		stages := newRepoStorage(newMarkerRegistry(), stagesRepo)
		stages.ContainerBackend = &brokenStageBackend{err: errors.New("pull failed: connection refused")}

		err := stages.FetchImage(opstats.NewContext(ctx, collector), container_backend.NewLegacyStageImage(nil, "repo:image", stages.ContainerBackend, ""))

		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("connection refused")))
		gomega.Expect(brokenCount(collector)).To(gomega.BeZero())
	})

	ginkgo.It("counts a stage mutation that the registry rejected as broken", func(ctx ginkgo.SpecContext) {
		collector := opstats.NewCollector()
		stages := newRepoStorage(&brokenStageRegistry{markerRegistry: newMarkerRegistry(), err: brokenErr}, stagesRepo)

		err := stages.MutateAndPushImage(opstats.NewContext(ctx, collector), "repo:src", "repo:dest", image.SpecConfig{},
			container_backend.NewLegacyStageImage(nil, "repo:dest", nil, ""))

		gomega.Expect(err).To(gomega.MatchError(ErrBrokenImage))
		gomega.Expect(brokenCount(collector)).To(gomega.Equal(1))
	})

	ginkgo.It("counts every independent detection of the same broken stage", func(ctx ginkgo.SpecContext) {
		collector := opstats.NewCollector()
		ctx2 := opstats.NewContext(ctx, collector)
		stages := newRepoStorage(&brokenStageRegistry{markerRegistry: newMarkerRegistry(), err: brokenErr}, stagesRepo)

		for range 2 {
			_, err := stages.GetStageDesc(ctx2, proj, *image.NewStageID("digest", 100))
			gomega.Expect(err).To(gomega.MatchError(ErrBrokenImage))
		}

		gomega.Expect(brokenCount(collector)).To(gomega.Equal(2))
	})
})
