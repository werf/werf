package stage_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/image"
)

var _ = Describe("GitMapping", func() {
	var gitMapping *stage.GitMapping

	BeforeEach(func() {
		gitMapping = stage.NewGitMapping()
	})

	type BuiltImageCommitInfoCheckData struct {
		CommitLabel                 string
		VirtualMergeLabel           string
		VirtualMergeFromCommitLabel string
		VirtualMergeIntoCommitLabel string
	}

	DescribeTable("getting built image commit info from labels",
		func(data BuiltImageCommitInfoCheckData, checkResult func(stage.ImageCommitInfo, BuiltImageCommitInfoCheckData)) {
			gitRepo := NewGitRepoStub("own", true, "irrelevant-commit-id")
			gitMapping.SetGitRepo(gitRepo)

			builtImageLabels := map[string]string{
				gitMapping.ImageGitCommitLabel():         data.CommitLabel,
				gitMapping.VirtualMergeLabel():           data.VirtualMergeLabel,
				gitMapping.VirtualMergeFromCommitLabel(): data.VirtualMergeFromCommitLabel,
				gitMapping.VirtualMergeIntoCommitLabel(): data.VirtualMergeIntoCommitLabel,
			}

			info, err := gitMapping.GetBuiltImageCommitInfo(builtImageLabels)

			Expect(err).To(Succeed())
			Expect(info.Commit).To(Equal(data.CommitLabel))

			checkResult(info, data)
		},

		Entry("when using virtual merge commit",
			BuiltImageCommitInfoCheckData{
				CommitLabel:                 "a4009816bf7d8f62b2d1cd3331d36f45b3df7d99",
				VirtualMergeLabel:           "true",
				VirtualMergeFromCommitLabel: "c1159b9ce05143df3dd6450ee9d749c9642320d2",
				VirtualMergeIntoCommitLabel: "cdb9a2e6ab3fec90f3ea12d4e4728ec4bf3f74dc",
			},
			func(info stage.ImageCommitInfo, data BuiltImageCommitInfoCheckData) {
				Expect(info.VirtualMerge).To(Equal(true))
				Expect(info.VirtualMergeFromCommit).To(Equal(data.VirtualMergeFromCommitLabel))
				Expect(info.VirtualMergeIntoCommit).To(Equal(data.VirtualMergeIntoCommitLabel))
			}),

		Entry("when not using virtual merge commit",
			BuiltImageCommitInfoCheckData{
				CommitLabel:       "08885f44226e48e19448e672b4db56b0f710a8b5",
				VirtualMergeLabel: "false",
			},
			func(info stage.ImageCommitInfo, data BuiltImageCommitInfoCheckData) {
				Expect(info.VirtualMerge).To(Equal(false))
			}),
	)

	type LatestCommitInfoCheckData struct {
		CurrentCommit               string
		IsCurrentCommitVirtualMerge bool
	}

	DescribeTable("getting latest commit info",
		func(data LatestCommitInfoCheckData, checkResultFunc func(stage.ImageCommitInfo, LatestCommitInfoCheckData)) {
			ctx := context.Background()
			c := NewConveyorStub(stage.VirtualMergeOptions{VirtualMerge: data.IsCurrentCommitVirtualMerge})

			gitRepo := NewGitRepoStub("own", true, data.CurrentCommit)
			gitMapping.SetGitRepo(gitRepo)

			info, err := gitMapping.GetLatestCommitInfo(ctx, c)
			Expect(err).To(Succeed())
			Expect(info.Commit).To(Equal(data.CurrentCommit))

			checkResultFunc(info, data)
		},

		Entry("when current commit is not virtual merge",
			LatestCommitInfoCheckData{
				CurrentCommit:               "ae6feb44da273003cfada392c33dfe33748a5e2f",
				IsCurrentCommitVirtualMerge: false,
			},
			func(info stage.ImageCommitInfo, data LatestCommitInfoCheckData) {
				Expect(info.Commit).To(Equal(data.CurrentCommit))
				Expect(info.VirtualMerge).To(Equal(false))
			},
		),

		Entry("when current commit is virtual merge",
			LatestCommitInfoCheckData{
				CurrentCommit:               "c69c9771d37c7cb8d8839c79079e83f59b29f343",
				IsCurrentCommitVirtualMerge: true,
			},
			func(info stage.ImageCommitInfo, data LatestCommitInfoCheckData) {
				Expect(info.VirtualMerge).To(Equal(true))
				Expect(info.VirtualMergeFromCommit).To(Equal(constructMergeCommitParentFrom(data.CurrentCommit)))
				Expect(info.VirtualMergeIntoCommit).To(Equal(constructMergeCommitParentInto(data.CurrentCommit)))
			},
		),
	)

	type BaseCommitForPrevBuiltImageCheckData struct {
		BuiltCommitLabel                 string
		BuiltVirtualMergeLabel           string
		BuiltVirtualMergeFromCommitLabel string
		BuiltVirtualMergeIntoCommitLabel string
		CurrentCommit                    string
		IsCurrentCommitVirtualMerge      bool
	}

	DescribeTable("getting base commit from prev built image",
		func(data BaseCommitForPrevBuiltImageCheckData, checkResultFunc func(string, BaseCommitForPrevBuiltImageCheckData)) {
			ctx := context.Background()
			c := NewConveyorStub(stage.VirtualMergeOptions{VirtualMerge: data.IsCurrentCommitVirtualMerge})
			containerBackend := stage.NewContainerBackendStub()

			gitRepo := NewGitRepoStub("own", true, data.CurrentCommit)
			gitMapping.SetGitRepo(gitRepo)

			prevBuiltImage := NewBuiltImageStub("stub", &image.StageDescription{
				Info: &image.Info{
					Labels: map[string]string{
						gitMapping.ImageGitCommitLabel():         data.BuiltCommitLabel,
						gitMapping.VirtualMergeLabel():           data.BuiltVirtualMergeLabel,
						gitMapping.VirtualMergeFromCommitLabel(): data.BuiltVirtualMergeFromCommitLabel,
						gitMapping.VirtualMergeIntoCommitLabel(): data.BuiltVirtualMergeIntoCommitLabel,
					},
				},
			})
			img := stage.NewStageImage(containerBackend, "", prevBuiltImage)

			baseCommit, err := gitMapping.GetBaseCommitForPrevBuiltImage(ctx, c, img)
			Expect(err).To(Succeed())

			checkResultFunc(baseCommit, data)
		},

		Entry("when current commit the same as previously built and is not virtual merge",
			BaseCommitForPrevBuiltImageCheckData{
				BuiltCommitLabel:                 "ae6feb44da273003cfada392c33dfe33748a5e2f",
				BuiltVirtualMergeLabel:           "false",
				BuiltVirtualMergeFromCommitLabel: "",
				BuiltVirtualMergeIntoCommitLabel: "",
				CurrentCommit:                    "ae6feb44da273003cfada392c33dfe33748a5e2f",
				IsCurrentCommitVirtualMerge:      false,
			},
			func(baseCommit string, data BaseCommitForPrevBuiltImageCheckData) {
				Expect(baseCommit).To(Equal(data.CurrentCommit))
			}),

		Entry("when current commit the same as previously built and is virtual merge",
			BaseCommitForPrevBuiltImageCheckData{
				BuiltCommitLabel:                 "a4009816bf7d8f62b2d1cd3331d36f45b3df7d99",
				BuiltVirtualMergeLabel:           "true",
				BuiltVirtualMergeFromCommitLabel: "c1159b9ce05143df3dd6450ee9d749c9642320d2",
				BuiltVirtualMergeIntoCommitLabel: "cdb9a2e6ab3fec90f3ea12d4e4728ec4bf3f74dc",
				CurrentCommit:                    "a4009816bf7d8f62b2d1cd3331d36f45b3df7d99",
				IsCurrentCommitVirtualMerge:      true,
			},
			func(baseCommit string, data BaseCommitForPrevBuiltImageCheckData) {
				Expect(baseCommit).To(Equal(data.CurrentCommit))
			}),

		Entry("when current commit is virtual merge and previous commit is virtual merge",
			BaseCommitForPrevBuiltImageCheckData{
				BuiltCommitLabel:                 "a4009816bf7d8f62b2d1cd3331d36f45b3df7d99",
				BuiltVirtualMergeLabel:           "true",
				BuiltVirtualMergeFromCommitLabel: "c1159b9ce05143df3dd6450ee9d749c9642320d2",
				BuiltVirtualMergeIntoCommitLabel: "cdb9a2e6ab3fec90f3ea12d4e4728ec4bf3f74dc",
				CurrentCommit:                    "d2a341ee66189d0927c47a13541ae1926fad05bb",
				IsCurrentCommitVirtualMerge:      true,
			},
			func(baseCommit string, data BaseCommitForPrevBuiltImageCheckData) {
				Expect(baseCommit).To(Equal(constructVirtualMergeCommit(data.BuiltVirtualMergeFromCommitLabel, data.BuiltVirtualMergeIntoCommitLabel)))
			}),

		Entry("when current commit is virtual merge and previous commit is not virtual merge",
			BaseCommitForPrevBuiltImageCheckData{
				BuiltCommitLabel:                 "ae6feb44da273003cfada392c33dfe33748a5e2f",
				BuiltVirtualMergeLabel:           "false",
				BuiltVirtualMergeFromCommitLabel: "",
				BuiltVirtualMergeIntoCommitLabel: "",
				CurrentCommit:                    "d2a341ee66189d0927c47a13541ae1926fad05bb",
				IsCurrentCommitVirtualMerge:      true,
			},
			func(baseCommit string, data BaseCommitForPrevBuiltImageCheckData) {
				Expect(baseCommit).To(Equal(data.BuiltCommitLabel))
			}),

		Entry("when current commit is not virtual merge and previous commit is virtual merge",
			BaseCommitForPrevBuiltImageCheckData{
				BuiltCommitLabel:                 "a4009816bf7d8f62b2d1cd3331d36f45b3df7d99",
				BuiltVirtualMergeLabel:           "true",
				BuiltVirtualMergeFromCommitLabel: "c1159b9ce05143df3dd6450ee9d749c9642320d2",
				BuiltVirtualMergeIntoCommitLabel: "cdb9a2e6ab3fec90f3ea12d4e4728ec4bf3f74dc",
				CurrentCommit:                    "913c4d1bac6ed0265e910616ad40f426d2e6e625",
				IsCurrentCommitVirtualMerge:      false,
			},
			func(baseCommit string, data BaseCommitForPrevBuiltImageCheckData) {
				Expect(baseCommit).To(Equal(constructVirtualMergeCommit(data.BuiltVirtualMergeFromCommitLabel, data.BuiltVirtualMergeIntoCommitLabel)))
			}),

		Entry("when current commit is not virtual merge and previous commit is not virtual merge",
			BaseCommitForPrevBuiltImageCheckData{
				BuiltCommitLabel:                 "ae6feb44da273003cfada392c33dfe33748a5e2f",
				BuiltVirtualMergeLabel:           "false",
				BuiltVirtualMergeFromCommitLabel: "",
				BuiltVirtualMergeIntoCommitLabel: "",
				CurrentCommit:                    "a4009816bf7d8f62b2d1cd3331d36f45b3df7d99",
				IsCurrentCommitVirtualMerge:      true,
			},
			func(baseCommit string, data BaseCommitForPrevBuiltImageCheckData) {
				Expect(baseCommit).To(Equal(data.BuiltCommitLabel))
			}),
	)
})

