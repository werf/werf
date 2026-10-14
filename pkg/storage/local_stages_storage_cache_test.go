package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/image"
)

var (
	cachedDigestA = fmt.Sprintf("%056x", 0xaa)
	cachedDigestB = fmt.Sprintf("%056x", 0xbb)
	cachedTagA    = fmt.Sprintf("%s-%d", cachedDigestA, 1700000000001)
	cachedTagA2   = fmt.Sprintf("%s-%d", cachedDigestA, 1700000000002)
	cachedTagB    = fmt.Sprintf("%s-%d", cachedDigestB, 1700000000003)
	brokenTagA    = cachedDigestA + "-notatimestamp"
	brokenTagB    = cachedDigestB + "-notatimestamp"
)

var _ = ginkgo.Describe("Local stage lookup cache", func() {
	ginkgo.It("reuses one project image list for different missing digests", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{}
		storage := NewLocalStagesStorage(backend)

		for i := range 32 {
			stages, err := storage.GetStagesIDsByDigest(ctx, "project", fmt.Sprintf("%056x", i), 0, WithCache())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(stages).To(gomega.BeEmpty())
		}

		gomega.Expect(backend.calls).To(gomega.Equal(1))
		gomega.Expect(backend.options.Filters).To(gomega.Equal([]util.Pair[string, string]{util.NewPair("reference", "project:*")}))
	})

	ginkgo.DescribeTable("selects stages of the requested digest from the project snapshot",
		func(ctx ginkgo.SpecContext, images image.ImagesList, expectedTags []string) {
			backend := &localImageListBackendStub{images: images}
			stages, err := NewLocalStagesStorage(backend).GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(stageStrings(stages)).To(gomega.ConsistOf(expectedTags))
		},
		ginkgo.Entry("unlabeled stage of the project", image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA}},
		}, []string{cachedTagA}),
		ginkgo.Entry("stage labeled for another project", image.ImagesList{
			{Labels: map[string]string{image.WerfLabel: "other"}, RepoTags: []string{"project:" + cachedTagA}},
		}, []string{cachedTagA}),
		ginkgo.Entry("only matching aliases of the same image", image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA, "project:" + cachedTagB, "other:" + cachedTagA2, "project:alias"}},
		}, []string{cachedTagA}),
		ginkgo.Entry("another digest is excluded", image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA}},
			{RepoTags: []string{"project:" + cachedTagB}},
		}, []string{cachedTagA}),
		ginkgo.Entry("malformed tag of another digest does not poison the lookup", image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA}},
			{RepoTags: []string{"project:" + brokenTagB}},
		}, []string{cachedTagA}),
		ginkgo.Entry("another project with the same digest is excluded", image.ImagesList{
			{RepoTags: []string{"other:" + cachedTagA}},
		}, []string{}),
		ginkgo.Entry("Buildah local reference", image.ImagesList{
			{RepoTags: []string{"localhost/project:" + cachedTagA}},
		}, []string{cachedTagA}),
		ginkgo.Entry("Buildah aliases remain scoped to project and digest", image.ImagesList{
			{RepoTags: []string{"localhost/project:" + cachedTagA, "localhost/project:" + cachedTagB, "localhost/other:" + cachedTagA2}},
		}, []string{cachedTagA}),
		ginkgo.Entry("another Buildah project is excluded", image.ImagesList{
			{RepoTags: []string{"localhost/other:" + cachedTagA}},
		}, []string{}),
		ginkgo.Entry("another registry is excluded", image.ImagesList{
			{RepoTags: []string{"registry.example/project:" + cachedTagA}},
		}, []string{}),
		ginkgo.Entry("a nested namespace is excluded", image.ImagesList{
			{RepoTags: []string{"localhost/other/project:" + cachedTagA}},
		}, []string{}),
		ginkgo.Entry("empty snapshot", image.ImagesList{}, []string{}),
	)

	ginkgo.It("preserves fresh lookup results for short Buildah references", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: image.ImagesList{
			{RepoTags: []string{"localhost/project:" + cachedTagA}},
		}}
		storage := NewLocalStagesStorage(backend)
		fresh, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(fresh)).To(gomega.ConsistOf(cachedTagA))
		cached, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(cached)).To(gomega.Equal(stageStrings(fresh)))
	})

	ginkgo.It("reports conversion errors of matching malformed tags", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: image.ImagesList{{RepoTags: []string{"project:" + brokenTagA}}}}
		_, err := NewLocalStagesStorage(backend).GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("creation timestamp")))
	})

	ginkgo.It("applies the parent timestamp filter per call, not per snapshot", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA}},
			{RepoTags: []string{"project:" + cachedTagA2}},
		}}
		storage := NewLocalStagesStorage(backend)

		all, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(all)).To(gomega.ConsistOf(cachedTagA, cachedTagA2))

		newer, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 1700000000002, WithCache())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(newer)).To(gomega.ConsistOf(cachedTagA2))

		gomega.Expect(backend.calls).To(gomega.Equal(1))
	})

	ginkgo.It("keeps other digest aliases available in the shared snapshot", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA, "project:" + cachedTagB}},
		}}
		storage := NewLocalStagesStorage(backend)
		for _, digestAndTag := range [][2]string{{cachedDigestA, cachedTagA}, {cachedDigestB, cachedTagB}} {
			stages, err := storage.GetStagesIDsByDigest(ctx, "project", digestAndTag[0], 0, WithCache())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(stageStrings(stages)).To(gomega.ConsistOf(digestAndTag[1]))
		}
		gomega.Expect(backend.calls).To(gomega.Equal(1))
	})

	ginkgo.It("keeps a separate snapshot per project", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA}},
			{RepoTags: []string{"other:" + cachedTagA2}},
		}}
		storage := NewLocalStagesStorage(backend)

		first, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(first)).To(gomega.ConsistOf(cachedTagA))

		second, err := storage.GetStagesIDsByDigest(ctx, "other", cachedDigestA, 0, WithCache())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(second)).To(gomega.ConsistOf(cachedTagA2))

		gomega.Expect(backend.calls).To(gomega.Equal(2))
		gomega.Expect(backend.options.Filters).To(gomega.Equal([]util.Pair[string, string]{util.NewPair("reference", "other:*")}))
	})

	ginkgo.It("does not cache failed listings", func(ctx ginkgo.SpecContext) {
		listErr := errors.New("list failed")
		backend := &localImageListBackendStub{err: listErr}
		storage := NewLocalStagesStorage(backend)

		_, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
		gomega.Expect(err).To(gomega.MatchError(listErr))

		backend.err = nil
		backend.images = image.ImagesList{{RepoTags: []string{"project:" + cachedTagA}}}
		stages, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(stages)).To(gomega.ConsistOf(cachedTagA))
		gomega.Expect(backend.calls).To(gomega.Equal(2))
	})

	ginkgo.It("lists the whole project without the cache option too", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: image.ImagesList{{RepoTags: []string{"project:" + cachedTagA}}}}
		storage := NewLocalStagesStorage(backend)

		cached, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(cached)).To(gomega.ConsistOf(cachedTagA))

		backend.images = image.ImagesList{{RepoTags: []string{"project:" + cachedTagA, "project:" + cachedTagA2}}}

		fresh, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(fresh)).To(gomega.ConsistOf(cachedTagA, cachedTagA2))
		gomega.Expect(backend.options.Filters).To(gomega.Equal([]util.Pair[string, string]{util.NewPair("reference", "project:*")}))
		gomega.Expect(backend.calls).To(gomega.Equal(2))
	})
})

