package docker_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Authoritative prefix tag listing", func() {
	prefix := strings.Repeat("a", 56)
	ginkgo.It("seeks past unrelated tags, preserves bare tags and markers across pages, and leaves the full cache intact", func(ctx ginkgo.SpecContext) {
		tags := []string{"000", prefix, "zzz"}
		expected := []string{prefix}
		for i := 0; i < 105; i++ {
			tag := fmt.Sprintf("%s-%013d", prefix, 1611836746968+int64(i))
			tags = append(tags, tag)
			expected = append(expected, tag)
		}
		marker := expected[99] + "-rejected"
		tags = append(tags, marker)
		expected = append(expected, marker)
		fixture := newPrefixRegistryFixture(tags, "")
		registry, err := NewDockerRegistry(ctx, fixture.repository(), DefaultImplementationName, DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		full, err := registry.Tags(ctx, fixture.repository(), WithCachedTags())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ranged, err := registry.TagsByPrefix(ctx, fixture.repository(), prefix)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(ranged).To(gomega.ConsistOf(expected))
		cached, err := registry.Tags(ctx, fixture.repository(), WithCachedTags())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(cached).To(gomega.ConsistOf(full))
		requests := fixture.requestLog()
		gomega.Expect(strings.Join(requests, "\n")).To(gomega.ContainSubstring("last=" + prefix + "&n=100"))
		fullListings := 0
		for _, request := range requests {
			if strings.Contains(request, "/tags/list?") && !strings.Contains(request, "last=") {
				fullListings++
			}
		}
		gomega.Expect(fullListings).To(gomega.Equal(1))
	})

	ginkgo.It("falls back when an ignored cursor looks ordered only after case folding", func(ctx ginkgo.SpecContext) {
		winner := prefix + "-100"
		tags := []string{winner}
		for i := 0; i < 101; i++ {
			tags = append(tags, fmt.Sprintf("Z%03d", i))
		}
		fixture := newPrefixRegistryFixture(tags, "ignored")
		registry, err := NewDockerRegistry(ctx, fixture.repository(), DefaultImplementationName, DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		result, err := registry.TagsByPrefix(ctx, fixture.repository(), prefix)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(result).To(gomega.Equal([]string{winner}))
	})

	ginkgo.DescribeTable("follows valid links and falls back on incompatible pagination", func(ctx ginkgo.SpecContext, mode string, fallback bool) {
		tags := []string{"000", prefix, prefix + "-1", prefix + "-2", prefix + "-2-rejected", prefix + "-3", "zzz"}
		if mode == "badlink" {
			for i := 0; i < 110; i++ {
				tags = append(tags, fmt.Sprintf("%s-%04d", prefix, i+100))
			}
		}
		fixture := newPrefixRegistryFixture(tags, mode)
		registry, err := NewDockerRegistry(ctx, fixture.repository(), DefaultImplementationName, DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		result, err := registry.TagsByPrefix(ctx, fixture.repository(), prefix)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		expected := []string{}
		for _, tag := range tags {
			if strings.HasPrefix(tag, prefix) {
				expected = append(expected, tag)
			}
		}
		gomega.Expect(result).To(gomega.ConsistOf(expected))
		full := false
		for _, request := range fixture.requestLog() {
			if strings.Contains(request, "/tags/list?") && !strings.Contains(request, "last=") {
				full = true
			}
		}
		gomega.Expect(full).To(gomega.Equal(fallback))
	}, ginkgo.Entry("short page with Link", "short", false), ginkgo.Entry("ignored cursor", "ignored", true),
		ginkgo.Entry("unsorted", "unsorted", true), ginkgo.Entry("unsupported cursor", "400", true),
		ginkgo.Entry("missing declaration", "undeclared", true), ginkgo.Entry("off-origin Link", "badlink", true))

	ginkgo.DescribeTable("propagates failures instead of returning empty inventory", func(ctx ginkgo.SpecContext, mode string) {
		fixture := newPrefixRegistryFixture([]string{prefix}, mode)
		registry, err := NewDockerRegistry(ctx, fixture.repository(), DefaultImplementationName, DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		inner := registry.(*DockerRegistryWithCache).Interface.(*timingDockerRegistry).Interface.(*defaultImplementation)
		inner.httpTransport = http.DefaultTransport
		inner.insecureHttpTransport = http.DefaultTransport
		_, err = registry.TagsByPrefix(ctx, fixture.repository(), prefix)
		gomega.Expect(err).To(gomega.HaveOccurred())
		for _, request := range fixture.requestLog() {
			if strings.Contains(request, "/tags/list?") {
				gomega.Expect(request).To(gomega.ContainSubstring("last="))
			}
		}
	}, ginkgo.Entry("unauthorized", "401"), ginkgo.Entry("throttled", "429"), ginkgo.Entry("server failure", "500"))

	ginkgo.DescribeTable("checks manifests without config or blob access", func(ctx ginkgo.SpecContext, mode string) {
		fixture := newPrefixRegistryFixture([]string{prefix}, mode)
		registry, err := NewDockerRegistry(ctx, fixture.repository(), DefaultImplementationName, DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		exists, err := registry.ManifestExists(ctx, fixture.repository()+":"+prefix)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(exists).To(gomega.BeTrue())
		absent, err := registry.ManifestExists(ctx, fixture.repository()+":absent")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(absent).To(gomega.BeFalse())
		log := strings.Join(fixture.requestLog(), "\n")
		gomega.Expect(log).To(gomega.ContainSubstring("HEAD /v2/repo/manifests/"))
		gomega.Expect(log).NotTo(gomega.ContainSubstring("/blobs/"))
		gomega.Expect(log).NotTo(gomega.ContainSubstring("/tags/list"))
	}, ginkgo.Entry("HEAD", ""), ginkgo.Entry("GET compatibility", "head405"))

	ginkgo.It("preserves bearer authentication and TLS configuration for bounded requests", func(ctx ginkgo.SpecContext) {
		fixture := newBearerRegistryFixtureWithTLS(true)
		ginkgo.DeferCleanup(fixture.server.Close)
		fixture.backend = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			if strings.Contains(r.URL.Path, "/manifests/") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			gomega.Expect(r.URL.Query().Get("last")).To(gomega.Equal(prefix))
			gomega.Expect(r.URL.Query().Get("n")).To(gomega.Equal("100"))
			w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
			gomega.Expect(json.NewEncoder(w).Encode(map[string]any{"name": "repo", "tags": []string{prefix + "-100", "zzz"}})).To(gomega.Succeed())
		})
		repository := strings.TrimPrefix(fixture.server.URL, "https://") + "/repo"
		registry, err := NewDockerRegistry(ctx, repository, DefaultImplementationName, DockerRegistryOptions{SkipTlsVerifyRegistry: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		tags, err := registry.TagsByPrefix(ctx, repository, prefix)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{prefix + "-100"}))
		gomega.Expect(fixture.exchanges.Load()).To(gomega.BeNumerically(">=", 1))
	})

	ginkgo.It("respects canceled contexts before listing", func(ctx ginkgo.SpecContext) {
		fixture := newPrefixRegistryFixture([]string{prefix}, "")
		registry, err := NewDockerRegistry(ctx, fixture.repository(), DefaultImplementationName, DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		_, err = registry.TagsByPrefix(canceled, fixture.repository(), prefix)
		gomega.Expect(err).To(gomega.MatchError(context.Canceled))
		gomega.Expect(fixture.requestLog()).To(gomega.BeEmpty())
	})
})
