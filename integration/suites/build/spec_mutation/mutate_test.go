package common_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/thirdparty/contruntime/manifest"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type simpleTestOptions struct {
	setupEnvOptions
}

var _ = Describe("build and mutate image spec", Label("integration", "build", "mutate spec config"), func() {
	DescribeTable("should succeed and produce expected image",
		func(ctx SpecContext, testOpts simpleTestOptions) {
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
				fixtureRelPath := "complex"
				buildReportName := "report0.json"

				By(fmt.Sprintf("%s: preparing test repo", testOpts.State))
				SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

				By(fmt.Sprintf("%s: building images", testOpts.State))
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				SuiteData.Stubs.SetEnv("WERF_LOG_DEBUG", "true")
				// SuiteData.Stubs.SetEnv("WERF_PARALLEL", "false")

				buildOut, buildReport := werfProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)
				Expect(buildOut).To(ContainSubstring("Building stage"))

				By("getting built images metadata")
				for imageName := range buildReport.Images {

					if testOpts.WithLocalRepo {
						contRuntime.Pull(ctx, buildReport.Images[imageName].DockerImageName)
					}

					inspectOfImage := contRuntime.GetImageInspect(ctx, buildReport.Images[imageName].DockerImageName)
					imgCfg := inspectOfImage.Config

					By("checking image metadata")
					switch imageName {
					case "basic-test":

						Expect(inspectOfImage.Author).Should(Equal("globalAuthor"))

						Expect(imgCfg.Env).Should(ContainElement("ADD=me"))
						Expect(imgCfg.Env).Should(ContainElement("ADD_ANOTHER=me"))
						Expect(imgCfg.Env).Should(ContainElement("PATH=/usr/bin:/add/path"))
						Expect(imgCfg.Env).ShouldNot(ContainElement("APP_ENV=test"))
						Expect(imgCfg.Env).ShouldNot(ContainElement("APP_VERSION=0.0.1"))
						Expect(imgCfg.Env).ShouldNot(ContainElement("REMOVE=ME"))

						Expect(imgCfg.Volumes).ShouldNot(HaveKey("/home/app/data"))
						Expect(imgCfg.Volumes).Should(HaveKey("/test/volume"))
						Expect(imgCfg.Volumes).Should(HaveKey("/second/test/volume"))
						Expect(imgCfg.Volumes).ShouldNot(HaveKey("/home/remove/me"))

						Expect(imgCfg.Cmd).Should(ContainElement("/bin/sh"))
						Expect(imgCfg.Entrypoint).Should(ContainElement("test"))

						Expect(imgCfg.Labels).Should(HaveKey("maintainer"))
						Expect(imgCfg.Labels).Should(HaveKey("save"))
						Expect(imgCfg.Labels).Should(HaveKey("test"))
						Expect(imgCfg.Labels).Should(HaveKey("werf"))
						Expect(imgCfg.Labels).Should(HaveKey("global_label"))
						Expect(imgCfg.Labels).Should(HaveKey("werf.io/parent-stage-id"))
						Expect(imgCfg.Labels).ShouldNot(HaveKey("pleaseremove"))
						Expect(imgCfg.Labels).ShouldNot(HaveKey("remove.completely"))
						Expect(imgCfg.Labels).ShouldNot(HaveKey("remove.all"))

						Expect(imgCfg.User).Should(Equal("testuser"))

						Expect(imgCfg.ExposedPorts).Should(Equal(manifest.Schema2PortSet{"99": {}}))
						Expect(imgCfg.ExposedPorts).ShouldNot(HaveKey("1234/tcp"))

						Expect(imgCfg.WorkingDir).Should(Equal("/test/work"))

						Expect(imgCfg.StopSignal).Should(Equal("SIGINT"))

					case "clean-test":

						Expect(inspectOfImage.Author).Should(Equal("globalAuthor"))

						Expect(imgCfg.Env).Should(BeEmpty())

						Expect(imgCfg.Volumes).ShouldNot(BeEmpty())

						Expect(imgCfg.Cmd).Should(BeEmpty())
						Expect(imgCfg.Entrypoint).Should(BeEmpty())

						Expect(imgCfg.Labels).Should(HaveKey("global_label"))

						Expect(imgCfg.User).Should(Equal(""))

						Expect(imgCfg.ExposedPorts).Should(Equal(manifest.Schema2PortSet{"": {}}))

						Expect(imgCfg.WorkingDir).Should(Equal(""))

					case "cmd-test":
						Expect(inspectOfImage.Author).Should(Equal("globalAuthor"))
						Expect(imgCfg.Cmd).Should(BeEmpty())
						Expect(imgCfg.Entrypoint).Should(ContainElement("/bin/test"))
					}
				}
			}
		},
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		FEntry("without local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}, FlakeAttempts(15)),
	)
})
