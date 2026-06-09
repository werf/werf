package cleaning

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("deleteOrphanedArtifacts", func() {
	DescribeTable("scenarios",
		func(setupMocks func(s *mock.MockStagesStorage), dryRun, expectError bool, expectedErrorSubstr string) {
			s := mock.NewMockStagesStorage(gomock.NewController(GinkgoT()))
			setupMocks(s)

			err := deleteOrphanedArtifacts(context.Background(), s, dryRun)
			if expectError {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErrorSubstr))
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("no orphans — returns nil",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedArtifactNames(gomock.Any()).Return(nil, nil)
			},
			false, false, "",
		),
		Entry("orphans deleted successfully",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedArtifactNames(gomock.Any()).Return([]string{"repo:sha256-abc123", "repo:sha256-def456"}, nil)
				s.EXPECT().DeleteArtifact(gomock.Any(), "repo:sha256-abc123").Return(nil)
				s.EXPECT().DeleteArtifact(gomock.Any(), "repo:sha256-def456").Return(nil)
			},
			false, false, "",
		),
		Entry("dry run — skips deletion",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedArtifactNames(gomock.Any()).Return([]string{"repo:sha256-abc123", "repo:sha256-def456"}, nil)
			},
			true, false, "",
		),
		Entry("GetOrphanedArtifactNames error — propagated",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedArtifactNames(gomock.Any()).Return(nil, errors.New("registry unavailable"))
			},
			false, true, "get orphaned artifacts",
		),
		Entry("non-fatal deletion error — continues to next",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedArtifactNames(gomock.Any()).Return([]string{"repo:sha256-abc123", "repo:sha256-def456"}, nil)
				s.EXPECT().DeleteArtifact(gomock.Any(), "repo:sha256-abc123").Return(errors.New("temporary network error"))
				s.EXPECT().DeleteArtifact(gomock.Any(), "repo:sha256-def456").Return(nil)
			},
			false, false, "",
		),
		Entry("fatal deletion error (UNAUTHORIZED) — stops and returns error",
			func(s *mock.MockStagesStorage) {
				s.EXPECT().GetOrphanedArtifactNames(gomock.Any()).Return([]string{"repo:sha256-abc123"}, nil)
				s.EXPECT().DeleteArtifact(gomock.Any(), "repo:sha256-abc123").Return(errors.New("UNAUTHORIZED"))
			},
			false, true, "UNAUTHORIZED",
		),
	)
})
