package e2e_stages_copy_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/werf"
)

type complexTestOptions struct {
	Platforms []string
	All       bool
}

var _ = Describe("Complex stages copy", Label("e2e", "stages copy", "complex"), func() {
	DescribeTable("should succeed and copy stages",
		func(ctx SpecContext, opts complexTestOptions) {
			By("initializing")
			{
				setupEnv()
			}

			By("state0: starting")
			{
				repoDirName := "repo"
				fixtureRelPath := "complex/state0"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirName, fixtureRelPath)

				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))

				By("state0: building images")
				buildOut := werfProject.Build(ctx, &werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: false,
						ExtraArgs:  []string{"--repo", SuiteData.WerfFromAddr},
					},
				})
				Expect(buildOut).To(ContainSubstring("Building stage"))

				By("state0: execute stages copy to archive")
				stagesCopyArgs := getStagesCopyArgs(SuiteData.WerfFromAddr, SuiteData.WerfToAddr, commonTestOptions{
					All: &opts.All,
				})
				stagesCopyOut := werfProject.StagesCopy(ctx, &werf.StagesCopyOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: false,
						ExtraArgs:  stagesCopyArgs,
					},
				})
				By("state0: check stages copy output")
				Expect(stagesCopyOut).To(ContainSubstring(fmt.Sprintf("Copy stages")))
				Expect(stagesCopyOut).To(ContainSubstring(fmt.Sprintf("From: %s", SuiteData.WerfFromAddr)))
				Expect(stagesCopyOut).To(ContainSubstring(fmt.Sprintf("To: %s", SuiteData.WerfToAddr)))

				By("state0: check that images were built successfully")
				werfProject.Build(ctx, &werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: false,
						ExtraArgs:  []string{"--require-built-images", "--repo", SuiteData.WerfToAddr},
					},
				})
				Expect(buildOut).To(ContainSubstring("Building stage"))
				Expect(buildOut).NotTo(ContainSubstring("Use previously build image"))
			}
			By("state1: with changing repo state")
			{
				repoDirName := "repo"
				fixtureRelPath := "complex/state1"

				By("state1: changing files in test repo")
				SuiteData.UpdateTestRepo(ctx, repoDirName, fixtureRelPath)

				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))

				By("state1: building images")
				buildOut := werfProject.Build(ctx, &werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: false,
						ExtraArgs:  []string{"--repo", SuiteData.WerfFromAddr},
					},
				})
				Expect(buildOut).To(ContainSubstring("Building stage"))

				By("state1: execute stages copy")
				stagesCopyArgs := getStagesCopyArgs(SuiteData.WerfFromAddr, SuiteData.WerfArchiveAddr, commonTestOptions{
					All: &opts.All,
				})
				stagesCopyOut := werfProject.StagesCopy(ctx, &werf.StagesCopyOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: false,
						ExtraArgs:  stagesCopyArgs,
					},
				})

				By("state1: check stages copy to archive output")
				Expect(stagesCopyOut).To(ContainSubstring(fmt.Sprintf("Copy stages")))
				Expect(stagesCopyOut).To(ContainSubstring(fmt.Sprintf("From: %s", SuiteData.WerfFromAddr)))
				Expect(stagesCopyOut).To(ContainSubstring(fmt.Sprintf("To: %s", SuiteData.WerfArchiveAddr)))

				stagesCopyArgs = getStagesCopyArgs(SuiteData.WerfArchiveAddr, SuiteData.WerfToAddr, commonTestOptions{})
				stagesCopyOut = werfProject.StagesCopy(ctx, &werf.StagesCopyOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: false,
						ExtraArgs:  stagesCopyArgs,
					},
				})

				By("state1: check stages copy to container registry output")
				Expect(stagesCopyOut).To(ContainSubstring(fmt.Sprintf("Copy stages")))
				Expect(stagesCopyOut).To(ContainSubstring(fmt.Sprintf("From: %s", SuiteData.WerfArchiveAddr)))
				Expect(stagesCopyOut).To(ContainSubstring(fmt.Sprintf("To: %s", SuiteData.WerfToAddr)))

				By("state1: check that images were built successfully")
				var shouldFail bool
				if opts.All {
					shouldFail = false
				} else {
					shouldFail = true
				}

				werfProject.Build(ctx, &werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: shouldFail,
						ExtraArgs:  []string{"--require-built-images", "--repo", SuiteData.WerfToAddr},
					},
				})
				Expect(buildOut).To(ContainSubstring("Building stage"))
				Expect(buildOut).NotTo(ContainSubstring("Use previously build image"))
			}
		},
		Entry("with copy all stages", complexTestOptions{
			Platforms: []string{
				"linux/amd64",
			},
			All: true,
		}),
		Entry("with copy only current build stages", complexTestOptions{
			Platforms: []string{
				"linux/amd64",
			},
			All: false,
		}),
		Entry("with copy all stages and multiplatform images ", complexTestOptions{
			Platforms: []string{
				"linux/amd64",
				"linux/arm64",
			},
			All: true,
		}),
		Entry("with copy only current build stages and multiplatform images", complexTestOptions{
			Platforms: []string{
				"linux/amd64",
				"linux/arm64",
			},
			All: false,
		}),
	)
})
