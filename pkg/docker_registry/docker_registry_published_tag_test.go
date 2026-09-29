package docker_registry

import (
	"fmt"
	"sync"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("AddCachedTag", func() {
	const repo = "registry.example/project"

	ginkgo.It("does not create a cache entry for a repo that was never listed", func(ctx ginkgo.SpecContext) {
		inner := newListingRegistryStub("stage-a")
		r := newCachedRegistryStub(inner)

		AddCachedTag(ctx, r, repo+":stage-b")

		_, ok := r.cachedTagsMap.Load(repo)
		gomega.Expect(ok).To(gomega.BeFalse())

		tags, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"stage-a"}))
	})

	ginkgo.It("serves a published tag from the cached listing without listing again", func(ctx ginkgo.SpecContext) {
		inner := newListingRegistryStub("stage-a")
		r := newCachedRegistryStub(inner)

		_, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(inner.callCount()).To(gomega.Equal(1))

		AddCachedTag(ctx, r, repo+":stage-b")

		exists, err := r.IsTagExist(ctx, repo+":stage-b", WithCachedTags())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(exists).To(gomega.BeTrue())
		gomega.Expect(inner.callCount()).To(gomega.Equal(1))
	})

	ginkgo.It("does not modify the tags slice already returned to a caller", func(ctx ginkgo.SpecContext) {
		r := newCachedRegistryStub(newListingRegistryStub())
		// Spare capacity: an in-place append would be visible to a caller holding the slice.
		tags := make([]string, 1, 2)
		tags[0] = "stage-a"
		r.cachedTagsMap.Store(repo, tagsCacheEntry{tags: tags, updatedAt: time.Now()})

		AddCachedTag(ctx, r, repo+":stage-b")

		gomega.Expect(tags).To(gomega.Equal([]string{"stage-a"}))
		gomega.Expect(tags[:cap(tags)][1]).To(gomega.BeEmpty())
		gomega.Expect(cachedEntry(r, repo).tags).To(gomega.Equal([]string{"stage-a", "stage-b"}))
	})

	ginkgo.It("keeps the freshness of the cached listing unchanged", func(ctx ginkgo.SpecContext) {
		r := newCachedRegistryStub(newListingRegistryStub())
		updatedAt := time.Now().Add(-time.Hour)
		r.cachedTagsMap.Store(repo, tagsCacheEntry{tags: []string{"stage-a"}, updatedAt: updatedAt})

		AddCachedTag(ctx, r, repo+":stage-b")

		gomega.Expect(cachedEntry(r, repo).updatedAt).To(gomega.Equal(updatedAt))
	})

	ginkgo.It("does not duplicate a tag that the cached listing already contains", func(ctx ginkgo.SpecContext) {
		r := newCachedRegistryStub(newListingRegistryStub())
		r.cachedTagsMap.Store(repo, tagsCacheEntry{tags: []string{"stage-a"}, updatedAt: time.Now()})

		AddCachedTag(ctx, r, repo+":stage-a")

		gomega.Expect(cachedEntry(r, repo).tags).To(gomega.Equal([]string{"stage-a"}))
	})

	ginkgo.It("keeps every tag published concurrently", func(ctx ginkgo.SpecContext) {
		r := newCachedRegistryStub(newListingRegistryStub())
		r.cachedTagsMap.Store(repo, tagsCacheEntry{tags: []string{"stage-a"}, updatedAt: time.Now()})

		var wg sync.WaitGroup
		expectedTags := []string{"stage-a"}
		for i := range 16 {
			tag := fmt.Sprintf("stage-%d", i)
			expectedTags = append(expectedTags, tag)
			wg.Add(1)
			go func() {
				defer ginkgo.GinkgoRecover()
				defer wg.Done()
				AddCachedTag(ctx, r, repo+":"+tag)
			}()
		}
		wg.Wait()

		gomega.Expect(cachedEntry(r, repo).tags).To(gomega.ConsistOf(expectedTags))
	})

	ginkgo.It("is a no-op for a registry without the tags cache", func(ctx ginkgo.SpecContext) {
		gomega.Expect(func() { AddCachedTag(ctx, newListingRegistryStub(), repo+":stage-b") }).NotTo(gomega.Panic())
	})

	ginkgo.It("panics on a reference without a parsable tag", func(ctx ginkgo.SpecContext) {
		r := newCachedRegistryStub(newListingRegistryStub())

		gomega.Expect(func() { AddCachedTag(ctx, r, "not a valid reference!") }).To(gomega.Panic())
	})
})

