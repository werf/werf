package host_cleaning_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/host_cleaning"
)

var _ = Describe("LocalBackendCleaner", func() {
	t := GinkgoT()

	var cleaner *host_cleaning.LocalBackendCleaner
	var backend *MockContainerBackend

	BeforeEach(func() {
		backend = NewMockContainerBackend(gomock.NewController(t))
		var err error
		cleaner, err = host_cleaning.NewLocalBackendCleaner(backend)
		Expect(errors.Is(err, host_cleaning.ErrUnsupportedContainerBackend)).To(BeTrue())
		Expect(cleaner).NotTo(BeNil())
	})

	Describe("ShouldRunAutoGC", func() {
		It("should return true if cleanup needed", func() {
			result, err := cleaner.ShouldRunAutoGC(context.Background(), host_cleaning.RunAutoGCOptions{
				AllowedStorageVolumeUsagePercentage: 0,
				StoragePath:                         t.TempDir(),
			})
			Expect(err).To(Succeed())
			Expect(result).To(BeTrue())
		})
		It("should return false if cleanup not needed", func() {
			result, err := cleaner.ShouldRunAutoGC(context.Background(), host_cleaning.RunAutoGCOptions{
				AllowedStorageVolumeUsagePercentage: 1000,
				StoragePath:                         t.TempDir(),
			})
			Expect(err).To(Succeed())
			Expect(result).To(BeFalse())
		})
	})

	// TODO: cover RunGC()
})
