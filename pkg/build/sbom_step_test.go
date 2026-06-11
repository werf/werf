package build

import (
	"context"
	"errors"
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
		It("should return error if image info is nil", func(ctx SpecContext) {
			step := &sbomStep{}
			_, err := step.GetImageBOM(ctx, "app", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("image info is nil"))
		})

		It("should return fatal error if image digest is empty", func(ctx SpecContext) {
			step := &sbomStep{}
			imgInfo := &werfImage.Info{Name: "app:latest"}
			_, err := step.GetImageBOM(ctx, "app", imgInfo)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ErrSbomNotRequired)).To(BeFalse())
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

	Describe("isTrustedBuilderImage()", func() {
		DescribeTable("should detect trusted builder images",
			func(labels map[string]string, expected bool) {
				Expect(isTrustedBuilderImage(labels)).To(Equal(expected))
			},
			Entry("nil labels", nil, false),
			Entry("empty labels", map[string]string{}, false),
			Entry("label set to false", map[string]string{werfImage.DeckhouseInternalBuilderLabel: "false"}, false),
			Entry("label set to true", map[string]string{werfImage.DeckhouseInternalBuilderLabel: "true"}, true),
			Entry("other labels without builder", map[string]string{"foo": "bar", "baz": "qux"}, false),
			Entry("other labels with builder true", map[string]string{"foo": "bar", werfImage.DeckhouseInternalBuilderLabel: "true", "baz": "qux"}, true),
		)
	})

	Describe("GetImageBOM() with trusted builder image", func() {
		It("should return fatal error even for builder image when error is not 'not found'", func(ctx SpecContext) {
			step := &sbomStep{}

			imageInfo := &werfImage.Info{
				Name:       "docker.io/namespace/repo:builder-tag",
				Repository: "docker.io/namespace/repo",
				Labels: map[string]string{
					werfImage.DeckhouseInternalBuilderLabel: "true",
				},
			}

			_, err := step.GetImageBOM(ctx, "builder-image", imageInfo)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ErrSbomNotRequired)).To(BeFalse(),
				"non-'not found' errors must NOT be treated as 'not required', got: %v", err)
		})

		It("should return fatal error for non-builder image when SBOM pull fails", func(ctx SpecContext) {
			step := &sbomStep{}

			imageInfo := &werfImage.Info{
				Name:       "docker.io/namespace/repo:some-tag",
				Repository: "docker.io/namespace/repo",
				Labels:     map[string]string{},
			}

			_, err := step.GetImageBOM(ctx, "app", imageInfo)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ErrSbomNotRequired)).To(BeFalse())
		})
	})
})
