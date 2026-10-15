package storage

import (
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
)

var _ = ginkgo.Describe("Digest-isolated stage discovery", func() {
	digest := strings.Repeat("a", 56)
	other := strings.Repeat("b", 56)
	ginkgo.DescribeTable("matches digest and rejection identity on both listing paths", func(ctx ginkgo.SpecContext, fresh bool, tags []string, parent int64, want []int64) {
		registry := &digestListingRegistry{tags: tags}
		storage := &RepoStagesStorage{RepoAddress: "example.test/repo", DockerRegistry: registry, skipMetaCheck: true}
		options := []Option{WithCache()}
		if fresh {
			options = append(options, WithFreshListing())
		}
		ids, err := storage.GetStagesIDsByDigest(ctx, "project", digest, parent, options...)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		expected := make([]image.StageID, 0, len(want))
		for _, ts := range want {
			expected = append(expected, *image.NewStageID(digest, ts))
		}
		gomega.Expect(ids).To(gomega.ConsistOf(expected))
		if fresh {
			gomega.Expect(registry.prefixCalls).To(gomega.Equal(1))
			gomega.Expect(registry.fullCalls).To(gomega.BeZero())
		} else {
			gomega.Expect(registry.fullCalls).To(gomega.Equal(1))
			gomega.Expect(registry.prefixCalls).To(gomega.BeZero())
		}
	}, ginkgo.Entry("cached positive control", false, []string{digest + "-100"}, int64(100), []int64{100}),
		ginkgo.Entry("fresh positive control", true, []string{digest + "-100"}, int64(100), []int64{100}),
		ginkgo.Entry("cached foreign rejection", false, []string{digest + "-100", other + "-100-rejected"}, int64(0), []int64{100}),
		ginkgo.Entry("fresh foreign rejection", true, []string{digest + "-100", other + "-100-rejected"}, int64(0), []int64{100}),
		ginkgo.Entry("own rejection", true, []string{digest + "-100", digest + "-100-rejected"}, int64(0), []int64{}),
		ginkgo.Entry("bare digest", true, []string{digest, digest + "-100"}, int64(0), []int64{0, 100}),
		ginkgo.Entry("bare digest rejected", true, []string{digest, digest + "-0-rejected"}, int64(0), []int64{}),
		ginkgo.Entry("parent boundary", true, []string{digest + "-99", digest + "-100", digest + "-101"}, int64(100), []int64{100, 101}),
		ginkgo.Entry("prefix collision", false, []string{digest + "b-100", digest + "-101"}, int64(0), []int64{101}))
})
