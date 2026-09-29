package storage

import (
	"errors"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/image"
)

var _ = ginkgo.Describe("RepoStagesStorage published stage tags", func() {
	ginkgo.It("makes a stored stage visible to cached lookups without another tags listing", func(ctx ginkgo.SpecContext) {
		storage, tagsListRequests := newTagCacheRepoStagesStorage(ctx, &pushStageBackendStub{})
		stageID := image.NewStageID(tagCacheStageDigest, 1700000000)
		stageImageName := storage.ConstructStageImageName("", stageID.Digest, stageID.CreationTs)

		stageIDs, err := storage.GetStagesIDsByDigest(ctx, "", tagCacheStageDigest, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageIDs).To(gomega.BeEmpty())
		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(1)))

		gomega.Expect(storage.StoreImage(ctx, container_backend.NewLegacyStageImage(nil, stageImageName, storage.ContainerBackend, ""))).To(gomega.Succeed())

		gomega.Expect(cachedStageIDs(ctx, storage)).To(gomega.Equal([]string{stageID.String()}))
		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(1)))
	})

	ginkgo.It("does not make a stage visible when the push fails", func(ctx ginkgo.SpecContext) {
		storage, _ := newTagCacheRepoStagesStorage(ctx, &pushStageBackendStub{pushErr: errors.New("push refused")})
		stageImageName := storage.ConstructStageImageName("", tagCacheStageDigest, 1700000000)

		_, err := storage.GetStagesIDsByDigest(ctx, "", tagCacheStageDigest, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		err = storage.StoreImage(ctx, container_backend.NewLegacyStageImage(nil, stageImageName, storage.ContainerBackend, ""))
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("push refused")))

		gomega.Expect(cachedStageIDs(ctx, storage)).To(gomega.BeEmpty())
	})

	ginkgo.It("makes a mutated stage visible to cached lookups without another tags listing", func(ctx ginkgo.SpecContext) {
		storage, tagsListRequests := newTagCacheRepoStagesStorage(ctx, &pushStageBackendStub{})
		stageID := image.NewStageID(tagCacheStageDigest, 1700000000)
		sourceReference := fmt.Sprintf("%s:source", storage.RepoAddress)
		pushRandomImage(sourceReference)

		stageIDs, err := storage.GetStagesIDsByDigest(ctx, "", tagCacheStageDigest, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageIDs).To(gomega.BeEmpty())
		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(1)))

		destinationReference := storage.ConstructStageImageName("", stageID.Digest, stageID.CreationTs)
		stageImage := container_backend.NewLegacyStageImage(nil, destinationReference, storage.ContainerBackend, "")
		gomega.Expect(storage.MutateAndPushImage(ctx, sourceReference, destinationReference, image.SpecConfig{}, stageImage)).To(gomega.Succeed())

		gomega.Expect(cachedStageIDs(ctx, storage)).To(gomega.Equal([]string{stageID.String()}))
		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(1)))
	})

	ginkgo.It("does not make a mutated stage visible when the mutation fails", func(ctx ginkgo.SpecContext) {
		storage, _ := newTagCacheRepoStagesStorage(ctx, &pushStageBackendStub{})
		destinationReference := storage.ConstructStageImageName("", tagCacheStageDigest, 1700000000)

		_, err := storage.GetStagesIDsByDigest(ctx, "", tagCacheStageDigest, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		stageImage := container_backend.NewLegacyStageImage(nil, destinationReference, storage.ContainerBackend, "")
		err = storage.MutateAndPushImage(ctx, fmt.Sprintf("%s:missing", storage.RepoAddress), destinationReference, image.SpecConfig{}, stageImage)
		gomega.Expect(err).To(gomega.HaveOccurred())

		gomega.Expect(cachedStageIDs(ctx, storage)).To(gomega.BeEmpty())
	})
})
