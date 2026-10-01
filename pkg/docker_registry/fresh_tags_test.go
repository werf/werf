package docker_registry

import (
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Authoritative tag listing", func() {
	ginkgo.It("does not join a snapshot started before publication", func(ctx ginkgo.SpecContext) {
		inner := &snapshotListingRegistry{started: make(chan struct{}), release: make(chan struct{})}
		registry := newCachedRegistryStub(inner)
		stale := make(chan []string, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			tags, err := registry.Tags(ctx, "example.test/repo")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			stale <- tags
		}()
		ginkgo.DeferCleanup(func() { close(inner.release) })
		gomega.Eventually(inner.started).Should(gomega.BeClosed())
		fresh := make(chan []string, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			tags, err := registry.Tags(ctx, "example.test/repo", WithTagsMaxAge(0))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			fresh <- tags
		}()
		gomega.Eventually(fresh).Should(gomega.Receive(gomega.Equal([]string{"before", "winner"})))
		gomega.Expect(inner.calls.Load()).To(gomega.Equal(int32(2)))
		gomega.Expect(cachedEntry(registry, "example.test/repo").tags).To(gomega.ContainElement("winner"))
	})

	ginkgo.It("bypasses a cached negative listing even when cached tags were requested", func(ctx ginkgo.SpecContext) {
		inner := newListingRegistryStub("winner")
		registry := newCachedRegistryStub(inner)
		registry.cachedTagsMap.Store("example.test/repo", tagsCacheEntry{tags: []string{}, updatedAt: time.Now()})
		tags, err := registry.Tags(ctx, "example.test/repo", WithCachedTags(), WithTagsMaxAge(0))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"winner"}))
		gomega.Expect(inner.callCount()).To(gomega.Equal(1))
	})
})
