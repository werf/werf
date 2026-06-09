package build

import (
	"context"
	"path/filepath"

	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	werfImage "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/sbom/gomod"
	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("SbomStep", func() {
	Describe("GetImageBOM()", func() {
		It("should return ErrSbomNotAvailable if image info is nil", func() {
			step := &sbomStep{}
			_, err := step.GetImageBOM(context.Background(), "app", "tag", nil)
			Expect(err).To(MatchError(ErrSbomNotAvailable))
		})

		It("should return ErrSbomNotAvailable if image digest is empty", func() {
			step := &sbomStep{}
			imgInfo := &werfImage.Info{Name: "app:latest"}
			_, err := step.GetImageBOM(context.Background(), "app", "tag", imgInfo)
			Expect(err).To(MatchError(ErrSbomNotAvailable))
		})
	})

	Describe("BOMPatcher (gomod)", func() {
		DescribeTable("Apply()",
			func(
				ctx context.Context,
				setupGitRepo func(ctx context.Context, repo *mock.MockGitRepo, commit, imageContext string),
			) {
				repo := mock.NewMockGitRepo(gomock.NewController(GinkgoT()))
				commit := "0123456789abcdef0123456789abcdef01234567"
				imageContext := "app"
				setupGitRepo(ctx, repo, commit, imageContext)

				patcher := gomod.NewBOMPatcher(repo, commit, imageContext)
				bom := &cdx.BOM{
					Metadata: &cdx.Metadata{
						Component: &cdx.Component{
							Name: "app",
						},
					},
				}

				res, err := patcher.Apply(ctx, bom)
				Expect(err).ToNot(HaveOccurred())
				Expect(res).ToNot(BeNil())
			},
			Entry(
				"[go.mod]: should skip version resolution when go.mod is missing",
				logging.WithLogger(context.Background()),
				func(ctx context.Context, repo *mock.MockGitRepo, commit, imageContext string) {
					repo.EXPECT().IsCommitFileExist(ctx, commit, filepath.Join(imageContext, "go.mod")).Return(false, nil)
				},
			),
			Entry(
				"[go.mod]: should use tag version when tag matches commit",
				logging.WithLogger(context.Background()),
				func(ctx context.Context, repo *mock.MockGitRepo, commit, imageContext string) {
					goModPath := filepath.Join(imageContext, "go.mod")
					repo.EXPECT().IsCommitFileExist(ctx, commit, goModPath).Return(true, nil)
					repo.EXPECT().ReadCommitFile(ctx, commit, goModPath).Return([]byte("module example.com/app\n"), nil)
					repo.EXPECT().TagsList(ctx).Return([]string{"v1.2.3"}, nil)
					repo.EXPECT().TagCommit(ctx, "v1.2.3").Return(commit, nil)
				},
			),
			Entry(
				"[go.mod]: should fallback to pseudo version when tag mismatch",
				logging.WithLogger(context.Background()),
				func(ctx context.Context, repo *mock.MockGitRepo, commit, imageContext string) {
					goModPath := filepath.Join(imageContext, "go.mod")
					repo.EXPECT().IsCommitFileExist(ctx, commit, goModPath).Return(true, nil)
					repo.EXPECT().ReadCommitFile(ctx, commit, goModPath).Return([]byte("module example.com/app\n"), nil)
					repo.EXPECT().TagsList(ctx).Return([]string{"v1.2.3"}, nil)
					repo.EXPECT().TagCommit(ctx, "v1.2.3").Return("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", nil)
				},
			),
		)
	})
})
