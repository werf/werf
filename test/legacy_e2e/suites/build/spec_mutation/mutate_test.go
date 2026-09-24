package common_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/report"
	"github.com/werf/werf/v3/test/pkg/thirdparty/contruntime/manifest"
	"github.com/werf/werf/v3/test/pkg/werf"
)

type simpleTestOptions struct {
	setupEnvOptions
}

var _ = Describe("build and mutate image spec", Label("integration", "build", "mutate spec config"), func() {
	DescribeTable("should succeed and produce expected image",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contRuntime := contback.NewContainerBackend(testOpts.ContainerBackendMode)

			basicPortsMatcher := Equal(manifest.Schema2PortSet{"99": {}})
			cleanPortsMatcher := Equal(manifest.Schema2PortSet{"": {}})
			if testOpts.ContainerBackendMode == "docker" {
				basicPortsMatcher = Or(basicPortsMatcher, Equal(manifest.Schema2PortSet{"99/tcp": {}}))
				cleanPortsMatcher = Or(cleanPortsMatcher, Equal(manifest.Schema2PortSet{"invalid port": {}}))
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
				reportProject := report.NewProjectWithReport(werfProject)

				buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)
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

						Expect(imgCfg.Volumes).Should(HaveKey("/home/app/data"))
						Expect(imgCfg.Volumes).Should(HaveKey("/test/volume"))
						Expect(imgCfg.Volumes).Should(HaveKey("/second/test/volume"))
						Expect(imgCfg.Volumes).ShouldNot(HaveKey("/home/remove/me"))

						Expect([]string(imgCfg.Cmd)).Should(Equal([]string{"/bin/sh", "-c", "echo cmd"}))
						Expect([]string(imgCfg.Entrypoint)).Should(Equal([]string{"command", "param1", "param2"}))

						Expect(imgCfg.Labels).Should(HaveKey("maintainer"))
						Expect(imgCfg.Labels).Should(HaveKey("save"))
						Expect(imgCfg.Labels).Should(HaveKeyWithValue("test", "test_value"))
						Expect(imgCfg.Labels).Should(HaveKey("werf"))
						Expect(imgCfg.Labels).Should(HaveKey("global_label"))
						Expect(imgCfg.Labels).Should(HaveKey("werf.io/parent-stage-id"))
						Expect(imgCfg.Labels).ShouldNot(HaveKey("pleaseremove"))
						Expect(imgCfg.Labels).ShouldNot(HaveKey("remove.completely"))
						Expect(imgCfg.Labels).ShouldNot(HaveKey("remove.all"))

						Expect(imgCfg.User).Should(Equal("testuser"))

						Expect(imgCfg.ExposedPorts).Should(basicPortsMatcher)
						Expect(imgCfg.ExposedPorts).ShouldNot(HaveKey("1234/tcp"))

						Expect(imgCfg.WorkingDir).Should(Equal("/test/work"))

						Expect(imgCfg.StopSignal).Should(Equal("SIGINT"))

						Expect(imgCfg.Healthcheck).ShouldNot(BeNil())
						Expect(imgCfg.Healthcheck.Test).Should(Equal([]string{"curl -f http://localhost/ || exit 1"}))
						Expect(imgCfg.Healthcheck.Retries).Should(Equal(3))

					case "clean-test":

						Expect(inspectOfImage.Author).Should(Equal("globalAuthor"))

						Expect(imgCfg.Env).Should(BeEmpty())

						Expect(imgCfg.Volumes).Should(BeEmpty())

						Expect(imgCfg.Cmd).Should(BeEmpty())
						Expect(imgCfg.Entrypoint).Should(BeEmpty())

						Expect(imgCfg.Labels).Should(HaveKey("global_label"))

						Expect(imgCfg.User).Should(Equal(""))

						Expect(imgCfg.ExposedPorts).Should(cleanPortsMatcher)

						Expect(imgCfg.WorkingDir).Should(Equal(""))

					case "cmd-test":
						Expect(inspectOfImage.Author).Should(Equal("globalAuthor"))
						Expect(imgCfg.Cmd).Should(BeEmpty())
						Expect(imgCfg.Entrypoint).Should(ContainElement("/bin/test"))
					}
				}
			}
		},
		// Docker
		Entry("with repo using Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("without repo using Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
		// Buildah rootless
		Entry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("without local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
		// Buildah chroot
		Entry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})
