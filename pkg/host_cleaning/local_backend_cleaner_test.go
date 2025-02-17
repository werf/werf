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

	Describe("checkBackendStorage", func() {
		It("should return 1 image, 0 volume usage and 0 total images bytes when we cover all nested cases", func() {
			// testdata covers all nested "if cases" in cleaner.checkBackendStorage()

			stubs.StubFunc(&cleaner.volumeutilsGetVolumeUsageByPath, volumeutils.VolumeUsage{}, nil)

			backend.EXPECT().Images(ctx, buildImagesOptions(
				util.NewPair("label", image.WerfLabel),
				util.NewPair("label", image.WerfStageDigestLabel),
			)).Return(image.ImagesList{
				{
					Labels:   map[string]string{image.WerfLabel: "project-name"},
					RepoTags: []string{"project-name:"},
				},
			}, nil).Times(1)

			backend.EXPECT().Images(ctx, buildImagesOptions(
				util.NewPair("label", image.WerfLabel),
				util.NewPair("label", "werf-stage-signature"), // v1.1 legacy images
			)).Return(image.ImagesList{
				{
					RepoTags: []string{"werf-stages-storage/something"},
				},
			}, nil).Times(1)

			stubs.StubFunc(&cleaner.werfGetWerfLastRunAtV1_1, time.Time{}, nil)

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
			)).Return(image.ImagesList{
				{
					Size: 1,
				},
				// -------------
				{
					RepoTags: []string{"<none>:<none>"},
				},
				{
					RepoTags: []string{"lru_tag"},
				},
			}, nil)

			stubs.StubFunc(&cleaner.lrumetaGetImageLastAccessTime, time.Time{}, nil)

			result, err := cleaner.checkBackendStorage(ctx, t.TempDir())
			Expect(err).To(Succeed())
			Expect(result.VolumeUsage).To(BeZero())
			Expect(result.TotalImagesBytes).To(BeZero())
			Expect(result.ImagesDescs).To(HaveLen(1))
			Expect(result.ImagesDescs[0].ImageSummary).To(Equal(image.Summary{
				RepoTags: []string{"lru_tag"},
			}))
		})
	})

	Describe("RunGC", func() {
		// TODO: cover it
	})
})
