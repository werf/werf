package cleaning

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("deleteOrphanedSbomImages", func() {
	DescribeTable("scenarios",
		func(setupMocks func(s *mock.MockStagesStorage), dryRun, expectError bool, expectedErrorSubstr string) {
			s := mock.NewMockStagesStorage(gomock.NewController(GinkgoT()))
			setupMocks(s)

			err := deleteOrphanedSbomImages(context.Background(), s, dryRun)
			if expectError {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErrorSubstr))
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("no orphans — returns nil",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedSbomImageNames(gomock.Any()).Return(nil, nil)
			},
			false, false, "",
		),
		Entry("orphans deleted successfully",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedSbomImageNames(gomock.Any()).Return([]string{"repo:abc123-sbom", "repo:def456-sbom"}, nil)
				s.EXPECT().DeleteSbomImage(gomock.Any(), "repo:abc123-sbom").Return(nil)
				s.EXPECT().DeleteSbomImage(gomock.Any(), "repo:def456-sbom").Return(nil)
			},
			false, false, "",
		),
		Entry("dry run — skips deletion",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedSbomImageNames(gomock.Any()).Return([]string{"repo:abc123-sbom", "repo:def456-sbom"}, nil)
			},
			true, false, "",
		),
		Entry("GetOrphanedSbomImageNames error — propagated",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedSbomImageNames(gomock.Any()).Return(nil, errors.New("registry unavailable"))
			},
			false, true, "get orphaned sbom images",
		),
		Entry("non-fatal deletion error — continues to next",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedSbomImageNames(gomock.Any()).Return([]string{"repo:abc123-sbom", "repo:def456-sbom"}, nil)
				s.EXPECT().DeleteSbomImage(gomock.Any(), "repo:abc123-sbom").Return(errors.New("temporary network error"))
				s.EXPECT().DeleteSbomImage(gomock.Any(), "repo:def456-sbom").Return(nil)
			},
			false, false, "",
		),
		Entry("fatal deletion error (UNAUTHORIZED) — stops and returns error",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedSbomImageNames(gomock.Any()).Return([]string{"repo:abc123-sbom"}, nil)
				s.EXPECT().DeleteSbomImage(gomock.Any(), "repo:abc123-sbom").Return(errors.New("UNAUTHORIZED"))
			},
			false, true, "UNAUTHORIZED",
		),
	)
})
