package docker_registry

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/opstats"
)

// collectingContext binds a fresh collector to ctx and returns both.
func collectingContext(ctx context.Context) (context.Context, *opstats.Collector) {
	collector := opstats.NewCollector()
	return opstats.NewContext(ctx, collector), collector
}

func tagsListCounters(collector *opstats.Collector, ctx context.Context) opstats.CacheSummary {
	summary := collector.CacheSummary(ctx)
	gomega.Expect(summary).To(gomega.HaveLen(1))
	gomega.Expect(summary[0].Operation).To(gomega.Equal(opstats.OperationRegistryTagsList))
	return summary[0]
}

// admittingRegistryStub lets the test hold the first (leader) listing in flight until every
// other caller is actually blocked inside singleflight, so that joiners are admitted by a
// barrier on observed state instead of by waiting for a sleep to elapse.
type admittingRegistryStub struct {
	Interface

	tags     []string
	admitted chan struct{}

	mu    sync.Mutex
	calls int
}

var _ Interface = (*admittingRegistryStub)(nil)

func (r *admittingRegistryStub) Tags(_ context.Context, _ string, _ ...Option) ([]string, error) {
	r.mu.Lock()
	r.calls++
	r.mu.Unlock()

	<-r.admitted

	return append([]string(nil), r.tags...), nil
}

func (r *admittingRegistryStub) parseReferenceParts(reference string) (referenceParts, error) {
	return (&api{}).parseReferenceParts(reference)
}

func (r *admittingRegistryStub) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

// callersInSingleflight counts the goroutines currently inside singleflight.Group.Do: the one
// running the shared call plus everyone waiting for its result.
func callersInSingleflight() int {
	buf := make([]byte, 1<<20)
	return strings.Count(string(buf[:runtime.Stack(buf, true)]), "singleflight.(*Group).Do(")
}

var _ = ginkgo.Describe("registry tags cache counters", func() {
	const repo = "example.org/project"

	ginkgo.It("counts a listing that asked for fresh tags as a bypass", func(specCtx ginkgo.SpecContext) {
		ctx, collector := collectingContext(specCtx)
		r := newCachedRegistryStub(newListingRegistryStub("stage-a"))

		_, err := r.Tags(ctx, repo, WithTagsMaxAge(0))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(tagsListCounters(collector, ctx)).To(gomega.Equal(opstats.CacheSummary{
			Operation: opstats.OperationRegistryTagsList,
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
			Miss:      1,
		}))
	})

	ginkgo.It("counts the caller that started the listing as a non-shared miss and every joiner as a shared miss", func(specCtx ginkgo.SpecContext) {
		const joiners = 4

		ctx, collector := collectingContext(specCtx)
		inner := &admittingRegistryStub{tags: []string{"stage-a"}, admitted: make(chan struct{})}
		r := newCachedRegistryStub(inner)

		var wg sync.WaitGroup
		for range joiners + 1 {
			wg.Add(1)
			go func() {
				defer ginkgo.GinkgoRecover()
				defer wg.Done()
				tags, err := r.Tags(ctx, repo)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(tags).To(gomega.Equal([]string{"stage-a"}))
			}()
		}
		// Release the shared listing only once every joiner is waiting for it.
		gomega.Eventually(callersInSingleflight).Should(gomega.Equal(joiners + 1))
		close(inner.admitted)
		wg.Wait()

		// One registry request for all of them: anything else means a joiner was not admitted
		// while the leader was still in flight.
		gomega.Expect(inner.callCount()).To(gomega.Equal(1))
		gomega.Expect(tagsListCounters(collector, ctx)).To(gomega.Equal(opstats.CacheSummary{
			Operation: opstats.OperationRegistryTagsList,
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
			Miss:      1,
			Bypass:    1,
		}))
	})
})
