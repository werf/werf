package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/opstats"
)

var _ = ginkgo.Describe("Local stage lookup cache counters", func() {
	ginkgo.It("counts a lookup without the cache option as a bypass", func(specCtx ginkgo.SpecContext) {
		ctx, collector := collectingContext(specCtx)
		backend := &localImageListBackendStub{images: image.ImagesList{{RepoTags: []string{"project:" + cachedTagA}}}}

		_, err := NewLocalStagesStorage(backend).GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(backend.options.Filters).To(gomega.Equal([]util.Pair[string, string]{
			util.NewPair("reference", "project"),
		}))
		gomega.Expect(collector.CacheSummary(ctx)).To(gomega.Equal([]opstats.CacheSummary{{
			Operation: opstats.OperationDockerImageList,
			Layer:     opstats.CacheLayerMemory,
			Bypass:    1,
		}}))
	})

	ginkgo.DescribeTable("counts the first cached lookup as a miss and the next one as a hit",
		func(specCtx ginkgo.SpecContext, images image.ImagesList) {
			ctx, collector := collectingContext(specCtx)
			storage := NewLocalStagesStorage(&localImageListBackendStub{images: images})

			for range 2 {
				_, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}

			gomega.Expect(collector.CacheSummary(ctx)).To(gomega.Equal([]opstats.CacheSummary{{
				Operation: opstats.OperationDockerImageList,
				Layer:     opstats.CacheLayerMemory,
				Hit:       1,
				Miss:      1,
			}}))
		},
		ginkgo.Entry("a snapshot with stages", image.ImagesList{{RepoTags: []string{"project:" + cachedTagA}}}),
		// A ready snapshot answers the lookup even when the project has no images at all.
		ginkgo.Entry("an empty snapshot", image.ImagesList{}),
	)

	ginkgo.It("keeps the classification of a failed listing and counts it once", func(specCtx ginkgo.SpecContext) {
		ctx, collector := collectingContext(specCtx)
		backend := &localImageListBackendStub{err: errors.New("list failed")}

		_, err := NewLocalStagesStorage(backend).GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
		gomega.Expect(err).To(gomega.HaveOccurred())

		gomega.Expect(collector.CacheSummary(ctx)).To(gomega.Equal([]opstats.CacheSummary{{
			Operation: opstats.OperationDockerImageList,
			Layer:     opstats.CacheLayerMemory,
			Miss:      1,
		}}))
	})

	ginkgo.It("names the row after the Buildah backend identity", func(specCtx ginkgo.SpecContext) {
		ctx, collector := collectingContext(specCtx)
		backend := &localImageListBackendStub{name: "buildah-backend"}

		_, err := NewLocalStagesStorage(backend).GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(collector.CacheSummary(ctx)).To(gomega.Equal([]opstats.CacheSummary{{
			Operation: opstats.OperationBuildahImageList,
			Layer:     opstats.CacheLayerMemory,
			Miss:      1,
		}}))
	})

	ginkgo.DescribeTable("records no row for an unrecognized backend",
		func(specCtx ginkgo.SpecContext, backendName string) {
			ctx, collector := collectingContext(specCtx)
			backend := &localImageListBackendStub{name: backendName}

			_, err := NewLocalStagesStorage(backend).GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(collector.CacheSummary(ctx)).To(gomega.BeEmpty())
		},
		ginkgo.Entry("an unrelated one", "podman-backend"),
		// Only the exact identities of the known backends name a row; a docker-like name of some
		// other implementation does not make it the docker image list.
		ginkgo.Entry("a docker-like one", "docker-fancy-backend"),
	)

	ginkgo.It("counts concurrent cache misses that share a listing", func(specCtx ginkgo.SpecContext) {
		const waiters = 3

		ctx, collector := collectingContext(specCtx)
		ctx, registrations := listingRegistrations(ctx)
		entered, release := make(chan struct{}), make(chan struct{})
		var releaseOnce sync.Once
		releaseListing := func() { releaseOnce.Do(func() { close(release) }) }
		backend := &localImageListBackendStub{
			images: image.ImagesList{{RepoTags: []string{"project:" + cachedTagA}}},
			onList: func(_ int) { close(entered); <-release },
		}
		storage := NewLocalStagesStorage(backend)

		done := make(chan struct{})
		var wg sync.WaitGroup
		ginkgo.DeferCleanup(func() {
			releaseListing()
			gomega.Eventually(done, 30*time.Second).Should(gomega.BeClosed())
		})

		lookup := func() {
			// wg.Done is deferred first so that it runs after GinkgoRecover, and the spec never
			// ends while a caller is still unwinding.
			defer wg.Done()
			defer ginkgo.GinkgoRecover()
			_, err := storage.GetStagesIDsByDigest(ctx, "project", cachedDigestA, 0, WithCache())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}

		wg.Add(1)
		go lookup()
		gomega.Eventually(entered).Should(gomega.BeClosed())
		for range waiters {
			wg.Add(1)
			go lookup()
		}
		go func() {
			wg.Wait()
			close(done)
		}()

		for range waiters + 1 {
			gomega.Eventually(registrations).Should(gomega.Receive())
		}
		releaseListing()
		gomega.Eventually(done, 30*time.Second).Should(gomega.BeClosed())

		gomega.Expect(backend.callCount()).To(gomega.Equal(1))
		gomega.Expect(collector.CacheSummary(ctx)).To(gomega.Equal([]opstats.CacheSummary{{
			Operation: opstats.OperationDockerImageList,
			Layer:     opstats.CacheLayerMemory,
			Miss:      waiters + 1,
			Shared:    waiters,
		}}))
	})
})
