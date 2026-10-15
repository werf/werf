package docker_registry

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Publications with uncached prefix discovery", func() {
	ginkgo.It("keeps own publications across an older full listing and accepts later deletions", func(ctx ginkgo.SpecContext) {
		fixture := newPrefixRegistryFixture([]string{"before"}, "")
		registry, err := NewDockerRegistry(ctx, fixture.repository(), DefaultImplementationName, DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		cache := registry.(*DockerRegistryWithCache)
		_, err = registry.Tags(ctx, fixture.repository(), WithCachedTags())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		older := cache.nextListingSequence()
		AddCachedTag(ctx, registry, fixture.repository()+":winner")
		cache.storeTagsToCache(fixture.repository(), []string{"before"}, older)
		tags, err := registry.Tags(ctx, fixture.repository(), WithCachedTags())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.ConsistOf("before", "winner"))
		cache.storeTagsToCache(fixture.repository(), []string{"before"}, cache.nextListingSequence())
		tags, err = registry.Tags(ctx, fixture.repository(), WithCachedTags())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"before"}))
	})

	ginkgo.It("returns the newer inventory when an older listing finishes last", func(ctx ginkgo.SpecContext) {
		fixture := newPrefixRegistryFixture([]string{"before"}, "")
		registry, err := NewDockerRegistry(ctx, fixture.repository(), DefaultImplementationName, DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		cache := registry.(*DockerRegistryWithCache)
		older := cache.nextListingSequence()
		cache.storeTagsToCache(fixture.repository(), []string{"before", "winner"}, cache.nextListingSequence())
		result := cache.storeTagsToCache(fixture.repository(), []string{"before"}, older)
		gomega.Expect(result).To(gomega.Equal([]string{"before", "winner"}))
		tags, err := registry.Tags(ctx, fixture.repository(), WithCachedTags())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal(result))
	})

	ginkgo.It("does not manufacture a complete inventory from a publication", func(ctx ginkgo.SpecContext) {
		fixture := newPrefixRegistryFixture([]string{"before"}, "")
		registry, err := NewDockerRegistry(ctx, fixture.repository(), DefaultImplementationName, DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		AddCachedTag(ctx, registry, fixture.repository()+":winner")
		_, loaded := registry.(*DockerRegistryWithCache).cachedTagsMap.Load(fixture.repository())
		gomega.Expect(loaded).To(gomega.BeFalse())
		tags, err := registry.Tags(ctx, fixture.repository(), WithCachedTags())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"before"}))
	})
})
