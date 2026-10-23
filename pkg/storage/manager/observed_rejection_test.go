package manager

import (
	"context"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/logging"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Observed stage rejection", func() {
	ginkgo.It("does not reuse a rejected stage from the local manifest cache after reconciliation", func(ctx ginkgo.SpecContext) {
		lookupCtx := logging.WithLogger(context.Context(ctx))
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
		previous := image.CommonManifestCache
		image.CommonManifestCache = image.NewManifestCache(ginkgo.GinkgoT().TempDir())
		ginkgo.DeferCleanup(func() { image.CommonManifestCache = previous })
		manager := newManifestStorageManager(lookupCtx)
		stageName := manager.StagesStorage.ConstructStageImageName(manager.ProjectName, missingStageDigest, 100)
		tag, err := name.NewTag(stageName, name.Insecure)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(remote.Write(tag, empty.Image, remote.WithContext(lookupCtx))).To(gomega.Succeed())
		cached, err := manager.GetStageDescSetByDigestFromStagesStorageCached(lookupCtx, "stage", missingStageDigest, 0, manager.StagesStorage)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(cached.Cardinality()).To(gomega.Equal(1))
		rejected, err := name.NewTag(stageName+"-rejected", name.Insecure)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(remote.Write(rejected, empty.Image, remote.WithContext(lookupCtx))).To(gomega.Succeed())
		fresh, err := manager.GetStageDescSetByDigest(lookupCtx, "stage", missingStageDigest, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fresh.IsEmpty()).To(gomega.BeTrue())
		cached, err = manager.GetStageDescSetByDigestFromStagesStorageCached(lookupCtx, "stage", missingStageDigest, 0, manager.StagesStorage)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(cached.IsEmpty()).To(gomega.BeTrue())
	})

	ginkgo.It("keeps a rejection observed while the first full listing is in flight", func(ctx ginkgo.SpecContext) {
		lookupCtx := logging.WithLogger(context.Context(ctx))
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
		previous := image.CommonManifestCache
		image.CommonManifestCache = image.NewManifestCache(ginkgo.GinkgoT().TempDir())
		ginkgo.DeferCleanup(func() { image.CommonManifestCache = previous })
		manager, blockListing, started, release := newRejectionSnapshotStorageManager(lookupCtx)
		stageName := manager.StagesStorage.ConstructStageImageName(manager.ProjectName, missingStageDigest, 100)
		tag, err := name.NewTag(stageName, name.Insecure)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(remote.Write(tag, empty.Image, remote.WithContext(lookupCtx))).To(gomega.Succeed())
		warm, err := manager.GetStageDescSetByDigest(lookupCtx, "stage", missingStageDigest, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(warm.Cardinality()).To(gomega.Equal(1))
		blockListing()
		older := make(chan image.StageDescSet, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			stages, err := manager.GetStageDescSetByDigestFromStagesStorageCached(lookupCtx, "stage", missingStageDigest, 0, manager.StagesStorage)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			older <- stages
		}()
		finished := false
		ginkgo.DeferCleanup(func() {
			release()
			if !finished {
				gomega.Eventually(older, 5*time.Second).Should(gomega.Receive())
			}
		})
		gomega.Eventually(started, 5*time.Second).Should(gomega.BeClosed())
		rejected, err := name.NewTag(stageName+"-rejected", name.Insecure)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(remote.Write(rejected, empty.Image, remote.WithContext(lookupCtx))).To(gomega.Succeed())
		fresh, err := manager.GetStageDescSetByDigest(lookupCtx, "stage", missingStageDigest, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fresh.IsEmpty()).To(gomega.BeTrue())
		release()
		var oldResult image.StageDescSet
		gomega.Eventually(older, 5*time.Second).Should(gomega.Receive(&oldResult))
		finished = true
		gomega.Expect(oldResult.IsEmpty()).To(gomega.BeTrue())
		cached, err := manager.GetStageDescSetByDigestFromStagesStorageCached(lookupCtx, "stage", missingStageDigest, 0, manager.StagesStorage)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(cached.IsEmpty()).To(gomega.BeTrue())
	})
})