var _ = ginkgo.Describe("AddCachedTag concurrent with a tags listing", func() {
	const repo = "registry.example/project"

	ginkgo.It("keeps a tag published while a listing started earlier is in flight", func(ctx ginkgo.SpecContext) {
		inner := newListingRegistryStub("stage-a")
		r := newCachedRegistryStub(inner)

		// Prime the cache, so that the publication has an entry to update.
		_, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		listedTags, release := startBlockedListing(ctx, r, inner, repo)
		AddCachedTag(ctx, r, repo+":stage-b")
		release()

		// The listing neither loses the tag for its own caller nor for the cache.
		gomega.Expect(<-listedTags).To(gomega.ConsistOf("stage-a", "stage-b"))
		gomega.Expect(cachedEntry(r, repo).tags).To(gomega.ConsistOf("stage-a", "stage-b"))
	})

	ginkgo.It("keeps a cached tag republished while a listing that no longer sees it is in flight", func(ctx ginkgo.SpecContext) {
		inner := newListingRegistryStub("stage-a", "stage-b")
		r := newCachedRegistryStub(inner)

		_, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// The listing snapshots the repo without stage-b (deleted externally), while stage-b is
		// republished before it returns.
		inner.setTags("stage-a")
		listedTags, release := startBlockedListing(ctx, r, inner, repo)
		AddCachedTag(ctx, r, repo+":stage-b")
		release()

		gomega.Expect(<-listedTags).To(gomega.ConsistOf("stage-a", "stage-b"))
		gomega.Expect(cachedEntry(r, repo).tags).To(gomega.ConsistOf("stage-a", "stage-b"))
	})

	ginkgo.It("drops a published tag that a later listing does not report", func(ctx ginkgo.SpecContext) {
		inner := newListingRegistryStub("stage-a")
		r := newCachedRegistryStub(inner)

		_, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		AddCachedTag(ctx, r, repo+":stage-b")
		gomega.Expect(cachedEntry(r, repo).tags).To(gomega.ConsistOf("stage-a", "stage-b"))

		tags, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"stage-a"}))
		gomega.Expect(cachedEntry(r, repo).tags).To(gomega.Equal([]string{"stage-a"}))
	})

	ginkgo.It("keeps the published tag reported by a later listing without duplicating it", func(ctx ginkgo.SpecContext) {
		inner := newListingRegistryStub("stage-a")
		r := newCachedRegistryStub(inner)

		_, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		AddCachedTag(ctx, r, repo+":stage-b")
		inner.setTags("stage-a", "stage-b")

		tags, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"stage-a", "stage-b"}))
		gomega.Expect(cachedEntry(r, repo).pushedTags).To(gomega.BeEmpty())
	})

	ginkgo.It("leaves the cache untouched when the listing fails", func(ctx ginkgo.SpecContext) {
		inner := newListingRegistryStub("stage-a")
		r := newCachedRegistryStub(inner)

		_, err := r.Tags(ctx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		AddCachedTag(ctx, r, repo+":stage-b")

		inner.failWith = fmt.Errorf("registry is down")
		_, err = r.Tags(ctx, repo)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("registry is down")))

		gomega.Expect(cachedEntry(r, repo).tags).To(gomega.ConsistOf("stage-a", "stage-b"))
	})
})
