package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/v3/pkg/container_backend"
	registry_api "github.com/werf/werf/v3/pkg/docker_registry/api"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/opstats"
)

type brokenStageRegistry struct {
	*markerRegistry
	err error
}

func (r *brokenStageRegistry) GetRepoImage(_ context.Context, _ string) (*image.Info, error) {
	return nil, r.err
}

func (r *brokenStageRegistry) MutateAndPushImage(_ context.Context, _, _ string, _ ...registry_api.MutateOption) error {
	return r.err
}

type brokenStageBackend struct {
	container_backend.ContainerBackend
	err error
}

func (b *brokenStageBackend) PullImageFromRegistry(_ context.Context, _ container_backend.LegacyImageInterface) error {
	return b.err
}

var _ = Describe("Broken stage counting", func() {
	brokenErr := fmt.Errorf("%s: repo/image", transport.BlobUnknownErrorCode)

	DescribeTable("counts a stage read that the registry answered with a broken image",
		func(ctx SpecContext, registryErr error, expectedErr types.GomegaMatcher, expectedCount int) {
			collector := opstats.NewCollector()
			stages := newRepoStorage(&brokenStageRegistry{markerRegistry: newMarkerRegistry(), err: registryErr}, stagesRepo)

			_, err := stages.GetStageDesc(opstats.NewContext(ctx, collector), proj, *image.NewStageID("digest", 100))

			Expect(err).To(expectedErr)
			Expect(brokenCount(collector)).To(Equal(expectedCount))
		},
		Entry("broken image", fmt.Errorf("%s: repo/image", transport.BlobUnknownErrorCode), MatchError(ErrBrokenImage), 1),
		Entry("missing image", fmt.Errorf("%s: repo/image", transport.ManifestUnknownErrorCode), MatchError(ErrStageNotFound), 0),
		Entry("authorization failure", errors.New("UNAUTHORIZED: authentication required"), MatchError(ContainSubstring("authentication required")), 0),
		Entry("rejected stage lookup timeout", errors.New("context deadline exceeded"), MatchError(ContainSubstring("context deadline exceeded")), 0),
	)

	It("counts a stage fetch that pulled an image with an unknown blob", func(ctx SpecContext) {
		collector := opstats.NewCollector()
		stages := newRepoStorage(newMarkerRegistry(), stagesRepo)
		stages.ContainerBackend = &brokenStageBackend{err: errors.New("pull failed: unknown blob")}

		err := stages.FetchImage(opstats.NewContext(ctx, collector), container_backend.NewLegacyStageImage(nil, "repo:image", stages.ContainerBackend, ""))

		Expect(err).To(MatchError(ErrBrokenImage))
		Expect(brokenCount(collector)).To(Equal(1))
	})

	It("counts no broken stage when the fetch failed for another reason", func(ctx SpecContext) {
		collector := opstats.NewCollector()
		stages := newRepoStorage(newMarkerRegistry(), stagesRepo)
		stages.ContainerBackend = &brokenStageBackend{err: errors.New("pull failed: connection refused")}

		err := stages.FetchImage(opstats.NewContext(ctx, collector), container_backend.NewLegacyStageImage(nil, "repo:image", stages.ContainerBackend, ""))

		Expect(err).To(MatchError(ContainSubstring("connection refused")))
		Expect(brokenCount(collector)).To(BeZero())
	})

	It("counts a stage mutation that the registry rejected as broken", func(ctx SpecContext) {
		collector := opstats.NewCollector()
		stages := newRepoStorage(&brokenStageRegistry{markerRegistry: newMarkerRegistry(), err: brokenErr}, stagesRepo)

		err := stages.MutateAndPushImage(opstats.NewContext(ctx, collector), "repo:src", "repo:dest", image.SpecConfig{},
			container_backend.NewLegacyStageImage(nil, "repo:dest", nil, ""))

		Expect(err).To(MatchError(ErrBrokenImage))
		Expect(brokenCount(collector)).To(Equal(1))
	})

	It("counts every independent detection of the same broken stage", func(ctx SpecContext) {
		collector := opstats.NewCollector()
		ctx2 := opstats.NewContext(ctx, collector)
		stages := newRepoStorage(&brokenStageRegistry{markerRegistry: newMarkerRegistry(), err: brokenErr}, stagesRepo)

		for range 2 {
			_, err := stages.GetStageDesc(ctx2, proj, *image.NewStageID("digest", 100))
			Expect(err).To(MatchError(ErrBrokenImage))
		}

		Expect(brokenCount(collector)).To(Equal(2))
	})
})

func brokenCount(collector *opstats.Collector) int {
	for _, e := range collector.EventSummary() {
		if e.Event == opstats.EventStageBroken {
			return e.Count
		}
	}
	return 0
}
