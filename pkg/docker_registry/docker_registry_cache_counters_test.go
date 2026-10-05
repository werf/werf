package docker_registry

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/opstats"
)

var _ = ginkgo.Describe("registry tags cache counters", func() {
	const repo = "example.org/project"

	ginkgo.It("counts a listing that asked for fresh tags as a bypass", func(specCtx ginkgo.SpecContext) {
		ctx, collector := collectingContext(specCtx)
		r := newCachedRegistryStub(newListingRegistryStub("stage-a"))

		_, err := r.Tags(ctx, repo, WithTagsMaxAge(0))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(tagsListCounters(collector, ctx)).To(gomega.Equal(opstats.CacheSummary{
			Operation: opstats.OperationRegistryTagsList,
			Layer:     opstats.CacheLayerMemory,
			Bypass:    1,
		}))
	})

	ginkgo.It("counts a cold listing as a non-shared miss", func(specCtx ginkgo.SpecContext) {
		ctx, collector := collectingContext(specCtx)
		r := newCachedRegistryStub(newListingRegistryStub("stage-a"))

		_, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(tagsListCounters(collector, ctx)).To(gomega.Equal(opstats.CacheSummary{
			Operation: opstats.OperationRegistryTagsList,
			Layer:     opstats.CacheLayerMemory,
			Miss:      1,
		}))
	})

	ginkgo.DescribeTable("counts a usable cached listing as a hit",
		func(specCtx ginkgo.SpecContext, tags []string) {
			ctx, collector := collectingContext(specCtx)
			inner := newListingRegistryStub(tags...)
			r := newCachedRegistryStub(inner)

			_, err := r.Tags(ctx, repo)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			cached, err := r.Tags(ctx, repo, WithCachedTags())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(cached).To(gomega.HaveLen(len(tags)))
			gomega.Expect(inner.callCount()).To(gomega.Equal(1))

			gomega.Expect(tagsListCounters(collector, ctx)).To(gomega.Equal(opstats.CacheSummary{
				Operation: opstats.OperationRegistryTagsList,
				Layer:     opstats.CacheLayerMemory,
				Hit:       1,
				Miss:      1,
			}))
		},
		ginkgo.Entry("a listing with tags", []string{"stage-a"}),
		// An empty listing is a usable answer, so serving it from the cache is a hit and not a
		// miss that happens to return nothing.
		ginkgo.Entry("an empty listing", []string{}),
	)

	ginkgo.It("keeps the classification of a failed listing and counts it once", func(specCtx ginkgo.SpecContext) {
		ctx, collector := collectingContext(specCtx)
		inner := newListingRegistryStub()
		inner.failWith = errors.New("registry unavailable")
		r := newCachedRegistryStub(inner)

		_, err := r.Tags(ctx, repo)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("registry unavailable")))

		gomega.Expect(tagsListCounters(collector, ctx)).To(gomega.Equal(opstats.CacheSummary{
			Operation: opstats.OperationRegistryTagsList,
			Layer:     opstats.CacheLayerMemory,
			Miss:      1,
		}))
	})

	ginkgo.It("keeps the classification of a canceled listing and counts it once", func(specCtx ginkgo.SpecContext) {
		ctx, collector := collectingContext(specCtx)
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		inner := newListingRegistryStub()
		inner.failWith = context.Canceled
		r := newCachedRegistryStub(inner)

		_, err := r.Tags(canceledCtx, repo)
		gomega.Expect(err).To(gomega.MatchError(context.Canceled))

		gomega.Expect(tagsListCounters(collector, ctx)).To(gomega.Equal(opstats.CacheSummary{
			Operation: opstats.OperationRegistryTagsList,
			Layer:     opstats.CacheLayerMemory,
			Miss:      1,
		}))
	})

	ginkgo.It("counts the caller that started the listing as a non-shared miss and every joiner as a shared miss", func(specCtx ginkgo.SpecContext) {
		const joiners = 4

		ctx, collector := collectingContext(specCtx)
		inner := newAdmittingRegistryStub("stage-a")
		r := newCachedRegistryStub(inner)

		done := make(chan struct{})
		var wg sync.WaitGroup
		ginkgo.DeferCleanup(func() {
			inner.release()
			gomega.Eventually(done, 30*time.Second).Should(gomega.BeClosed())
		})

		for range joiners + 1 {
			wg.Add(1)
			go func() {
				// wg.Done is deferred first so that it runs after GinkgoRecover, and the spec
				// never ends while a caller is still unwinding.
				defer wg.Done()
				defer ginkgo.GinkgoRecover()
				tags, err := r.Tags(ctx, repo)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(tags).To(gomega.Equal([]string{"stage-a"}))
			}()
		}
		go func() {
			wg.Wait()
			close(done)
		}()

		// Admit the listing only once it has started and every joiner has registered with
		// singleflight: a joiner that merely entered Do could still start a second listing.
		gomega.Eventually(func() bool {
			return inner.callCount() == 1 && joinersWaitingForTagsList() == joiners
		}).Should(gomega.BeTrue())
		inner.release()
		gomega.Eventually(done, 30*time.Second).Should(gomega.BeClosed())

		gomega.Expect(inner.callCount()).To(gomega.Equal(1))
		gomega.Expect(tagsListCounters(collector, ctx)).To(gomega.Equal(opstats.CacheSummary{
			Operation: opstats.OperationRegistryTagsList,
			Layer:     opstats.CacheLayerMemory,
			Miss:      joiners + 1,
			Shared:    joiners,
		}))
	})

	ginkgo.It("counts a lookup that bypassed the cache as non-shared", func(specCtx ginkgo.SpecContext) {
		ctx, collector := collectingContext(specCtx)
		inner := newListingRegistryStub("stage-a")
		r := newCachedRegistryStub(inner)

		_, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		_, err = r.Tags(ctx, repo, WithTagsMaxAge(0))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(tagsListCounters(collector, ctx)).To(gomega.Equal(opstats.CacheSummary{
			Operation: opstats.OperationRegistryTagsList,
			Layer:     opstats.CacheLayerMemory,
			Miss:      1,
			Bypass:    1,
		}))
	})
})

