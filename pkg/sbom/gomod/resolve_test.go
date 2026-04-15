package gomod

import (
	"context"

	cdxgo "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("ResolveUnknownVersions", func() {
	DescribeTable("ResolveUnknownVersions",
		func(ctx context.Context, setupMocks func(ctx context.Context, gitRepo *mock.MockGitRepo) (git_repo.GitRepo, string, string, *cdxgo.BOM), errMatcher OmegaMatcher, verify func(*cdxgo.BOM)) {
			ctx = logging.WithLogger(ctx)

			var gitRepo git_repo.GitRepo
			var commit string
			var imageContext string
			var bom *cdxgo.BOM

			if setupMocks != nil {
				mockRepo := mock.NewMockGitRepo(gomock.NewController(GinkgoT()))
				gitRepo, commit, imageContext, bom = setupMocks(ctx, mockRepo)
			}

			result, err := NewBOMPatcher(gitRepo, commit, imageContext).Apply(ctx, bom)
			Expect(err).To(errMatcher)

			if verify != nil {
				verify(result)
			}
		},
		Entry(
			"resolves versions from git tag",
			func(ctx context.Context, gitRepo *mock.MockGitRepo) (git_repo.GitRepo, string, string, *cdxgo.BOM) {
				commit := "8d0a3fced4f1a98b6f51442e2a73c8417b8f45af"
				goModContent := []byte("module example.com/app\n\nreplace example.com/replaced => ./local\n")

				gitRepo.EXPECT().IsCommitFileExist(ctx, commit, "app/go.mod").Return(true, nil)
				gitRepo.EXPECT().ReadCommitFile(ctx, commit, "app/go.mod").Return(goModContent, nil)
				gitRepo.EXPECT().TagsList(ctx).Return([]string{"v0.9.0", "v1.2.0"}, nil)
				gitRepo.EXPECT().TagCommit(ctx, "v0.9.0").Return("deadbeef", nil)
				gitRepo.EXPECT().TagCommit(ctx, "v1.2.0").Return(commit, nil)

				return gitRepo, commit, "app", &cdxgo.BOM{Components: &[]cdxgo.Component{
					{Name: "example.com/app", Version: "UNKNOWN", Type: cdxgo.ComponentTypeLibrary},
				}}
			},
			Succeed(),
			func(result *cdxgo.BOM) {
				Expect((*result.Components)[0].Version).To(Equal("v1.2.0"))
			},
		),
		Entry(
			"keeps BOM unchanged when go.mod is missing",
			func(ctx context.Context, gitRepo *mock.MockGitRepo) (git_repo.GitRepo, string, string, *cdxgo.BOM) {
				commit := "8d0a3fced4f1a98b6f51442e2a73c8417b8f45af"

				gitRepo.EXPECT().IsCommitFileExist(ctx, commit, "app/go.mod").Return(false, nil)

				return gitRepo, commit, "app", &cdxgo.BOM{Components: &[]cdxgo.Component{
					{Name: "example.com/app", Version: "UNKNOWN", Type: cdxgo.ComponentTypeLibrary},
				}}
			},
			Succeed(),
			func(result *cdxgo.BOM) {
				Expect((*result.Components)[0].Version).To(Equal("UNKNOWN"))
			},
		),
		Entry(
			"does not update non-go components",
			func(ctx context.Context, gitRepo *mock.MockGitRepo) (git_repo.GitRepo, string, string, *cdxgo.BOM) {
				commit := "8d0a3fced4f1a98b6f51442e2a73c8417b8f45af"

				gitRepo.EXPECT().IsCommitFileExist(ctx, commit, "app/go.mod").Return(true, nil)
				gitRepo.EXPECT().ReadCommitFile(ctx, commit, "app/go.mod").Return([]byte("module example.com/app\n"), nil)
				gitRepo.EXPECT().TagsList(ctx).Return([]string{}, nil)

				return gitRepo, commit, "app", &cdxgo.BOM{Components: &[]cdxgo.Component{
					{Name: "example.com/other", Version: "UNKNOWN", Type: cdxgo.ComponentTypeApplication},
				}}
			},
			Succeed(),
			func(result *cdxgo.BOM) {
				Expect((*result.Components)[0].Version).To(Equal("UNKNOWN"))
			},
		),
		Entry(
			"returns error for non-local replace",
			func(ctx context.Context, gitRepo *mock.MockGitRepo) (git_repo.GitRepo, string, string, *cdxgo.BOM) {
				commit := "8d0a3fced4f1a98b6f51442e2a73c8417b8f45af"

				gitRepo.EXPECT().IsCommitFileExist(ctx, commit, "app/go.mod").Return(true, nil)
				gitRepo.EXPECT().ReadCommitFile(ctx, commit, "app/go.mod").Return([]byte("module example.com/app\n\nreplace example.com/replaced => example.com/other v1.2.3\n"), nil)

				return gitRepo, commit, "app", &cdxgo.BOM{}
			},
			MatchError(ContainSubstring("non-local replace")),
			nil,
		),
	)
})
