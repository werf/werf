package e2e_export_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/werf"
)

type simpleTestOptions struct {
	ExtraArgs []string
}

var _ = Describe("Simple export", Label("e2e", "export", "simple"), func() {
	DescribeTable("should succeed and export images",
		func(opts simpleTestOptions) {
			SuiteData.WerfRepo = strings.Join([]string{SuiteData.RegistryLocalAddress, SuiteData.ProjectName}, "/")
			SuiteData.Stubs.SetEnv("WERF_REPO", SuiteData.WerfRepo)
			SuiteData.Stubs.SetEnv("DOCKER_BUILDKIT", "1")
			By("initializating")
			{
				repoDirname := "repo"
				fixtureRelPath := "simple"

				By("preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By("running export")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				exportOut := werfProject.Export(&werf.ExportOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: append([]string{
							"--tag",
							fmt.Sprintf("%s/test", SuiteData.RegistryLocalAddress),
						}, opts.ExtraArgs...)},
				})
				Expect(exportOut).To(ContainSubstring("Exporting image..."))
			}
		},
		Entry("Export single platform image", simpleTestOptions{
			ExtraArgs: []string{},
		}),
		Entry("Export multiplatform image", simpleTestOptions{
			ExtraArgs: []string{
				"--platform",
				"linux/amd64,linux/arm64",
			},
		},
		),
	)
})
