package e2e_export_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/utils/docker"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type complexTestOptions struct {
	Platforms  []string
	ImageNames []string
}

var _ = Describe("Complex converge", Label("e2e", "converge", "complex"), func() {
	DescribeTable("should succeed and export images",
		func(opts complexTestOptions) {
			By("initializating")
			setupEnv()
			repoDirname := "repo0"
			By("state0: starting")
			{
				fixtureRelPath := "complex/state0"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By("state0: running export")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				imageTemplate := `werf-export-%image%`
				tag := utils.GetRandomString(10)
				imageName := fmt.Sprintf("%s/%s:%s", SuiteData.RegistryLocalAddress, imageTemplate, tag)

				exportArgs := getExportArgs(imageName, commonTestOptions{
					Platforms: opts.Platforms,
				})

				exportOut := werfProject.Export(&werf.ExportOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: exportArgs,
					},
				})
				for _, imageName := range opts.ImageNames {
					Expect(exportOut).To(ContainSubstring(fmt.Sprintf("Exporting image %s", imageName)))
				}

				By("state0: checking result")

				var lookUpImages []string
				for _, imageName := range opts.ImageNames {
					lookUpImages = append(lookUpImages, fmt.Sprintf("werf-export-%s", imageName))
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				ok := docker.LookUpForRepositories(ctx, SuiteData.RegistryLocalAddress, lookUpImages)
				Expect(ok).To(BeTrue())
				for _, imageName := range lookUpImages {
					ok := docker.LookUpForTag(ctx, SuiteData.RegistryLocalAddress, imageName, tag)
					Expect(ok).To(BeTrue())
				}
			}
		},
		Entry("base", complexTestOptions{
			ImageNames: []string{
				"backend",
				"frombackend",
				"frombackend2",
			},
		}),
		Entry("multiplatform", complexTestOptions{
			Platforms: []string{
				"linux/amd64",
			},
			ImageNames: []string{
				"backend",
				"frombackend",
				"frombackend2",
			},
		}),
	)
})
