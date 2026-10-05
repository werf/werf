package manager

import (
	"context"
	"fmt"
	"net/http"
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

const missingStageDigest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab"

var _ = ginkgo.Describe("GetStageDescSetByDigestFromStagesStorageCached", func() {
	ginkgo.It("should fetch tags once for repeated cache-only misses", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)
		manager, tagsListRequests := newTagsListStorageManager(http.StatusOK)

		for i := range 3 {
			stageDescSet, err := manager.GetStageDescSetByDigestFromStagesStorageCached(ctx, "stage", fmt.Sprintf("%056x", i+1), 0, manager.StagesStorage)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(stageDescSet.IsEmpty()).To(gomega.BeTrue())
		}

		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(1)))
	})

	ginkgo.It("should propagate storage errors", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)
		manager, _ := newTagsListStorageManager(http.StatusInternalServerError)

		_, err := manager.GetStageDescSetByDigestFromStagesStorageCached(ctx, "stage", missingStageDigest, 0, manager.StagesStorage)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("unable to get stages ids")))
	})
})

var _ = ginkgo.Describe("GetStageDescSetByDigestFromStagesStorageWithCache", func() {
	ginkgo.It("should refetch tags on a cached miss", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)
		manager, tagsListRequests := newTagsListStorageManager(http.StatusOK)

		_, err := manager.GetStageDescSetByDigestFromStagesStorageCached(ctx, "stage", missingStageDigest, 0, manager.StagesStorage)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(1)))

		stageDescSet, err := manager.GetStageDescSetByDigestFromStagesStorageWithCache(ctx, "stage", missingStageDigest, 0, manager.StagesStorage)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageDescSet.IsEmpty()).To(gomega.BeTrue())
		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(2)))
	})

	ginkgo.It("should propagate storage errors", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)
		manager, _ := newTagsListStorageManager(http.StatusInternalServerError)

		_, err := manager.GetStageDescSetByDigestFromStagesStorageWithCache(ctx, "stage", missingStageDigest, 0, manager.StagesStorage)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("unable to get stages ids")))
	})
})

var _ = ginkgo.Describe("GetStageDescSetByDigestWithRecentCache", func() {
	ginkgo.It("should reuse a recent tags listing on a cached miss", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)
		manager, tagsListRequests := newTagsListStorageManager(http.StatusOK)

		_, err := manager.GetStageDescSetByDigestFromStagesStorageCached(ctx, "stage", missingStageDigest, 0, manager.StagesStorage)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(1)))

		stageDescSet, err := manager.GetStageDescSetByDigestWithRecentCache(ctx, "stage", missingStageDigest, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageDescSet.IsEmpty()).To(gomega.BeTrue())
		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(1)))
	})

	ginkgo.It("should refetch tags on a cached miss when the listing is stale", func(ctx context.Context) {
		ginkgo.GinkgoT().Setenv(stagesTagListMaxAgeEnvVar, "1ns")

		ctx = logging.WithLogger(ctx)
		manager, tagsListRequests := newTagsListStorageManager(http.StatusOK)

		_, err := manager.GetStageDescSetByDigestFromStagesStorageCached(ctx, "stage", missingStageDigest, 0, manager.StagesStorage)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(1)))

		stageDescSet, err := manager.GetStageDescSetByDigestWithRecentCache(ctx, "stage", missingStageDigest, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageDescSet.IsEmpty()).To(gomega.BeTrue())
		gomega.Expect(tagsListRequests.Load()).To(gomega.Equal(int32(2)))
	})
})

var _ = ginkgo.DescribeTable("getStagesTagListMaxAge",
	func(value string, expected time.Duration) {
		ginkgo.GinkgoT().Setenv(stagesTagListMaxAgeEnvVar, value)
		gomega.Expect(getStagesTagListMaxAge()).To(gomega.Equal(expected))
	},
	ginkgo.Entry("empty value falls back to the default", "", stagesTagListMaxAgeDefault),
	ginkgo.Entry("zero disables the recent-listing window", "0", time.Duration(0)),
	ginkgo.Entry("invalid value falls back to the default", "bogus", stagesTagListMaxAgeDefault),
	ginkgo.Entry("negative value falls back to the default", "-5s", stagesTagListMaxAgeDefault),
	ginkgo.Entry("custom value is honored", "30s", 30*time.Second),
)