var _ = ginkgo.Describe("Local stage lookup cache maintenance", func() {
	warmSnapshot := image.ImagesList{{RepoTags: []string{"project:" + cachedTagA, "project:" + cachedTagB}}}

	warmedStorage := func(ctx ginkgo.SpecContext, backend container_backend.ContainerBackend) *LocalStagesStorage {
		storage := NewLocalStagesStorage(backend)
		stages, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(stages)).To(gomega.ConsistOf(cachedTagA))
		return storage
	}

	cachedStages := func(ctx context.Context, storage *LocalStagesStorage, digest string) []string {
		stages, err := storage.GetStagesIDsByDigest(ctx, "project", digest, 0, WithCache())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return stageStrings(stages)
	}

	ginkgo.DescribeTable("replaces the whole project snapshot with a fresh listing",
		func(ctx ginkgo.SpecContext, freshListing image.ImagesList, expectedA, expectedB []string) {
			backend := &localImageListBackendStub{images: warmSnapshot}
			storage := warmedStorage(ctx, backend)

			backend.images = freshListing
			_, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.ConsistOf(expectedA))
			gomega.Expect(cachedStages(ctx, storage, cachedDigestB)).To(gomega.ConsistOf(expectedB))
			gomega.Expect(backend.calls).To(gomega.Equal(2))
		},
		ginkgo.Entry("a new tag of the requested digest becomes visible", image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA}},
			{RepoTags: []string{"project:" + cachedTagA2}},
			{RepoTags: []string{"project:" + cachedTagB}},
		}, []string{cachedTagA, cachedTagA2}, []string{cachedTagB}),
		ginkgo.Entry("a replaced tag of the requested digest drops the old one", image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA2, "project:" + cachedTagB}},
		}, []string{cachedTagA2}, []string{cachedTagB}),
		ginkgo.Entry("a tag of another digest missing from the listing is dropped too", image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA2}},
		}, []string{cachedTagA2}, []string{}),
		ginkgo.Entry("an empty listing clears the whole project", image.ImagesList{},
			[]string{}, []string{}),
		ginkgo.Entry("a Buildah listing refreshes the whole project", image.ImagesList{
			{RepoTags: []string{"localhost/project:" + cachedTagA2, "localhost/project:" + cachedTagB}},
		}, []string{cachedTagA2}, []string{cachedTagB}),
	)

	ginkgo.It("keeps the cached tags when a fresh listing fails", func(ctx ginkgo.SpecContext) {
		listErr := errors.New("list failed")
		backend := &localImageListBackendStub{images: warmSnapshot}
		storage := warmedStorage(ctx, backend)

		backend.images, backend.err = nil, listErr
		_, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0)
		gomega.Expect(err).To(gomega.MatchError(listErr))

		backend.err = nil
		gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.ConsistOf(cachedTagA))
		gomega.Expect(cachedStages(ctx, storage, cachedDigestB)).To(gomega.ConsistOf(cachedTagB))
		gomega.Expect(backend.calls).To(gomega.Equal(2))
	})

	ginkgo.It("initializes the snapshot from a fresh listing", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: warmSnapshot}
		storage := NewLocalStagesStorage(backend)

		fresh, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(fresh)).To(gomega.ConsistOf(cachedTagA))

		gomega.Expect(cachedStages(ctx, storage, cachedDigestB)).To(gomega.ConsistOf(cachedTagB))
		gomega.Expect(backend.calls).To(gomega.Equal(1))
	})

	ginkgo.It("caches the fresh listing before the parent timestamp filter", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: warmSnapshot}
		storage := warmedStorage(ctx, backend)

		backend.images = image.ImagesList{
			{RepoTags: []string{"project:" + cachedTagA}},
			{RepoTags: []string{"project:" + cachedTagA2}},
		}
		filtered, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 1700000000002)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageStrings(filtered)).To(gomega.ConsistOf(cachedTagA2))

		gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.ConsistOf(cachedTagA, cachedTagA2))
	})

	ginkgo.DescribeTable("makes a stage published by this process visible to cached lookups",
		func(ctx ginkgo.SpecContext, publish func(ctx ginkgo.SpecContext, storage *LocalStagesStorage, backend *localPublishBackendStub) error, expectedTags []string) {
			backend := newLocalPublishBackendStub(image.ImagesList{{RepoTags: []string{"project:" + cachedTagB}}})
			storage := NewLocalStagesStorage(backend)
			gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.BeEmpty())

			err := publish(ctx, storage, backend)
			if len(expectedTags) == 0 {
				gomega.Expect(err).To(gomega.HaveOccurred())
			} else {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}

			gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.ConsistOf(expectedTags))
			gomega.Expect(cachedStages(ctx, storage, cachedDigestB)).To(gomega.ConsistOf(cachedTagB))
			gomega.Expect(backend.calls).To(gomega.Equal(1))
		},
		ginkgo.Entry("a tagged stage image", func(ctx ginkgo.SpecContext, storage *LocalStagesStorage, _ *localPublishBackendStub) error {
			return storage.StoreImage(ctx, &localStageImageStub{name: "project:" + cachedTagA})
		}, []string{cachedTagA}),
		ginkgo.Entry("a stage image that could not be tagged", func(ctx ginkgo.SpecContext, storage *LocalStagesStorage, backend *localPublishBackendStub) error {
			backend.tagImageErr = errors.New("tag failed")
			return storage.StoreImage(ctx, &localStageImageStub{name: "project:" + cachedTagA})
		}, nil),
		ginkgo.Entry("a natively mutated stage image", func(ctx ginkgo.SpecContext, storage *LocalStagesStorage, _ *localPublishBackendStub) error {
			return storage.MutateAndPushImage(ctx, "project:"+cachedTagB, "project:"+cachedTagA, image.SpecConfig{}, &localStageImageStub{name: "project:" + cachedTagB})
		}, []string{cachedTagA}),
		ginkgo.Entry("a stage image mutated through the save/load fallback", func(ctx ginkgo.SpecContext, storage *LocalStagesStorage, backend *localPublishBackendStub) error {
			backend.nativeErr = container_backend.ErrNativeMutationUnsupported
			return storage.MutateAndPushImage(ctx, "project:"+cachedTagB, "project:"+cachedTagA, image.SpecConfig{}, &localStageImageStub{name: "project:" + cachedTagB})
		}, []string{cachedTagA}),
		ginkgo.Entry("a stage image whose native mutation failed", func(ctx ginkgo.SpecContext, storage *LocalStagesStorage, backend *localPublishBackendStub) error {
			backend.nativeErr = errors.New("mutation failed")
			return storage.MutateAndPushImage(ctx, "project:"+cachedTagB, "project:"+cachedTagA, image.SpecConfig{}, &localStageImageStub{name: "project:" + cachedTagB})
		}, nil),
		ginkgo.Entry("a stage image whose fallback tag failed", func(ctx ginkgo.SpecContext, storage *LocalStagesStorage, backend *localPublishBackendStub) error {
			backend.nativeErr = container_backend.ErrNativeMutationUnsupported
			backend.tagErr = errors.New("tag failed")
			return storage.MutateAndPushImage(ctx, "project:"+cachedTagB, "project:"+cachedTagA, image.SpecConfig{}, &localStageImageStub{name: "project:" + cachedTagB})
		}, nil),
		ginkgo.Entry("a Buildah stage image", func(ctx ginkgo.SpecContext, storage *LocalStagesStorage, _ *localPublishBackendStub) error {
			return storage.StoreImage(ctx, &localStageImageStub{name: "localhost/project:" + cachedTagA})
		}, []string{cachedTagA}),
	)

	ginkgo.It("keeps a stage published right after a listing started", func(ctx ginkgo.SpecContext) {
		backend := newLocalPublishBackendStub(nil)
		storage := NewLocalStagesStorage(backend)
		gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.BeEmpty())

		listingStartedSeq := storage.nextSequence()
		gomega.Expect(storage.StoreImage(ctx, &localStageImageStub{name: "project:" + cachedTagA})).To(gomega.Succeed())
		storage.storeProjectSnapshot("project", nil, listingStartedSeq)

		gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.ConsistOf(cachedTagA), "a publication recorded after the listing started must survive that listing, however close the two are")
	})

	ginkgo.DescribeTable("does not lose a stage published while a listing is in flight", func(ctx ginkgo.SpecContext, fresh bool) {
		backend := newLocalPublishBackendStub(nil)
		storage := NewLocalStagesStorage(backend)
		if fresh {
			gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.BeEmpty())
		}
		listing, release := blockNextListing(backend.localImageListBackendStub)

		lookup := make(chan []string, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			var opts []Option
			if !fresh {
				opts = append(opts, WithCache())
			}
			stages, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, opts...)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			lookup <- stageStrings(stages)
		}()
		<-listing

		stored := make(chan error, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			stored <- storage.StoreImage(ctx, &localStageImageStub{name: "project:" + cachedTagA})
		}()
		gomega.Eventually(stored, blockedCallTimeout).Should(gomega.Receive(gomega.BeNil()))
		close(release)

		gomega.Eventually(lookup, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagA)), "the listing must not drop the stage published while it was running")
		gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.ConsistOf(cachedTagA))
		expectedCalls := 1
		if fresh {
			expectedCalls++
		}
		gomega.Expect(backend.calls).To(gomega.Equal(expectedCalls))
	},
		ginkgo.Entry("project warm-up", false),
		ginkgo.Entry("project refresh", true),
	)

	ginkgo.It("drops a published stage absent from a later fresh listing", func(ctx ginkgo.SpecContext) {
		backend := newLocalPublishBackendStub(nil)
		storage := NewLocalStagesStorage(backend)
		gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.BeEmpty())

		gomega.Expect(storage.StoreImage(ctx, &localStageImageStub{name: "project:" + cachedTagA})).To(gomega.Succeed())
		gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.ConsistOf(cachedTagA))

		fresh, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fresh).To(gomega.BeEmpty())
		gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.BeEmpty())
	})

	ginkgo.It("serves cached lookups and records publications while a listing is in flight", func(ctx ginkgo.SpecContext) {
		backend := newLocalPublishBackendStub(image.ImagesList{{RepoTags: []string{"project:" + cachedTagB}}})
		storage := NewLocalStagesStorage(backend)
		gomega.Expect(cachedStages(ctx, storage, cachedDigestB)).To(gomega.ConsistOf(cachedTagB))
		listing, release := blockNextListing(backend.localImageListBackendStub)

		refreshed := make(chan []string, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			stages, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestB, 0)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			refreshed <- stageStrings(stages)
		}()
		<-listing

		warm := make(chan []string, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			warm <- cachedStages(ctx, storage, cachedDigestB)
		}()
		stored := make(chan error, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			stored <- storage.StoreImage(ctx, &localStageImageStub{name: "project:" + cachedTagA})
		}()

		gomega.Eventually(warm, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagB)))
		gomega.Eventually(stored, blockedCallTimeout).Should(gomega.Receive(gomega.BeNil()))
		close(release)

		gomega.Eventually(refreshed, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagB)), "the listing in flight must answer the fresh lookup that started it")
		gomega.Expect(cachedStages(ctx, storage, cachedDigestA)).To(gomega.ConsistOf(cachedTagA))
		gomega.Expect(backend.calls).To(gomega.Equal(2))
	})

	ginkgo.It("lists the project once for concurrent cold lookups of different digests", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: warmSnapshot}
		storage := NewLocalStagesStorage(backend)
		listing, release := blockNextListing(backend)

		first := make(chan []string, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			first <- cachedStages(ctx, storage, cachedDigestA)
		}()
		<-listing

		joiningCtx, joined := listingRegistrations(ctx)
		second := make(chan []string, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			second <- cachedStages(joiningCtx, storage, cachedDigestB)
		}()
		gomega.Eventually(joined, blockedCallTimeout).Should(gomega.Receive())
		gomega.Consistently(second).ShouldNot(gomega.Receive())
		close(release)

		gomega.Eventually(first, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagA)))
		gomega.Eventually(second, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagB)))
		gomega.Expect(backend.calls).To(gomega.Equal(1))
	})

	ginkgo.It("reports which lookups joined the listing another one ran", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: warmSnapshot}
		storage := NewLocalStagesStorage(backend)
		listing, release := blockNextListing(backend)

		leader := make(chan bool, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			_, joined, err := storage.refreshProjectListing(ctx, "project", 0)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			leader <- joined
		}()
		<-listing

		joiningCtx, registered := listingRegistrations(ctx)
		joiner := make(chan bool, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			_, joined, err := storage.refreshProjectListing(joiningCtx, "project", 0)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			joiner <- joined
		}()
		gomega.Eventually(registered, blockedCallTimeout).Should(gomega.Receive())
		close(release)

		gomega.Eventually(leader, blockedCallTimeout).Should(gomega.Receive(gomega.BeFalse()), "the lookup that started the listing must not report a join")
		gomega.Eventually(joiner, blockedCallTimeout).Should(gomega.Receive(gomega.BeTrue()), "the lookup admitted to a listing in flight must report a join")
		gomega.Expect(backend.callCount()).To(gomega.Equal(1))
	})

	ginkgo.It("keeps reporting the join of a lookup canceled while it waited", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: warmSnapshot}
		storage := NewLocalStagesStorage(backend)
		listing, release := blockNextListing(backend)

		go func() {
			defer ginkgo.GinkgoRecover()
			_, _, _ = storage.refreshProjectListing(ctx, "project", 0)
		}()
		<-listing

		joinerCtx, cancelJoiner := context.WithCancel(ctx)
		joiningCtx, registered := listingRegistrations(joinerCtx)
		joiner := make(chan bool, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			_, joined, err := storage.refreshProjectListing(joiningCtx, "project", 0)
			gomega.Expect(err).To(gomega.MatchError(context.Canceled))
			joiner <- joined
		}()
		gomega.Eventually(registered, blockedCallTimeout).Should(gomega.Receive())
		cancelJoiner()

		// Giving up on a listing somebody else is running does not unmake the join: the role is
		// settled when the caller is admitted, before it waits for anything.
		gomega.Eventually(joiner, blockedCallTimeout).Should(gomega.Receive(gomega.BeTrue()), "a canceled joiner must still report the join it was admitted to")
		close(release)
	})

	ginkgo.It("makes every fresh lookup wait for a listing started after it", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: warmSnapshot}
		storage := NewLocalStagesStorage(backend)
		started := make(chan int, 4)
		releases := []chan struct{}{make(chan struct{}), make(chan struct{})}
		ginkgo.DeferCleanup(func() {
			for _, release := range releases {
				closeIfOpen(release)
			}
		})
		backend.onList = func(listing int) {
			if listing == 2 {
				backend.images = image.ImagesList{{RepoTags: []string{"project:" + cachedTagA, "project:" + cachedTagA2}}}
			}
			started <- listing
			if listing <= len(releases) {
				<-releases[listing-1]
			}
		}

		warm := make(chan []string, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			warm <- cachedStages(ctx, storage, cachedDigestA)
		}()
		gomega.Eventually(started, blockedCallTimeout).Should(gomega.Receive(gomega.Equal(1)))

		fresh := make(chan []string, 2)
		registrations := make([]chan struct{}, 2)
		for i := range registrations {
			freshCtx, registered := listingRegistrations(ctx)
			registrations[i] = registered
			go func() {
				defer ginkgo.GinkgoRecover()
				stages, err := storage.GetStagesIDsByDigest(freshCtx, "project", cachedDigestA, 0)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				fresh <- stageStrings(stages)
			}()
		}
		for _, registered := range registrations {
			gomega.Eventually(registered, blockedCallTimeout).Should(gomega.Receive())
		}
		close(releases[0])
		gomega.Consistently(fresh).ShouldNot(gomega.Receive())

		gomega.Eventually(warm, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagA)))
		gomega.Eventually(started, blockedCallTimeout).Should(gomega.Receive(gomega.Equal(2)))
		for _, registered := range registrations {
			gomega.Eventually(registered, blockedCallTimeout).Should(gomega.Receive())
		}
		close(releases[1])

		gomega.Eventually(fresh, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagA, cachedTagA2)), "a fresh lookup must be answered by the listing started after it")
		gomega.Eventually(fresh, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagA, cachedTagA2)), "both fresh lookups must share the listing started after them")
		gomega.Expect(backend.callCount()).To(gomega.Equal(2))
	})

	ginkgo.It("starts another listing for a fresh lookup that arrives during the second one", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: warmSnapshot}
		storage := NewLocalStagesStorage(backend)
		started, releases := blockListings(backend, 3)

		results := make(chan []string, 3)
		freshLookup := func() chan struct{} {
			freshCtx, registered := listingRegistrations(ctx)
			go func() {
				defer ginkgo.GinkgoRecover()
				stages, err := storage.GetStagesIDsByDigest(freshCtx, "project", cachedDigestA, 0)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				results <- stageStrings(stages)
			}()
			return registered
		}

		freshLookup()
		gomega.Eventually(started, blockedCallTimeout).Should(gomega.Receive(gomega.Equal(1)))

		secondRegistered := freshLookup()
		gomega.Eventually(secondRegistered, blockedCallTimeout).Should(gomega.Receive())
		close(releases[0])
		gomega.Eventually(results, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagA)))
		gomega.Consistently(results).ShouldNot(gomega.Receive())
		gomega.Eventually(started, blockedCallTimeout).Should(gomega.Receive(gomega.Equal(2)))

		thirdRegistered := freshLookup()
		gomega.Eventually(thirdRegistered, blockedCallTimeout).Should(gomega.Receive())
		close(releases[1])
		gomega.Eventually(results, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagA)))
		gomega.Consistently(results).ShouldNot(gomega.Receive())
		gomega.Eventually(started, blockedCallTimeout).Should(gomega.Receive(gomega.Equal(3)))
		close(releases[2])

		gomega.Eventually(results, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagA)))
		gomega.Expect(backend.callCount()).To(gomega.Equal(3))
	})

	ginkgo.DescribeTable("does not list images for a lookup canceled before it started", func(ctx ginkgo.SpecContext, fresh bool) {
		backend := &localImageListBackendStub{images: warmSnapshot}
		storage := NewLocalStagesStorage(backend)
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		var opts []Option
		if !fresh {
			opts = append(opts, WithCache())
		}
		_, err := storage.GetStagesIDsByDigest(canceledCtx, "project", cachedDigestA, 0, opts...)
		gomega.Expect(err).To(gomega.MatchError(context.Canceled))
		gomega.Consistently(backend.callCount).Should(gomega.BeZero())
	},
		ginkgo.Entry("a fresh lookup", true),
		ginkgo.Entry("a cold cached lookup", false),
	)

	ginkgo.It("releases a canceled lookup waiting for a listing", func(ctx ginkgo.SpecContext) {
		backend := &localImageListBackendStub{images: warmSnapshot}
		storage := NewLocalStagesStorage(backend)
		listing, release := blockNextListing(backend)

		leader := make(chan []string, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			leader <- cachedStages(ctx, storage, cachedDigestA)
		}()
		<-listing

		canceledCtx, cancel := context.WithCancel(ctx)
		waiter := make(chan error, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			_, err := storage.GetStagesIDsByDigest(canceledCtx, "project", cachedDigestB, 0, WithCache())
			waiter <- err
		}()
		gomega.Consistently(waiter).ShouldNot(gomega.Receive())

		cancel()
		gomega.Eventually(waiter, blockedCallTimeout).Should(gomega.Receive(gomega.MatchError(context.Canceled)))
		close(release)

		gomega.Eventually(leader, blockedCallTimeout).Should(gomega.Receive(gomega.ConsistOf(cachedTagA)))
		gomega.Expect(backend.calls).To(gomega.Equal(1))
	})
})
