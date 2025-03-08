package host_cleaning

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"
	"go.uber.org/mock/gomock"

	"github.com/werf/werf/v2/pkg/container_backend/info"
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

	Describe("RunGC", func() {
		// TODO: cover it
	})
})