var _ = ginkgo.Describe("Authoritative stage reconciliation wiring", func() {
	ginkgo.It("sees the published winner without joining an earlier tag snapshot", func(ctx ginkgo.SpecContext) {
		lookupCtx, cancel := context.WithCancel(context.Background())
		ginkgo.DeferCleanup(cancel)
		ctxWithLogger := logging.WithLogger(lookupCtx)
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
		previousCache := image.CommonManifestCache
		image.CommonManifestCache = image.NewManifestCache(ginkgo.GinkgoT().TempDir())
		ginkgo.DeferCleanup(func() { image.CommonManifestCache = previousCache })
		manager, listingStarted, releaseListing := newBlockedTagsStorageManager(ctxWithLogger)
		stale := make(chan image.StageDescSet, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			stages, err := manager.GetStageDescSetByDigestFromStagesStorageCached(ctxWithLogger, "stage", missingStageDigest, 0, manager.StagesStorage)
			gomega.Expect(err).To(gomega.Succeed())
			stale <- stages
		}()
		ginkgo.DeferCleanup(func() {
			releaseListing()
			var stages image.StageDescSet
			gomega.Eventually(stale, 5*time.Second).Should(gomega.Receive(&stages))
			gomega.Expect(stages.IsEmpty()).To(gomega.BeTrue())
		})
		gomega.Eventually(listingStarted, 5*time.Second).Should(gomega.BeClosed())
		winnerName := manager.StagesStorage.ConstructStageImageName(manager.ProjectName, missingStageDigest, 100)
		winner, err := name.NewTag(winnerName, name.Insecure)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(remote.Write(winner, empty.Image, remote.WithContext(ctxWithLogger))).To(gomega.Succeed())
		fresh := make(chan image.StageDescSet, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			stages, err := manager.GetStageDescSetByDigest(ctxWithLogger, "stage", missingStageDigest, 0)
			gomega.Expect(err).To(gomega.Succeed())
			fresh <- stages
		}()
		var stages image.StageDescSet
		gomega.Eventually(fresh, 5*time.Second).Should(gomega.Receive(&stages))
		gomega.Expect(stages.Cardinality()).To(gomega.Equal(1))
		gomega.Expect(stages.ToSlice()[0].Info.Name).To(gomega.Equal(winnerName))
		gomega.Expect(stages.ToSlice()[0].StageID).To(gomega.Equal(image.NewStageID(missingStageDigest, 100)))
	})
})

var _ = ginkgo.Describe("Check mode stage lookup wiring", func() {
	ginkgo.It("refreshes a negative lookup without joining a listing started before it", func(ctx ginkgo.SpecContext) {
		lookupCtx, cancel := context.WithCancel(context.Background())
		ginkgo.DeferCleanup(cancel)
		ctxWithLogger := logging.WithLogger(lookupCtx)
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
		previousCache := image.CommonManifestCache
		image.CommonManifestCache = image.NewManifestCache(ginkgo.GinkgoT().TempDir())
		ginkgo.DeferCleanup(func() { image.CommonManifestCache = previousCache })
		manager, listingStarted, releaseListing := newTagsStorageManagerBlockingListing(ctxWithLogger, 2)

		warm, err := manager.GetStageDescSetByDigestFromStagesStorageCached(ctxWithLogger, "stage", missingStageDigest, 0, manager.StagesStorage)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(warm.IsEmpty()).To(gomega.BeTrue())

		stale := make(chan []image.StageID, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			stages, err := manager.StagesStorage.GetStagesIDsByDigest(ctxWithLogger, manager.ProjectName, missingStageDigest, 0)
			gomega.Expect(err).To(gomega.Succeed())
			stale <- stages
		}()
		ginkgo.DeferCleanup(func() {
			releaseListing()
			gomega.Eventually(stale, 5*time.Second).Should(gomega.Receive(gomega.BeEmpty()))
		})
		gomega.Eventually(listingStarted, 5*time.Second).Should(gomega.BeClosed())

		winnerName := manager.StagesStorage.ConstructStageImageName(manager.ProjectName, missingStageDigest, 100)
		winner, err := name.NewTag(winnerName, name.Insecure)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(remote.Write(winner, empty.Image, remote.WithContext(ctxWithLogger))).To(gomega.Succeed())

		checked := make(chan image.StageDescSet, 1)
		go func() {
			defer ginkgo.GinkgoRecover()
			stages, err := manager.GetStageDescSetByDigestWithCache(ctxWithLogger, "stage", missingStageDigest, 0)
			gomega.Expect(err).To(gomega.Succeed())
			checked <- stages
		}()

		var stages image.StageDescSet
		gomega.Eventually(checked, 5*time.Second).Should(gomega.Receive(&stages), "the check must list the repo afresh instead of waiting for the listing started before the publication")
		gomega.Expect(stages.Cardinality()).To(gomega.Equal(1))
		gomega.Expect(stages.ToSlice()[0].Info.Name).To(gomega.Equal(winnerName))
	})
})
