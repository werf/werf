package host_cleaning

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"
	"go.uber.org/mock/gomock"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/lockgate"
	"github.com/werf/werf/v2/pkg/cleanup_report"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/filter"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/volumeutils"
	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("LocalBackendCleaner host cleanup report", func() {
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
		stubs = gostub.New()
	})
	AfterEach(func() {
		stubs.Reset()
	})

	containers := image.ContainerList{
		{ID: "cont-1", Names: []string{fmt.Sprintf("/%ssome-name", image.StageContainerNamePrefix)}},
	}

	expectBackendCalls := func(ctx context.Context) {
		// Volume usage drops from 500 to 200 once the backend prune phases are done,
		// so the container phase reclaims nothing and the loop over werf images stops immediately.
		usages := []volumeutils.VolumeUsage{
			{UsedBytes: 500, TotalBytes: 1000},
			{UsedBytes: 200, TotalBytes: 1000},
		}
		callIndex := 0
		stubs.Stub(&cleaner.volumeutilsGetVolumeUsageByPath, func(_ context.Context, _ string) (volumeutils.VolumeUsage, error) {
			usage := usages[min(callIndex, len(usages)-1)]
			callIndex++
			return usage, nil
		})

		stubs.StubFunc(&cleaner.werfGetWerfLastRunAtV1_1, time.Unix(1, 0), nil)

		backend.EXPECT().PruneVolumes(ctx, prune.Options{}).
			Return(prune.Report{ItemsDeleted: []string{"vol-1"}, SpaceReclaimed: 100}, nil)
		backend.EXPECT().PruneImages(ctx, prune.Options{Filters: filter.FilterList{
			filter.DanglingTrue,
			filter.NewFilter("label", image.WerfLabel),
			filter.NewFilter("until", "15m"),
		}}).Return(prune.Report{ItemsDeleted: []string{"img-1"}, SpaceReclaimed: 200}, nil)

		backend.EXPECT().Containers(ctx, buildContainersOptions(
			image.ContainerFilter{Name: image.StageContainerNamePrefix},
			image.ContainerFilter{Name: image.ImportServerContainerNamePrefix},
		)).Return(containers, nil)
		locker.EXPECT().Acquire(container_backend.ContainerLockName(werfContainerName(containers[0])), lockgate.AcquireOptions{NonBlocking: true}).
			Return(true, lockgate.LockHandle{}, nil)
		locker.EXPECT().Release(lockgate.LockHandle{}).Return(nil)
		backend.EXPECT().Rm(ctx, containers[0].ID, container_backend.RmOpts{}).Return(nil)

		backend.EXPECT().Images(ctx, buildImagesOptions(
			filter.DanglingFalse.ToPair(),
			util.NewPair("label", image.WerfLabel),
		)).Return(image.ImagesList{}, nil)
		backend.EXPECT().Images(ctx, buildImagesOptions(
			filter.DanglingFalse.ToPair(),
			util.NewPair("label", image.WerfLabel),
			util.NewPair("label", "werf-stage-signature"),
		)).Return(image.ImagesList{}, nil)
	}

	It("should record deleted volumes, images and containers with reclaimed space", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)
		expectBackendCalls(ctx)

		report := cleanup_report.NewHostReport("host cleanup", false)

		Expect(cleaner.RunGC(ctx, RunGCOptions{StoragePath: t.TempDir(), Report: report})).To(Succeed())

		Expect(report.Deleted).To(ConsistOf(
			cleanup_report.Item{Type: cleanup_report.ItemTypeVolume, ID: "vol-1"},
			cleanup_report.Item{Type: cleanup_report.ItemTypeImage, ID: "img-1"},
			cleanup_report.Item{Type: cleanup_report.ItemTypeContainer, ID: "cont-1"},
		))
		Expect(report.SpaceReclaimed).To(Equal(uint64(300)))
	})

	It("should not fail without a report", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)
		expectBackendCalls(ctx)

		Expect(cleaner.RunGC(ctx, RunGCOptions{StoragePath: t.TempDir()})).To(Succeed())
	})
})
