package common_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type simpleTestOptions struct {
	setupEnvOptions
}

var _ = Describe("build and mutate image spec", Label("integration", "build", "mutate spec config"), func() {
	DescribeTable("should succeed and produce expected image",
		func(testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By(fmt.Sprintf("%s: starting", testOpts.State))
			{
				repoDirname := "repo0"
				fixtureRelPath := "simple"
				buildReportName := "report0.json"

				By(fmt.Sprintf("%s: preparing test repo", testOpts.State))
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By(fmt.Sprintf("%s: building images", testOpts.State))
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut, buildReport := werfProject.BuildWithReport(SuiteData.GetBuildReportPath(buildReportName), nil)
				Expect(buildOut).To(ContainSubstring("Building stage"))

				By("getting built images metadata")
				for imageName := range buildReport.Images {
					contRuntime.Pull(buildReport.Images[imageName].DockerImageName)
					inspectOfImage := contRuntime.GetImageInspect(buildReport.Images[imageName].DockerImageName)
					imgCfg := inspectOfImage.Config

					By("checking image metadata")
					Expect(imgCfg.Env).Should(ContainElement("TEST=test"))
					Expect(imgCfg.Env).ShouldNot(ContainElement(ContainSubstring("PATH")))
					Expect(imgCfg.Volumes).Should(HaveKey("/var/lib/test/data"))
					Expect(imgCfg.Entrypoint).Should(ContainElement("/bin/sh"))
					Expect(imgCfg.Cmd).Should(ContainElement("echo"))
					Expect(imgCfg.Labels).Should(HaveKey("Test"))
					Expect(imgCfg.Labels).Should(HaveKey("Test_Global"))
					Expect(imgCfg.User).Should(Equal("root"))
				}
			}
		},
		Entry("without local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})
