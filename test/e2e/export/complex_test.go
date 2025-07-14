package e2e_export_test

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type complexTestOptions struct {
	Platforms  []string
	ImageNames []string
}

var _ = Describe("Complex converge", Label("e2e", "converge", "complex"), func() {
	DescribeTable("should succeed and export images",
		func(ctx SpecContext, opts complexTestOptions) {
			By("initializating")
			setupEnv()
			repoDirname := "repo0"
			By("state0: starting")
			{
				fixtureRelPath := "complex/state0"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

				By("state0: running export")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				imageTemplate := `werf-export-%image%`
				tag := utils.GetRandomString(10)
				imageName := fmt.Sprintf("%s/%s:%s", SuiteData.RegistryLocalAddress, imageTemplate, tag)

				exportArgs := getExportArgs(imageName, commonTestOptions{
					Platforms: opts.Platforms,
				})

				exportOut := werfProject.Export(ctx, &werf.ExportOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: exportArgs,
					},
				})
				for _, imageName := range opts.ImageNames {
					Expect(exportOut).To(ContainSubstring(fmt.Sprintf("Exporting image %s", imageName)))
				}

				By("state0: checking result")
				_, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()

				for _, imageName := range opts.ImageNames {
					i := fmt.Sprintf("%s/werf-export-%s:%s", SuiteData.RegistryLocalAddress, imageName, tag)
					ref, err := name.ParseReference(i)
					Expect(err).ShouldNot(HaveOccurred())

					_, err = remote.Head(ref)
					Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("image %s should be exported", imageName))

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
