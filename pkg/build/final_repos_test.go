package build

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	buildImage "github.com/werf/werf/v3/pkg/build/image"
	imagePkg "github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Build with multiple final repos", func() {
	ginkgo.DescribeTable("publishes images and custom tags into every destination",
		func(ctx ginkgo.SpecContext, multiplatform, promote bool) {
			gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
			phase, storageManager := newFinalRepoBuildPhase(ctx)
			phase.PromoteBuiltImages = promote
			first := publishFinalRepoTestImage(ctx, storageManager, "amd64", *imagePkg.NewStageID(fmt.Sprintf("%056x", 1), 1611836746968))
			var published *imagePkg.StageDesc
			var publish func() error
			var publishTags func() error
			if multiplatform {
				second := publishFinalRepoTestImage(ctx, storageManager, "arm64", *imagePkg.NewStageID(fmt.Sprintf("%056x", 2), 1611836746968))
				img := buildImage.NewMultiplatformImage("app", []*buildImage.Image{first, second}, 0, 1)
				gomega.Expect(phase.publishMultiplatformImageMetadata(ctx, img.Name, img)).To(gomega.Succeed())
				publish = func() error {
					err := phase.publishMultiplatformFinalImage(ctx, img.Name, img)
					published = img.GetFinalStageDesc()
					return err
				}
				publishTags = func() error { return phase.publishMultiplatformImageCustomTags(ctx, img.Name, img) }
			} else {
				publish = func() error {
					err := phase.publishFinalImage(ctx, first.Name, first)
					published = first.GetContentTagDesc()
					return err
				}
				publishTags = func() error { return phase.publishImageMetadata(ctx, first.Name, first) }
			}

			finalStorages := storageManager.FinalStagesStorageList
			storageManager.FinalStagesStorageList = finalStorages[:1]
			gomega.Expect(publish()).To(gomega.Succeed())
			stageID := *published.StageID
			storageManager.FinalStagesStorageList = finalStorages
			phase.ShouldBeBuiltMode = true
			gomega.Expect(publish()).To(gomega.MatchError(gomega.ContainSubstring(finalStorages[1].Address())))
			_, err := finalStorages[1].GetStageDesc(ctx, storageManager.ProjectName, stageID)
			gomega.Expect(storage.IsErrStageNotFound(err)).To(gomega.BeTrue())
			phase.ShouldBeBuiltMode = false
			gomega.Expect(publish()).To(gomega.Succeed())
			gomega.Expect(publishTags()).To(gomega.Succeed())
			gomega.Expect(published.Info.Repository).To(gomega.Equal(storageManager.FinalStagesStorageList[0].Address()))
			for _, repo := range storageManager.FinalStagesStorageList {
				desc, err := repo.GetStageDesc(ctx, storageManager.ProjectName, *published.StageID)
				gomega.Expect(err).To(gomega.Succeed())
				gomega.Expect(desc.Info.GetDigest()).To(gomega.Equal(published.Info.GetDigest()))
				gomega.Expect(repo.CheckStageCustomTag(ctx, desc, "release")).To(gomega.Succeed())
				if multiplatform {
					ref, err := name.ParseReference(desc.Info.Name)
					gomega.Expect(err).To(gomega.Succeed())
					index, err := remote.Index(ref, remote.WithContext(ctx))
					gomega.Expect(err).To(gomega.Succeed())
					manifest, err := index.IndexManifest()
					gomega.Expect(err).To(gomega.Succeed())
					gomega.Expect(manifest.Manifests).To(gomega.HaveLen(2))
					for _, child := range manifest.Manifests {
						childImage, err := index.Image(child.Digest)
						gomega.Expect(err).To(gomega.Succeed())
						_, err = childImage.ConfigFile()
						gomega.Expect(err).To(gomega.Succeed())
					}
				}
			}

			phase.ShouldBeBuiltMode = true
			gomega.Expect(publish()).To(gomega.Succeed())
			gomega.Expect(publishTags()).To(gomega.Succeed())
			wrongTag, err := name.NewTag(storageManager.FinalStagesStorageList[1].Address() + ":release")
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(remote.Write(wrongTag, empty.Image, remote.WithContext(ctx))).To(gomega.Succeed())
			gomega.Expect(publishTags()).To(gomega.MatchError(gomega.ContainSubstring("must be the same")))
		},
		ginkgo.Entry("single-platform image", false, false),
		ginkgo.Entry("multiplatform image including child manifests", true, false),
		ginkgo.Entry("promote single-platform image", false, true),
		ginkgo.Entry("promote multiplatform image including child manifests", true, true),
	)

	ginkgo.DescribeTable("bounds copies across images and final repos", func(ctx ginkgo.SpecContext, limit int64, expectedActive int32) {
		release := make(chan struct{})
		releaseOnce := sync.OnceFunc(func() { close(release) })
		phase, storageManager, expectedCopies := newFinalPublicationTestPhase(3, 2, limit, release)
		result := make(chan error, 1)
		finished := make(chan struct{})
		ginkgo.DeferCleanup(func() {
			releaseOnce()
			gomega.Eventually(finished, 5*time.Second).Should(gomega.BeClosed())
		})
		go func() {
			defer ginkgo.GinkgoRecover()
			defer close(finished)
			result <- phase.AfterImages(ctx)
		}()
		gomega.Eventually(storageManager.maximum.Load, 5*time.Second).Should(gomega.BeNumerically(">=", expectedActive))
		gomega.Consistently(storageManager.maximum.Load, 500*time.Millisecond).Should(gomega.BeNumerically("<=", expectedActive))
		releaseOnce()
		gomega.Eventually(result, 5*time.Second).Should(gomega.Receive(gomega.BeNil()))
		gomega.Expect(storageManager.copied).To(gomega.ConsistOf(expectedCopies))
		gomega.Expect(storageManager.maximum.Load()).To(gomega.Equal(expectedActive))
	},
		ginkgo.Entry("fills five slots with two destinations", int64(5), int32(5)),
		ginkgo.Entry("respects a limit of two", int64(2), int32(2)),
		ginkgo.Entry("serializes with a limit of one", int64(1), int32(1)),
		ginkgo.Entry("allows unlimited copies with zero", int64(0), int32(6)),
	)

	ginkgo.DescribeTable("uses free publication workers while one copy is slow", func(ctx ginkgo.SpecContext, imageCount, repoCount int) {
		release := make(chan struct{})
		releaseOnce := sync.OnceFunc(func() { close(release) })
		phase, storageManager, expectedCopies := newFinalPublicationTestPhase(imageCount, repoCount, 5, release)
		storageManager.blockedCopies = map[string]bool{expectedCopies[0]: true}
		result := make(chan error, 1)
		finished := make(chan struct{})
		ginkgo.DeferCleanup(func() {
			releaseOnce()
			gomega.Eventually(finished, 5*time.Second).Should(gomega.BeClosed())
		})
		go func() {
			defer ginkgo.GinkgoRecover()
			defer close(finished)
			result <- phase.AfterImages(ctx)
		}()
		gomega.Eventually(storageManager.started.Load, 5*time.Second).Should(gomega.Equal(int32(len(expectedCopies))))
		releaseOnce()
		gomega.Eventually(result, 5*time.Second).Should(gomega.Receive(gomega.BeNil()))
		gomega.Expect(storageManager.copied).To(gomega.ConsistOf(expectedCopies))
	},
		ginkgo.Entry("images", 6, 2),
		ginkgo.Entry("destinations", 1, 6),
	)

	ginkgo.DescribeTable("stops pending publication on cancellation or copy failure", func(ctx ginkgo.SpecContext, fail bool) {
		publicationCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		release := make(chan struct{})
		releaseOnce := sync.OnceFunc(func() { close(release) })
		phase, storageManager, _ := newFinalPublicationTestPhase(3, 2, 2, release)
		var expectedErr error = context.Canceled
		if fail {
			expectedErr = errors.New("copy failed")
			storageManager.copyErr = expectedErr
		}
		result := make(chan error, 1)
		finished := make(chan struct{})
		ginkgo.DeferCleanup(func() {
			cancel()
			releaseOnce()
			gomega.Eventually(finished, 5*time.Second).Should(gomega.BeClosed())
		})
		go func() {
			defer ginkgo.GinkgoRecover()
			defer close(finished)
			result <- phase.AfterImages(publicationCtx)
		}()
		gomega.Eventually(storageManager.active.Load, 5*time.Second).Should(gomega.Equal(int32(2)))
		if fail {
			releaseOnce()
		} else {
			cancel()
		}
		gomega.Eventually(result, 5*time.Second).Should(gomega.Receive(gomega.MatchError(gomega.ContainSubstring(expectedErr.Error()))))
		gomega.Expect(storageManager.active.Load()).To(gomega.BeZero())
		gomega.Expect(storageManager.copied).To(gomega.BeEmpty())
	},
		ginkgo.Entry("cancellation", false),
		ginkgo.Entry("copy failure", true),
	)
})
