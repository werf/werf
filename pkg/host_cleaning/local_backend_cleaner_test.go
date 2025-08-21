package host_cleaning

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/lockgate"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/filter"
	"github.com/werf/werf/v2/pkg/container_backend/info"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/volumeutils"
	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("LocalBackendCleaner", func() {
	t := GinkgoT()

	var cleaner *LocalBackendCleaner
	var backend *mock.MockContainerBackend
	var locker *mock.MockLocker

	var stubs *gostub.Stubs

	BeforeEach(func() {
		backend = mock.NewMockContainerBackend(gomock.NewController(t))
		locker = mock.NewMockLocker(gomock.NewController(t))
		var err error
		cleaner, err = NewLocalBackendCleaner(backend, locker)
		Expect(errors.Is(err, ErrUnsupportedContainerBackend)).To(BeTrue())
		Expect(cleaner).NotTo(BeNil())
		stubs = gostub.New()
	})
	AfterEach(func() {
		stubs.Reset()
	})

	DescribeTable("normalizeReference",
		func(backendType containerBackendType, inputRef, expectedRef string) {
			cleaner.backendType = backendType

			outputRef, err := cleaner.normalizeReference(inputRef)
			Expect(err).To(BeNil())
			Expect(outputRef).To(Equal(expectedRef))
		},
		Entry("should work for docker backend type",
			containerBackendDocker,
			"werf-guide-app:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1739265936011",
			"werf-guide-app:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1739265936011"),

		Entry("should work for test backend type",
			containerBackendTest,
			"werf-guide-app:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1739265936011",
			"werf-guide-app:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1739265936011"),

		Entry("should work for buildah backend type",
			containerBackendBuildah,
			"localhost/werf-guide-app:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1739265965865",
			"werf-guide-app:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1739265965865"),
	)

	Describe("backendStoragePath", func() {
		It("should return err if backend.Info() returns err", func(ctx SpecContext) {
			err0 := errors.New("some err")
			backend.EXPECT().Info(ctx).Return(info.Info{}, err0)

			_, err := cleaner.backendStoragePath(ctx, "")

			Expect(errors.Is(err, err0)).To(BeTrue())
		})
	})

	Describe("ShouldRunAutoGC", func() {
		It("should return true if cleanup needed", func(ctx SpecContext) {
			result, err := cleaner.ShouldRunAutoGC(ctx, RunAutoGCOptions{
				AllowedStorageVolumeUsagePercentage: 0,
				StoragePath:                         t.TempDir(),
			})
			Expect(err).To(Succeed())
			Expect(result).To(BeTrue())
		})
		It("should return false if cleanup not needed", func(ctx SpecContext) {
			result, err := cleaner.ShouldRunAutoGC(ctx, RunAutoGCOptions{
				AllowedStorageVolumeUsagePercentage: 1000,
				StoragePath:                         t.TempDir(),
			})
			Expect(err).To(Succeed())
			Expect(result).To(BeFalse())
		})
	})

	Describe("pruneImages", func() {
		var filters filter.FilterList
		BeforeEach(func() {
			filters = filter.FilterList{
				filter.DanglingTrue,
				filter.NewFilter("label", image.WerfLabel),
				filter.NewFilter("until", "15m"),
			}
		})
		It("should return err=nil and full report if opts.DryRun=true calling backend.Images() to find dandling images", func(ctx SpecContext) {
			list := image.ImagesList{
				{ID: "one"},
			}
			backend.EXPECT().Images(ctx, buildImagesOptions(filters.ToPairs()...)).Return(list, nil)

			report, err := cleaner.pruneImages(ctx, RunGCOptions{
				DryRun: true,
			})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(mapImageListToCleanupReport(list)))
		})
		It("should return err=some_err and empty report if opts.DryRun=false calling backend.PruneImages() which returns pruneErr=err", func(ctx SpecContext) {
			err0 := errors.New("some_err")
			backend.EXPECT().PruneImages(ctx, prune.Options{Filters: filters}).Return(prune.Report{}, err0)

			report, err := cleaner.pruneImages(ctx, RunGCOptions{})
			Expect(err).To(Equal(err0))
			Expect(report).To(Equal(newCleanupReport(0)))
		})
		It("should return err=nil and empty report if opts.DryRun=false calling backend.PruneImages() which returns pruneErr=ErrImageUsedByContainer", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			backend.EXPECT().PruneImages(ctx, prune.Options{Filters: filters}).Return(prune.Report{}, container_backend.ErrImageUsedByContainer)

			report, err := cleaner.pruneImages(ctx, RunGCOptions{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport(0)))
		})
		It("should return err=nil and empty report if opts.DryRun=false calling backend.PruneImages() which returns pruneErr=ErrPruneIsAlreadyRunning", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			backend.EXPECT().PruneImages(ctx, prune.Options{Filters: filters}).Return(prune.Report{}, container_backend.ErrPruneIsAlreadyRunning)

			report, err := cleaner.pruneImages(ctx, RunGCOptions{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport(0)))
		})
		It("should return err=nil and full report if opts.DryRun=false calling backend.PruneImages() which returns pruneErr=nil", func(ctx SpecContext) {
			pruneReport := prune.Report{
				ItemsDeleted: []string{"one"},
			}
			backend.EXPECT().PruneImages(ctx, prune.Options{Filters: filters}).Return(pruneReport, nil)

			report, err := cleaner.pruneImages(ctx, RunGCOptions{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(mapPruneReportToCleanupReport(pruneReport)))
		})
	})

	Describe("pruneVolumes", func() {
		It("should return err=errOptionDryRunNotSupported and empty report if opts.DryRun=true", func(ctx SpecContext) {
			report, err := cleaner.pruneVolumes(ctx, RunGCOptions{
				DryRun: true,
			})
			Expect(errors.Is(err, errOptionDryRunNotSupported)).To(BeTrue())
			Expect(report).To(Equal(newCleanupReport(0)))
		})
		It("should return err=nil and empty report if opts.DryRun=false calling backend.PruneVolumes() which returns returns pruneErr=ErrPruneIsAlreadyRunning", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			backend.EXPECT().PruneVolumes(ctx, prune.Options{}).Return(prune.Report{}, container_backend.ErrPruneIsAlreadyRunning)

			report, err := cleaner.pruneVolumes(ctx, RunGCOptions{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport(0)))
		})
		It("should return err=some_err and empty report if opts.DryRun=false calling backend.PruneVolumes() which returns returns pruneErr=err", func(ctx SpecContext) {
			err0 := errors.New("some_err")
			backend.EXPECT().PruneVolumes(ctx, prune.Options{}).Return(prune.Report{}, err0)

			report, err := cleaner.pruneVolumes(ctx, RunGCOptions{})
			Expect(err).To(Equal(err0))
			Expect(report).To(Equal(newCleanupReport(0)))
		})
		It("should return err=nil and full report if opts.DryRun=false calling backend.PruneVolumes() which returns pruneErr=nil", func(ctx SpecContext) {
			pruneReport := prune.Report{
				ItemsDeleted: []string{"one"},
			}
			backend.EXPECT().PruneVolumes(ctx, prune.Options{}).Return(pruneReport, nil)

			report, err := cleaner.pruneVolumes(ctx, RunGCOptions{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(mapPruneReportToCleanupReport(pruneReport)))
		})
	})

	DescribeTable("cleanupWerfContainers",
		func(ctx context.Context, runOpts RunGCOptions, container image.Container, isAcquired bool, rmErr error, expectedReport cleanupReport) {
			ctx = logging.WithLogger(ctx)

			backend.EXPECT().Containers(ctx, buildContainersOptions(
				image.ContainerFilter{Name: image.StageContainerNamePrefix},
				image.ContainerFilter{Name: image.ImportServerContainerNamePrefix},
			)).Return(image.ContainerList{container}, nil)

			locker.EXPECT().Acquire(container_backend.ContainerLockName(container.Names[0][1:]), lockgate.AcquireOptions{NonBlocking: true}).Return(isAcquired, lockgate.LockHandle{}, nil)

			if isAcquired {
				locker.EXPECT().Release(lockgate.LockHandle{}).Return(nil)
			}
			if isAcquired && !runOpts.DryRun {
				backend.EXPECT().Rm(ctx, container.ID, container_backend.RmOpts{Force: false}).Return(rmErr)
			}

			stubs.StubFunc(&cleaner.volumeutilsGetVolumeUsageByPath, volumeutils.VolumeUsage{}, nil)

			report, err := cleaner.cleanupWerfContainers(ctx, runOpts, volumeutils.VolumeUsage{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(expectedReport))
		},
		Entry(
			"should not return err if backend.Rm() returns 'container is paused' error",
			RunGCOptions{},
			image.Container{
				ID:    "some-id",
				Names: []string{fmt.Sprintf("/%s", image.StageContainerNamePrefix)},
			},
			true,
			container_backend.ErrCannotRemovePausedContainer,
			newCleanupReport(0),
		),
		Entry(
			"should not return err if backend.Rm() returns 'container is running' error",
			RunGCOptions{},
			image.Container{
				ID:    "some-id",
				Names: []string{fmt.Sprintf("/%s", image.StageContainerNamePrefix)},
			},
			true,
			container_backend.ErrCannotRemoveRunningContainer,
			newCleanupReport(0),
		),
		Entry(
			"should not call backend.Rm() if lock was not acquired",
			RunGCOptions{},
			image.Container{
				ID:    "some-id",
				Names: []string{fmt.Sprintf("/%s", image.StageContainerNamePrefix)},
			},
			false,
			nil,
			newCleanupReport(0),
		),
		Entry(
			"should return full report in dry run mode if lock was acquired",
			RunGCOptions{
				DryRun: true,
			},
			image.Container{
				ID:    "some-id",
				Names: []string{fmt.Sprintf("/%s", image.StageContainerNamePrefix)},
			},
			true,
			nil,
			cleanupReport{
				ItemsDeleted:   []string{"some-id"},
				SpaceReclaimed: 0,
			},
		),
		Entry(
			"should return empty report in dry run mode if lock was not acquired",
			RunGCOptions{
				DryRun: true,
			},
			image.Container{
				ID:    "some-id",
				Names: []string{fmt.Sprintf("/%s", image.StageContainerNamePrefix)},
			},
			false,
			nil,
			newCleanupReport(0),
		),
	)

	Describe("werfImages", func() {
		It("should return images as merged and sorted result of several backend calls", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			expectedImages := []image.Summary{
				{ID: "one", Created: time.Unix(300, 0)},
				{ID: "two", Created: time.Unix(200, 0)},
				{ID: "three", Created: time.Unix(100, 0)},
			}

			backend.EXPECT().Images(ctx, buildImagesOptions(
				filter.DanglingFalse.ToPair(),
				util.NewPair("label", image.WerfLabel),
			)).Return(image.ImagesList{expectedImages[0]}, nil)

			backend.EXPECT().Images(ctx, buildImagesOptions(
				filter.DanglingFalse.ToPair(),
				util.NewPair("label", image.WerfLabel),
				util.NewPair("label", "werf-stage-signature"), // v1.1 legacy images
			)).Return(image.ImagesList{expectedImages[1]}, nil)

			backend.EXPECT().Images(ctx, buildImagesOptions(
				filter.DanglingFalse.ToPair(),

				util.NewPair("reference", "*client-id-*"),
				util.NewPair("reference", "*managed-image-*"),
				util.NewPair("reference", "*meta-*"),
				util.NewPair("reference", "*import-metadata-*"),
				util.NewPair("reference", "*-rejected"),

				util.NewPair("reference", "werf-client-id/*"),
				util.NewPair("reference", "werf-managed-images/*"),
				util.NewPair("reference", "werf-images-metadata-by-commit/*"),
				util.NewPair("reference", "werf-import-metadata/*"),
			)).Return(image.ImagesList{expectedImages[2]}, nil)

			stubs.StubFunc(&cleaner.werfGetWerfLastRunAtV1_1, time.Time{}, nil)

			images, err := cleaner.werfImages(ctx)
			Expect(err).To(Succeed())
			Expect(images).To(HaveLen(len(expectedImages)))

			Expect(images[0].ID).To(Equal(expectedImages[2].ID))
			Expect(images[1].ID).To(Equal(expectedImages[1].ID))
			Expect(images[2].ID).To(Equal(expectedImages[0].ID))
		})
	})

	DescribeTable("safeCleanupWerfImages",
		func(ctx SpecContext, runOpts RunGCOptions) {
			backend.EXPECT().Images(ctx, buildImagesOptions(
				filter.DanglingFalse.ToPair(),
				util.NewPair("label", image.WerfLabel),
			)).Return(image.ImagesList{}, nil)

			backend.EXPECT().Images(ctx, buildImagesOptions(
				filter.DanglingFalse.ToPair(),
				util.NewPair("label", image.WerfLabel),
				util.NewPair("label", "werf-stage-signature"), // v1.1 legacy images
			)).Return(image.ImagesList{}, nil)

			stubs.StubFunc(&cleaner.werfGetWerfLastRunAtV1_1, time.Unix(1, 0), nil)

			report, err := cleaner.safeCleanupWerfImages(ctx, runOpts, volumeutils.VolumeUsage{}, 0)

			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport(0)))
		},
		Entry(
			"should call backend.Images() multiple times and return result if opts.DryRun=true",
			RunGCOptions{
				DryRun: true,
			},
		),
		Entry(
			"should call backend.Images() multiple times and call cleanup.doSafeCleanupWerfImages() if opts.DryRun=false",
			RunGCOptions{},
		),
	)

	DescribeTable("doSafeCleanupWerfImages",
		func(ctx SpecContext, vu volumeutils.VolumeUsage, imgList image.ImagesList, vuStub volumeutils.VolumeUsage, rmiRefs []string, expectedReport cleanupReport) {
			if len(rmiRefs) > 0 {
				backend.EXPECT().Rmi(ctx, gomock.AnyOf(toAnySlice(rmiRefs)...), container_backend.RmiOpts{}).Return(nil).Times(3)
			}

			stubs.StubFunc(&cleaner.volumeutilsGetVolumeUsageByPath, vuStub, nil)

			report, err := cleaner.doSafeCleanupWerfImages(ctx, RunGCOptions{}, vu, 40.00, imgList)
			Expect(err).To(Succeed())
			Expect(report).To(Equal(expectedReport))
		},
		Entry(
			"should call cleaner.volumeutilsGetVolumeUsageByPath() two times if after first cleanup iteration reclaimed space not enough",
			volumeutils.VolumeUsage{
				UsedBytes:  800,
				TotalBytes: 1000,
			},
			image.ImagesList{
				{ID: "one", Size: 300, RepoDigests: []string{"one-digest"}},
				{ID: "two", Size: 500, RepoDigests: []string{"two-digest"}},
				{ID: "three", Size: 200, RepoDigests: []string{"three-digest"}},
			},
			volumeutils.VolumeUsage{
				UsedBytes:  600,
				TotalBytes: 1000,
			},
			[]string{"one-digest", "two-digest", "three-digest"},
			cleanupReport{
				ItemsDeleted:   []string{"one", "two", "three"},
				SpaceReclaimed: 200,
			},
		),
		Entry(
			"should not remove a dangling image and return empty report",
			volumeutils.VolumeUsage{
				UsedBytes:  800,
				TotalBytes: 1000,
			},
			image.ImagesList{
				{ID: "one-dangling", Size: 300}, // img without RepoTags and RegoDigests is always dangling image
			},
			volumeutils.VolumeUsage{
				UsedBytes:  800,
				TotalBytes: 1000,
			},
			[]string{},
			newCleanupReport(0),
		),
	)

	DescribeTable("removeImageByRepoTags",
		func(ctx context.Context, runOpts RunGCOptions, img image.Summary, lockNames, acquiredLocks []string, rmiErr error, expectedOk bool) {
			ctx = logging.WithLogger(ctx)

			locker.EXPECT().
				Acquire(gomock.AnyOf(toAnySlice(lockNames)...), lockgate.AcquireOptions{NonBlocking: true}).
				Return(len(acquiredLocks) > 0, lockgate.LockHandle{}, nil).
				Times(len(lockNames))

			locker.EXPECT().
				Release(lockgate.LockHandle{}).
				Return(nil).
				Times(len(acquiredLocks))

			rmiRefs := lo.Filter(img.RepoTags, func(ref string, _ int) bool {
				if len(lockNames) == 0 {
					return true
				}
				return lo.SomeBy(acquiredLocks, func(lockName string) bool {
					return container_backend.ImageLockName(ref) == lockName
				})
			})

			backend.EXPECT().Rmi(ctx, gomock.AnyOf(toAnySlice(rmiRefs)...), container_backend.RmiOpts{
				Force: runOpts.Force,
			}).Return(rmiErr).Times(len(rmiRefs))

			ok, err := cleaner.removeImageByRepoTags(ctx, runOpts, img)
			Expect(err).To(Succeed())
			Expect(ok).To(Equal(expectedOk))
		},
		Entry(
			"should return (false,nil) if no repo tags",
			RunGCOptions{},
			image.Summary{},
			[]string{},
			[]string{},
			nil,
			false,
		),
		Entry(
			"should return (true,nil) if repo has one tag '<none>:<none>'",
			RunGCOptions{},
			image.Summary{
				RepoTags: []string{"<none>:<none>"},
			},
			[]string{},
			[]string{},
			nil,
			true,
		),
		Entry(
			"should return (false,nil) if repo has one tag '<none>:<none>' and rmiErr!=nil",
			RunGCOptions{},
			image.Summary{
				RepoTags: []string{"<none>:<none>"},
			},
			[]string{},
			[]string{},
			errors.New("some rmi err"),
			false,
		),
		Entry(
			"should return (false,nil) if repo has one tag but the tag is locked",
			RunGCOptions{},
			image.Summary{
				RepoTags: []string{"locked-tag"},
			},
			[]string{
				container_backend.ImageLockName("locked-tag"),
			},
			[]string{},
			nil,
			false,
		),
		Entry(
			"should return (false,nil) if repo has one tag and the tag is not locked",
			RunGCOptions{},
			image.Summary{
				RepoTags: []string{"free-tag"},
			},
			[]string{
				container_backend.ImageLockName("free-tag"),
			},
			[]string{
				container_backend.ImageLockName("free-tag"),
			},
			nil,
			true,
		),
	)

	DescribeTable("removeImageByRepoDigests",
		func(ctx SpecContext, expectedOk bool) {
			ok, err := cleaner.removeImageByRepoDigests(ctx, RunGCOptions{}, image.Summary{})
			Expect(err).To(Succeed())
			Expect(ok).To(Equal(expectedOk))
		},
		Entry(
			"should return (false,nil) if no repo digests",
			false,
		),
	)

	Describe("RunGC", func() {
		It("should keep order of backend calls", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			options := RunGCOptions{
				AllowedStorageVolumeUsagePercentage:       0,
				AllowedStorageVolumeUsageMarginPercentage: 0,
				StoragePath: t.TempDir(),
				Force:       true,
			}

			stubs.StubFunc(&cleaner.volumeutilsGetVolumeUsageByPath, volumeutils.VolumeUsage{
				UsedBytes:  500,
				TotalBytes: 1000,
			}, nil)

			// prevent backend.Images() werfImagesByLastRun call
			stubs.StubFunc(&cleaner.werfGetWerfLastRunAtV1_1, time.Unix(1, 0), nil)

			containers := image.ContainerList{
				{ID: "import-server", Names: []string{fmt.Sprintf("/%s", image.ImportServerContainerNamePrefix)}},
			}

			images := image.ImagesList{
				{ID: "one", RepoDigests: []string{"digest one"}},
			}

			imagesFilters := filter.FilterList{
				filter.DanglingTrue,
				filter.NewFilter("label", image.WerfLabel),
				filter.NewFilter("until", "15m"),
			}

			// keep orders of backend calls
			gomock.InOrder(
				// use backend native GC pruning
				backend.EXPECT().PruneVolumes(ctx, prune.Options{}).Return(prune.Report{}, nil),
				backend.EXPECT().PruneImages(ctx, prune.Options{Filters: imagesFilters}).Return(prune.Report{}, nil),

				// list and remove werf containers
				backend.EXPECT().Containers(ctx, buildContainersOptions(
					image.ContainerFilter{Name: image.StageContainerNamePrefix},
					image.ContainerFilter{Name: image.ImportServerContainerNamePrefix},
				)).Return(containers, nil),

				locker.EXPECT().Acquire(container_backend.ContainerLockName(containers[0].Names[0][1:]), lockgate.AcquireOptions{NonBlocking: true}).Return(true, lockgate.LockHandle{}, nil),
				locker.EXPECT().Release(lockgate.LockHandle{}).Return(nil),

				backend.EXPECT().Rm(ctx, containers[0].ID, container_backend.RmOpts{Force: options.Force}).Return(nil),

				// list and remove werf images
				backend.EXPECT().Images(ctx, buildImagesOptions(
					filter.DanglingFalse.ToPair(),
					util.NewPair("label", image.WerfLabel),
				)).Return(images, nil),
				backend.EXPECT().Images(ctx, buildImagesOptions(
					filter.DanglingFalse.ToPair(),
					util.NewPair("label", image.WerfLabel),
					util.NewPair("label", "werf-stage-signature"), // v1.1 legacy images
				)).Return(image.ImagesList{}, nil),
				backend.EXPECT().Rmi(ctx, images[0].RepoDigests[0], container_backend.RmiOpts{
					Force: options.Force,
				}).Return(nil),
			)

			err := cleaner.RunGC(ctx, options)
			Expect(err).To(Succeed())
		})
	})
})

func toAnySlice[T any](s []T) []any {
	return lo.Map(s, func(item T, _ int) any {
		return any(item)
	})
}
