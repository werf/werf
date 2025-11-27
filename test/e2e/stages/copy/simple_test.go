package e2e_stages_copy_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/werf"
)

type simpleTestOptions struct {
	Platforms []string
	All       bool
}

var _ = Describe("Simple stages copy", Label("e2e", "stages copy", "simple"), func() {
	DescribeTable("should succeed and copy stages",
		func(ctx SpecContext, opts simpleTestOptions) {
			By("initializing")
			{
				setupEnv()
			}

			By("state0: starting")
			{
				repoDirName := "repo"
				fixtureRelPath := "simple"

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
				Expect(buildOut).NotTo(ContainSubstring("Use previously built image"))

				By("state0: execute stages copy")
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
				buildOut = werfProject.Build(ctx, &werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: false,
						ExtraArgs:  []string{"--require-built-images", "--repo", SuiteData.WerfToAddr},
					},
				})
				Expect(buildOut).To(ContainSubstring("Use previously built image"))
			}
		},
		Entry("with copy all stages", simpleTestOptions{
			Platforms: []string{
				"linux/amd64",
			},
			All: true,
		}),
		//Entry("with copy only current build stages", simpleTestOptions{
		//	Platforms: []string{
		//		"linux/amd64",
		//	},
		//	All: false,
		//}),
		//Entry("with copy all stages and multiplatform images ", simpleTestOptions{
		//	Platforms: []string{
		//		"linux/amd64",
		//		"linux/arm64",
		//	},
		//	All: true,
		//}),
		//Entry("with copy only current build stages and multiplatform images", simpleTestOptions{
		//	Platforms: []string{
		//		"linux/amd64",
		//		"linux/arm64",
		//	},
		//	All: false,
		//}),
	)
})