type BuiltImageStub struct {
	container_backend.LegacyImageInterface

	name             string
	stageDescription *image.StageDescription
}

func NewBuiltImageStub(name string, stageDescription *image.StageDescription) *BuiltImageStub {
	return &BuiltImageStub{
		name:             name,
		stageDescription: stageDescription,
	}
}

func (img *BuiltImageStub) Name() string {
	return img.name
}

func (img *BuiltImageStub) GetStageDescription() *image.StageDescription {
	return img.stageDescription
}

type ConveyorStub struct {
	stage.Conveyor

	VirtualMergeOptions stage.VirtualMergeOptions
}

func NewConveyorStub(virtualMergeOptions stage.VirtualMergeOptions) *ConveyorStub {
	return &ConveyorStub{
		VirtualMergeOptions: virtualMergeOptions,
	}
}

func (c *ConveyorStub) GetLocalGitRepoVirtualMergeOptions() stage.VirtualMergeOptions {
	return c.VirtualMergeOptions
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

func (gitRepo *GitRepoStub) CreateDetachedMergeCommit(ctx context.Context, fromCommit, toCommit string) (string, error) {
	return constructVirtualMergeCommit(fromCommit, toCommit), nil
}

func (gitRepo *GitRepoStub) GetMergeCommitParents(ctx context.Context, commit string) ([]string, error) {
	return []string{
		constructMergeCommitParentInto(commit),
		constructMergeCommitParentFrom(commit),
	}, nil
}

func constructMergeCommitParentFrom(commit string) string {
	return fmt.Sprintf("%s-from-commit", commit)
}

func constructMergeCommitParentInto(commit string) string {
	return fmt.Sprintf("%s-into-commit", commit)
}

func constructVirtualMergeCommit(fromCommit, toCommit string) string {
	return fmt.Sprintf("%s-%s", fromCommit, toCommit)
}
