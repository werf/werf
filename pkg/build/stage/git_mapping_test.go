package stage_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
)

var _ = Describe("GitMapping", func() {
	var gitMapping *stage.GitMapping

	BeforeEach(func() {
		gitMapping = stage.NewGitMapping()
	})

	DescribeTable("getting built image commit info from labels",
		func(commitLabel string) {
			gitRepo := NewGitRepoStub("own", true, "irrelevant-commit-id")
			gitMapping.SetGitRepo(gitRepo)

			builtImageLabels := map[string]string{
				gitMapping.ImageGitCommitLabel(): commitLabel,
			}

			info, err := gitMapping.GetBuiltImageCommitInfo(builtImageLabels)

			Expect(err).To(Succeed())
			Expect(info.Commit).To(Equal(commitLabel))
		},

		Entry("with regular commit",
			"08885f44226e48e19448e672b4db56b0f710a8b5"),
	)

	DescribeTable("getting latest commit info",
		func(ctx context.Context, currentCommit string) {
			ctx = logging.WithLogger(ctx)

			c := NewConveyorStub()

			gitRepo := NewGitRepoStub("own", true, currentCommit)
			gitMapping.SetGitRepo(gitRepo)

			info, err := gitMapping.GetLatestCommitInfo(ctx, c)
			Expect(err).To(Succeed())
			Expect(info.Commit).To(Equal(currentCommit))
		},

		Entry("with regular commit",
			"ae6feb44da273003cfada392c33dfe33748a5e2f"),
	)

	type BaseCommitForPrevBuiltImageCheckData struct {
		BuiltCommitLabel string
		CurrentCommit    string
	}

	DescribeTable("getting base commit from prev built image",
		func(ctx context.Context, data BaseCommitForPrevBuiltImageCheckData, checkResultFunc func(string, BaseCommitForPrevBuiltImageCheckData)) {
			ctx = logging.WithLogger(ctx)

			c := NewConveyorStub()
			containerBackend := stage.NewContainerBackendStub()

			gitRepo := NewGitRepoStub("own", true, data.CurrentCommit)
			gitMapping.SetGitRepo(gitRepo)

			prevBuiltImage := NewBuiltImageStub("stub", &image.StageDesc{
				Info: &image.Info{
					Labels: map[string]string{
						gitMapping.ImageGitCommitLabel(): data.BuiltCommitLabel,
					},
				},
			})
			img := stage.NewStageImage(containerBackend, "", prevBuiltImage)

			baseCommit, err := gitMapping.GetBaseCommitForPrevBuiltImage(ctx, c, img)
			Expect(err).To(Succeed())

			checkResultFunc(baseCommit, data)
		},

		Entry("when current commit the same as previously built",
			BaseCommitForPrevBuiltImageCheckData{
				BuiltCommitLabel: "ae6feb44da273003cfada392c33dfe33748a5e2f",
				CurrentCommit:    "ae6feb44da273003cfada392c33dfe33748a5e2f",
			},
			func(baseCommit string, data BaseCommitForPrevBuiltImageCheckData) {
				Expect(baseCommit).To(Equal(data.CurrentCommit))
			}),

		Entry("when current commit differs from previously built",
			BaseCommitForPrevBuiltImageCheckData{
				BuiltCommitLabel: "ae6feb44da273003cfada392c33dfe33748a5e2f",
				CurrentCommit:    "a4009816bf7d8f62b2d1cd3331d36f45b3df7d99",
			},
			func(baseCommit string, data BaseCommitForPrevBuiltImageCheckData) {
				Expect(baseCommit).To(Equal(data.BuiltCommitLabel))
			}),
	)
})

type BuiltImageStub struct {
	container_backend.LegacyImageInterface

	name      string
	stageDesc *image.StageDesc
}

func NewBuiltImageStub(name string, stageDesc *image.StageDesc) *BuiltImageStub {
	return &BuiltImageStub{
		name:      name,
		stageDesc: stageDesc,
	}
}

func (img *BuiltImageStub) Name() string {
	return img.name
}

func (img *BuiltImageStub) GetStageDesc() *image.StageDesc {
	return img.stageDesc
}

type ConveyorStub struct {
	stage.Conveyor
}

func NewConveyorStub() *ConveyorStub {
	return &ConveyorStub{}
}

type GitRepoStub struct {
	git_repo.GitRepo

	isLocal        bool
	name           string
	headCommitHash string
}

func NewGitRepoStub(name string, isLocal bool, headCommitHash string) *GitRepoStub {
	return &GitRepoStub{
		name:           name,
		isLocal:        isLocal,
		headCommitHash: headCommitHash,
	}
}

func (gitRepo *GitRepoStub) GetName() string {
	return gitRepo.name
}

func (gitRepo *GitRepoStub) IsLocal() bool {
	return gitRepo.isLocal
}

func (gitRepo *GitRepoStub) HeadCommitHash(_ context.Context) (string, error) {
	return gitRepo.headCommitHash, nil
}
