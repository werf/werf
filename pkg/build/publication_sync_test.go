package build

import (
	"errors"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	imagePkg "github.com/werf/werf/v3/pkg/image"
)

var _ = ginkgo.Describe("Distributed stage publication", func() {
	ginkgo.BeforeEach(func() { ginkgo.GinkgoT().Setenv("WERF_DISABLE_PUBLISH_TAG_CACHE_SYNC", "") })
	ginkgo.DescribeTable("adopts the winner after waiting for another publisher", func(ctx ginkgo.SpecContext, anchor bool, secondParentTs int64) {
		srv, attempts := newPublicationLockServer()
		primary := &publicationStorage{}
		firstManager := &publicationStorageManager{primary: primary, lookupStarted: make(chan struct{}), continueLookup: make(chan struct{})}
		secondManager := &publicationStorageManager{primary: primary}
		first, firstImage, firstStage := newPublicationPhase(ctx, firstManager, srv.URL, anchor, 10)
		second, secondImage, secondStage := newPublicationPhase(ctx, secondManager, srv.URL, anchor, secondParentTs)
		firstDone, secondDone := make(chan error, 1), make(chan error, 1)
		go func() { firstDone <- errorOf(first.atomicBuildStageImage(ctx, firstImage, firstStage)) }()
		gomega.Eventually(firstManager.lookupStarted, 5*time.Second).Should(gomega.BeClosed())
		gomega.Eventually(attempts, 5*time.Second).Should(gomega.Receive())
		go func() { secondDone <- errorOf(second.atomicBuildStageImage(ctx, secondImage, secondStage)) }()
		gomega.Eventually(attempts, 5*time.Second).Should(gomega.Receive())
		blockedLookups := secondManager.lookups.Load()
		close(firstManager.continueLookup)
		gomega.Eventually(firstDone, 10*time.Second).Should(gomega.Receive(gomega.Succeed()))
		gomega.Eventually(secondDone, 10*time.Second).Should(gomega.Receive(gomega.Succeed()))
		gomega.Expect(blockedLookups).To(gomega.BeZero(), "the waiter must not inspect registry before acquiring the distributed lock")
		gomega.Expect(primary.writes).To(gomega.Equal(1))
		gomega.Expect(secondStage.GetStageImage().Image.GetStageDesc()).To(gomega.Equal(firstStage.GetStageImage().Image.GetStageDesc()))
		gomega.Expect(secondStage.GetContentDigest()).To(gomega.Equal("winner-content"))
		gomega.Expect(secondManager.lookups.Load()).To(gomega.Equal(int32(1)))
		if anchor {
			gomega.Expect(secondManager.parentTs).To(gomega.BeZero())
		} else {
			gomega.Expect(secondManager.parentTs).To(gomega.Equal(secondParentTs))
		}
	},
		ginkgo.Entry("ordinary stage", false, int64(20)),
		ginkgo.Entry("anchor with a newer, different parent chain", true, int64(200)),
	)

	ginkgo.DescribeTable("secondary-copy waiter adopts the primary publisher", func(ctx ginkgo.SpecContext, anchor bool, secondParentTs int64) {
		srv, attempts := newPublicationLockServer()
		primary := &publicationStorage{}
		firstManager := &publicationStorageManager{primary: primary, lookupStarted: make(chan struct{}), continueLookup: make(chan struct{})}
		secondManager := &publicationStorageManager{primary: primary, secondary: &publicationStorage{}, secondaryDesc: &imagePkg.StageDesc{StageID: imagePkg.NewStageID("shared-digest", 300), Info: &imagePkg.Info{Name: "repo:alternate", Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "alternate-content"}}}}
		first, firstImage, firstStage := newPublicationPhase(ctx, firstManager, srv.URL, anchor, 10)
		second, secondImage, secondStage := newPublicationPhase(ctx, secondManager, srv.URL, anchor, secondParentTs)
		firstDone, secondDone := make(chan error, 1), make(chan error, 1)
		go func() { firstDone <- errorOf(first.atomicBuildStageImage(ctx, firstImage, firstStage)) }()
		gomega.Eventually(firstManager.lookupStarted, 5*time.Second).Should(gomega.BeClosed())
		gomega.Eventually(attempts, 5*time.Second).Should(gomega.Receive())
		go func() {
			found, err := second.findAndFetchStageFromSecondaryStagesStorage(ctx, secondImage, secondStage)
			if err == nil && !found {
				err = errors.New("secondary stage was not selected")
			}
			secondDone <- err
		}()
		gomega.Eventually(attempts, 5*time.Second).Should(gomega.Receive())
		blockedLookups := secondManager.lookups.Load()
		close(firstManager.continueLookup)
		gomega.Eventually(firstDone, 10*time.Second).Should(gomega.Receive(gomega.Succeed()))
		gomega.Eventually(secondDone, 10*time.Second).Should(gomega.Receive(gomega.Succeed()))
		gomega.Expect(blockedLookups).To(gomega.BeZero())
		gomega.Expect(primary.writes).To(gomega.Equal(1))
		gomega.Expect(secondManager.copies.Load()).To(gomega.BeZero())
		gomega.Expect(secondStage.GetStageImage().Image.GetStageDesc()).To(gomega.Equal(firstStage.GetStageImage().Image.GetStageDesc()))
		gomega.Expect(secondStage.GetContentDigest()).To(gomega.Equal("winner-content"))
	},
		ginkgo.Entry("ordinary stage", false, int64(20)),
		ginkgo.Entry("anchor from a different parent chain", true, int64(200)),
	)

	ginkgo.DescribeTable("does not block unrelated publication keys", func(ctx ginkgo.SpecContext, project, digest string) {
		srv, _ := newPublicationLockServer()
		first, _, _ := newPublicationPhase(ctx, &publicationStorageManager{}, srv.URL, false, 10)
		second, _, _ := newPublicationPhase(ctx, &publicationStorageManager{}, srv.URL, false, 10)
		held, err := first.Conveyor.StorageLockManager.LockStage(ctx, "publication-project", "shared-digest")
		gomega.Expect(err).To(gomega.Succeed())
		ginkgo.DeferCleanup(func() { gomega.Expect(first.Conveyor.StorageLockManager.Unlock(ctx, held)).To(gomega.Succeed()) })
		done := make(chan error, 1)
		go func() {
			handle, err := second.Conveyor.StorageLockManager.LockStage(ctx, project, digest)
			if err == nil {
				err = second.Conveyor.StorageLockManager.Unlock(ctx, handle)
			}
			done <- err
		}()
		gomega.Eventually(done, 5*time.Second).Should(gomega.Receive(gomega.Succeed()))
	},
		ginkgo.Entry("different project", "other-project", "shared-digest"),
		ginkgo.Entry("different stage digest", "publication-project", "other-digest"),
	)

	ginkgo.DescribeTable("releases publication lock after errors", func(ctx ginkgo.SpecContext, failure string) {
		srv, attempts := newPublicationLockServer()
		primary := &publicationStorage{}
		storageManager := &publicationStorageManager{primary: primary}
		sentinel := errors.New("publication failure")
		switch failure {
		case "lookup":
			storageManager.lookupErr = sentinel
		case "selection":
			storageManager.selectionErr = sentinel
		case "push":
			primary.writeErr = sentinel
		case "description":
			primary.descriptionErr = sentinel
		case "cache":
			storageManager.cacheErr = sentinel
		}
		phase, img, stg := newPublicationPhase(ctx, storageManager, srv.URL, true, 200)
		_, err := phase.atomicBuildStageImage(ctx, img, stg)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("publication failure")))
		gomega.Eventually(attempts, time.Second).Should(gomega.Receive())
		other, _, _ := newPublicationPhase(ctx, &publicationStorageManager{primary: primary}, srv.URL, true, 200)
		acquired := make(chan error, 1)
		go func() {
			handle, err := other.Conveyor.StorageLockManager.LockStage(ctx, "publication-project", "shared-digest")
			if err == nil {
				err = other.Conveyor.StorageLockManager.Unlock(ctx, handle)
			}
			acquired <- err
		}()
		gomega.Eventually(acquired, 5*time.Second).Should(gomega.Receive(gomega.Succeed()))
	},
		ginkgo.Entry("registry lookup fails", "lookup"),
		ginkgo.Entry("stage selection fails", "selection"),
		ginkgo.Entry("image push fails", "push"),
		ginkgo.Entry("description read fails after push", "description"),
		ginkgo.Entry("cache copy fails after publication", "cache"),
	)
})
