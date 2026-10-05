package manager

import (
	"context"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/werf"
)

const missingStageDigest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab"

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
			cached, err := manager.GetStageDescSetByDigestFromStagesStorageCached(ctxWithLogger, "stage", missingStageDigest, 0, manager.StagesStorage)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(cached.Cardinality()).To(gomega.Equal(1), "a late older listing must not overwrite the fresh winner in the cache")
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
