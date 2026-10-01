package storage

import (
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/image"
)

var _ = ginkgo.Describe("synchronization configuration marker", func() {
	const fingerprint = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	ginkgo.It("persists per-project configuration in the primary registry and excludes it from metadata migration", func(ctx ginkgo.SpecContext) {
		store, _ := newTagCacheRepoStagesStorage(ctx, &pushStageBackendStub{})
		value, found, err := store.GetSynchronizationMarker(ctx, "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(found).To(gomega.BeFalse())
		gomega.Expect(value).To(gomega.BeEmpty())
		gomega.Expect(store.PutSynchronizationMarker(ctx, "project", fingerprint)).To(gomega.Succeed())
		value, found, err = store.GetSynchronizationMarker(ctx, "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(found).To(gomega.BeTrue())
		gomega.Expect(value).To(gomega.Equal(fingerprint))
		_, found, err = store.GetSynchronizationMarker(ctx, "another-project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(found).To(gomega.BeFalse())
		records, err := store.collectMetadataRecords(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(records).To(gomega.BeEmpty())
		gomega.Expect(store.PutSynchronizationMarker(ctx, "project", strings.Repeat("b", 64))).To(gomega.MatchError(gomega.ContainSubstring("differs")))
		value, _, err = store.GetSynchronizationMarker(ctx, "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(value).To(gomega.Equal(fingerprint))
	})
	ginkgo.It("detects a competing marker visible after publishing", func(ctx ginkgo.SpecContext) {
		store, _ := newTagCacheRepoStagesStorage(ctx, &pushStageBackendStub{})
		store.DockerRegistry = &synchronizationCompetingRegistry{Interface: store.DockerRegistry}
		gomega.Expect(store.PutSynchronizationMarker(ctx, "project", fingerprint)).To(gomega.MatchError(gomega.ContainSubstring("changed while configuring")))
	})

	ginkgo.DescribeTable("rejects occupied or malformed reserved tags without overwriting them", func(ctx ginkgo.SpecContext, labels map[string]string) {
		store, _ := newTagCacheRepoStagesStorage(ctx, &pushStageBackendStub{})
		ref := synchronizationMarkerName(store.RepoAddress, "project")
		gomega.Expect(store.DockerRegistry.PushImage(ctx, ref, &docker_registry.PushImageOptions{Labels: labels})).To(gomega.Succeed())
		before, err := store.DockerRegistry.GetRepoImage(ctx, ref)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(store.PutSynchronizationMarker(ctx, "project", fingerprint)).NotTo(gomega.Succeed())
		after, err := store.DockerRegistry.GetRepoImage(ctx, ref)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(after.Labels).To(gomega.Equal(before.Labels))
	},
		ginkgo.Entry("missing fingerprint", map[string]string{image.WerfLabel: "project"}),
		ginkgo.Entry("invalid fingerprint", map[string]string{image.WerfLabel: "project", synchronizationMarkerFingerprintLabel: "invalid"}),
		ginkgo.Entry("wrong project", map[string]string{image.WerfLabel: "another", synchronizationMarkerFingerprintLabel: fingerprint}),
		ginkgo.Entry("built image alias with valid marker labels", map[string]string{image.WerfLabel: "project", synchronizationMarkerFingerprintLabel: fingerprint, image.WerfStageContentDigestLabel: "content"}),
	)
})
