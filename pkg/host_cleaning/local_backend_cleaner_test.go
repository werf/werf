package host_cleaning

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"
	"go.uber.org/mock/gomock"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/container_backend/info"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/volumeutils"
	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("LocalBackendCleaner", func() {
	t := GinkgoT()

	var cleaner *LocalBackendCleaner
	var backend *mock.MockContainerBackend

	var stubs *gostub.Stubs
	var ctx context.Context

	BeforeEach(func() {
		backend = mock.NewMockContainerBackend(gomock.NewController(t))
		var err error
		cleaner, err = NewLocalBackendCleaner(backend)
		Expect(errors.Is(err, ErrUnsupportedContainerBackend)).To(BeTrue())
		Expect(cleaner).NotTo(BeNil())
		ctx = context.Background()
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
		It("should return err if backend.Info() returns err", func() {
			err0 := errors.New("some err")
			backend.EXPECT().Info(ctx).Return(info.Info{}, err0)

			_, err := cleaner.backendStoragePath(ctx, "")

			Expect(errors.Is(err, err0)).To(BeTrue())
		})
	})

	Describe("ShouldRunAutoGC", func() {
		It("should return true if cleanup needed", func() {
			result, err := cleaner.ShouldRunAutoGC(ctx, RunAutoGCOptions{
				AllowedStorageVolumeUsagePercentage: 0,
				StoragePath:                         t.TempDir(),
			})
			Expect(err).To(Succeed())
			Expect(result).To(BeTrue())
		})
		It("should return false if cleanup not needed", func() {
			result, err := cleaner.ShouldRunAutoGC(ctx, RunAutoGCOptions{
				AllowedStorageVolumeUsagePercentage: 1000,
				StoragePath:                         t.TempDir(),
			})
			Expect(err).To(Succeed())
			Expect(result).To(BeFalse())
		})
	})

	Describe("pruneBuildCache", func() {
		It("should return err if opts.DryRun=true", func() {
			_, err := cleaner.pruneBuildCache(ctx, RunGCOptions{
				DryRun: true,
			})
			Expect(errors.Is(err, errOptionDryRunNotSupported)).To(BeTrue())
		})
		It("should call backend.PruneBuildCache() if opts.DryRun=false", func() {
			backend.EXPECT().PruneBuildCache(ctx, prune.Options{}).Return(prune.Report{}, nil)

			report, err := cleaner.pruneBuildCache(ctx, RunGCOptions{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport()))
		})
	})

	Describe("pruneContainers", func() {
		It("should return err if opts.DryRun=true", func() {
			_, err := cleaner.pruneContainers(ctx, RunGCOptions{
				DryRun: true,
			})
			Expect(errors.Is(err, errOptionDryRunNotSupported)).To(BeTrue())
		})
		It("should call backend.PruneContainers() if opts.DryRun=true", func() {
			backend.EXPECT().PruneContainers(ctx, prune.Options{}).Return(prune.Report{}, nil)

			report, err := cleaner.pruneContainers(ctx, RunGCOptions{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport()))
		})
	})

	Describe("pruneImages", func() {
		It("should call backend.Images() to find dandling images if opts.DryRun=true", func() {
			backend.EXPECT().Images(ctx, buildImagesOptions(
				util.NewPair("dangling", "true"),
			)).Return(image.ImagesList{}, nil)

			report, err := cleaner.pruneImages(ctx, RunGCOptions{
				DryRun: true,
			})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport()))
		})
		It("should call backend.PruneImages() if opts.DryRun=false", func() {
			backend.EXPECT().PruneImages(ctx, prune.Options{}).Return(prune.Report{}, nil)

			report, err := cleaner.pruneImages(ctx, RunGCOptions{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport()))
		})
	})

	Describe("pruneVolumes", func() {
		It("should return err if opts.DryRun=true", func() {
			_, err := cleaner.pruneVolumes(ctx, RunGCOptions{
				DryRun: true,
			})
			Expect(errors.Is(err, errOptionDryRunNotSupported)).To(BeTrue())
		})
		It("should call backend.PruneVolumes() if opts.DryRun=false", func() {
			backend.EXPECT().PruneVolumes(ctx, prune.Options{}).Return(prune.Report{}, nil)

			report, err := cleaner.pruneVolumes(ctx, RunGCOptions{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport()))
		})
	})

	Describe("safeCleanupWerfContainers", func() {
		BeforeEach(func() {
			backend.EXPECT().Containers(ctx, buildContainersOptions(
				image.ContainerFilter{Name: image.StageContainerNamePrefix},
				image.ContainerFilter{Name: image.ImportServerContainerNamePrefix},
			)).Return(image.ContainerList{}, nil)
		})
		It("should call backend.Containers() and return result if opts.DryRun=true", func() {
			report, err := cleaner.safeCleanupWerfContainers(ctx, RunGCOptions{
				DryRun: true,
			}, volumeutils.VolumeUsage{})

			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport()))
		})
		It("should call backend.Containers() and call cleanup.doSafeCleanupWerfContainers() if opts.DryRun=false", func() {
			report, err := cleaner.safeCleanupWerfContainers(ctx, RunGCOptions{}, volumeutils.VolumeUsage{})
			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport()))
		})
	})

	Describe("werfImages", func() {
		It("should return images as merged and sorted result of several backend calls", func() {
			expectedImages := []image.Summary{
				{ID: "one", Created: time.Unix(300, 0)},
				{ID: "two", Created: time.Unix(200, 0)},
				{ID: "three", Created: time.Unix(100, 0)},
			}

			backend.EXPECT().Images(ctx, buildImagesOptions(
				util.NewPair("label", image.WerfLabel),
				util.NewPair("label", image.WerfStageDigestLabel),
			)).Return(image.ImagesList{expectedImages[0]}, nil)

			backend.EXPECT().Images(ctx, buildImagesOptions(
				util.NewPair("label", image.WerfLabel),
				util.NewPair("label", "werf-stage-signature"), // v1.1 legacy images
			)).Return(image.ImagesList{expectedImages[1]}, nil)

			backend.EXPECT().Images(ctx, buildImagesOptions(
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

	Describe("safeCleanupWerfImages", func() {
		BeforeEach(func() {
			backend.EXPECT().Images(ctx, buildImagesOptions(
				util.NewPair("label", image.WerfLabel),
				util.NewPair("label", image.WerfStageDigestLabel),
			)).Return(image.ImagesList{}, nil)

			backend.EXPECT().Images(ctx, buildImagesOptions(
				util.NewPair("label", image.WerfLabel),
				util.NewPair("label", "werf-stage-signature"), // v1.1 legacy images
			)).Return(image.ImagesList{}, nil)

			stubs.StubFunc(&cleaner.werfGetWerfLastRunAtV1_1, time.Unix(1, 0), nil)
		})
		It("should call backend.Images() multiple times and return result if opts.DryRun=true", func() {
			report, err := cleaner.safeCleanupWerfImages(ctx, RunGCOptions{
				DryRun: true,
			}, volumeutils.VolumeUsage{}, 0)

			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport()))
		})
		It("should call backend.Images() multiple times and call cleanup.doSafeCleanupWerfImages() if opts.DryRun=false", func() {
			report, err := cleaner.safeCleanupWerfImages(ctx, RunGCOptions{}, volumeutils.VolumeUsage{}, 0)
			Expect(err).To(Succeed())
			Expect(report).To(Equal(newCleanupReport()))
		})
	})

	Describe("RunGC", func() {
		// TODO: cover it
	})
})
