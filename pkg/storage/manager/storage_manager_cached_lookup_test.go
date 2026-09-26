package manager

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/logging"
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