var _ = ginkgo.Describe("the barrier admitting a shared tags listing", func() {
	const dump = `goroutine 21 [chan receive]:
github.com/werf/werf/v3/pkg/docker_registry.(*admittingRegistryStub).Tags(0xc0001)
golang.org/x/sync/singleflight.(*Group).doCall(0xc0002)
golang.org/x/sync/singleflight.(*Group).Do(0xc0002)
github.com/werf/werf/v3/pkg/docker_registry.(*DockerRegistryWithCache).getTagsListFromRegistry(0xc0003)

goroutine 22 [semacquire]:
sync.runtime_SemacquireWaitGroup(0xc0004)
sync.(*WaitGroup).Wait(0xc0002)
golang.org/x/sync/singleflight.(*Group).Do(0xc0002)
github.com/werf/werf/v3/pkg/docker_registry.(*DockerRegistryWithCache).getTagsListFromRegistry(0xc0003)

goroutine 23 [runnable]:
github.com/werf/werf/v3/pkg/docker_registry.(*DockerRegistryWithCache).getTagsListFromRegistry(0xc0003)

goroutine 24 [semacquire]:
sync.(*WaitGroup).Wait(0xc0005)
github.com/werf/werf/v3/pkg/docker_registry_test.somethingElse(0xc0006)
`

	ginkgo.It("counts only the callers already registered as waiters of the listing", func() {
		gomega.Expect(goroutinesWithFrames(dump,
			"(*DockerRegistryWithCache).getTagsListFromRegistry(",
			"singleflight.(*Group).Do(",
			"sync.(*WaitGroup).Wait(")).To(gomega.Equal(1))
	})
})
