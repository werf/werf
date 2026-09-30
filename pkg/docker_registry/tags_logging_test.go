package docker_registry

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v3/pkg/opstats"
)

var _ = ginkgo.It("logs registry tag requests but counts cache hits without repeated listings", func(ctx ginkgo.SpecContext) {
	var output bytes.Buffer
	logger := logboek.NewLogger(&output, &output)
	logger.SetAcceptedLevel(level.Debug)
	collector := opstats.NewCollector()
	runCtx := opstats.NewContext(logboek.NewContext(ctx, logger), collector)
	inner := newListingRegistryStub("one", "two")
	registry := newCachedRegistryStub(inner)
	const repo = "registry.example/project"
	tags, err := registry.Tags(runCtx, repo)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(tags).To(gomega.Equal([]string{"one", "two"}))
	gomega.Expect(output.String()).To(gomega.MatchRegexp(`Listed 2 tags for repo registry.example/project \(\d+\.\d{2} seconds\)`))
	for range 3 {
		tags, err = registry.Tags(runCtx, repo, WithCachedTags())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"one", "two"}))
	}
	gomega.Expect(inner.callCount()).To(gomega.Equal(1))
	gomega.Expect(strings.Count(output.String(), "Listed ")).To(gomega.Equal(1))
	gomega.Expect(collector.EventSummary()).To(gomega.ContainElement(opstats.EventSummary{Event: opstats.EventRegistryTagsCacheHit, Count: 3}))
	_, err = registry.Tags(runCtx, repo)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(inner.callCount()).To(gomega.Equal(2))
	gomega.Expect(strings.Count(output.String(), "Listed ")).To(gomega.Equal(2))
	inner.failWith = errors.New("registry unavailable")
	_, err = registry.Tags(runCtx, repo)
	gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("registry unavailable")))
	gomega.Expect(strings.Count(output.String(), "Listed ")).To(gomega.Equal(2))
})

var _ = ginkgo.It("logs one registry tag request shared by concurrent callers", func(ctx ginkgo.SpecContext) {
	var output bytes.Buffer
	logger := logboek.NewLogger(&output, &output)
	logger.SetAcceptedLevel(level.Debug)
	collector := opstats.NewCollector()
	runCtx := opstats.NewContext(logboek.NewContext(ctx, logger), collector)
	inner := newListingRegistryStub("one", "two")
	registry := newCachedRegistryStub(inner)
	const repo = "registry.example/project"
	first, release := startBlockedListing(runCtx, registry, inner, repo)
	var releaseOnce sync.Once
	defer releaseOnce.Do(release)
	second := make(chan []string, 1)
	go func() {
		defer ginkgo.GinkgoRecover()
		tags, err := registry.Tags(runCtx, repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		second <- tags
	}()
	gomega.Consistently(second, 100*time.Millisecond).ShouldNot(gomega.Receive())
	releaseOnce.Do(release)
	gomega.Eventually(first).Should(gomega.Receive(gomega.Equal([]string{"one", "two"})))
	gomega.Eventually(second).Should(gomega.Receive(gomega.Equal([]string{"one", "two"})))
	gomega.Expect(inner.callCount()).To(gomega.Equal(1))
	gomega.Expect(strings.Count(output.String(), "Listed ")).To(gomega.Equal(1))
	gomega.Expect(output.String()).NotTo(gomega.ContainSubstring("was reused"))
	gomega.Expect(collector.EventSummary()).To(gomega.ContainElement(opstats.EventSummary{Event: opstats.EventRegistryTagsSharedResult, Count: 2}))
})
