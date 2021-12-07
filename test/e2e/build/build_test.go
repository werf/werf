package e2e_build_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/contruntime"
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
	"github.com/werf/werf/test/pkg/werf"
)

var _ = Describe("Build", func() {
	DescribeTable("should succeed and produce expected image",
		func(withLocalRepo bool, buildahMode string) {
			By("initializing")
			setupEnv(withLocalRepo, buildahMode)
			contRuntime, err := contruntime.NewContainerRuntime(buildahMode)
			if err == contruntime.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("state0: starting")
			{
				repoDirname := "repo0"
				fixtureRelPath := "state0"
				buildReportName := "report0.json"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By("state0: building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut, buildReport := werfProject.BuildWithReport(SuiteData.GetBuildReportPath(buildReportName))
				Expect(buildOut).To(ContainSubstring("Building stage"))
				Expect(buildOut).NotTo(ContainSubstring("Use cache image"))

				By("state0: rebuilding same images")
				Expect(werfProject.Build()).To(And(
					ContainSubstring("Use cache image"),
					Not(ContainSubstring("Building stage")),
				))

				By(`state0: getting built "dockerfile" image metadata`)
				config := contRuntime.GetImageInspectConfig(buildReport.Images["dockerfile"].DockerImageName)

				By("state0: checking built images metadata")
				Expect(config.Env).To(ContainElement("COMPOSED_ENV=env-was_changed"))
				Expect(config.Labels).To(HaveKeyWithValue("COMPOSED_LABEL", "label-was_changed"))
				Expect(config.Shell).To(ContainElements("/bin/sh", "-c"))
				Expect(config.User).To(Equal("0:0"))
				Expect(config.WorkingDir).To(Equal("/"))
				Expect(config.Entrypoint).To(ContainElements("sh", "-ec"))
				Expect(config.Cmd).To(ContainElement("tail -f /dev/null"))
				Expect(config.Volumes).To(HaveKey("/persistent"))
				Expect(config.OnBuild).To(ContainElement("RUN echo onbuild"))
				Expect(config.StopSignal).To(Equal("SIGTERM"))
				Expect(config.ExposedPorts).To(HaveKey(manifest.Schema2Port("80/tcp")))
				Expect(config.Healthcheck.Test).To(ContainElements("CMD-SHELL", "echo healthcheck"))

				By("state0: checking built images content")
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["dockerfile"].DockerImageName,
					"test -f /app/added/file2",
					"test -f /app/copied/file1",
					"test -f /app/copied/file2",
					"echo 'file1content' | diff /app/added/file1 -",
					"echo 'file2content' | diff /app/added/file2 -",
					"echo 'file1content' | diff /app/copied/file1 -",
					"echo 'file2content' | diff /app/copied/file2 -",
					"test -f /helloworld.tgz",
					"tar xOf /helloworld.tgz | grep 'Hello World!'",
					"test -f /created-by-run",
					"test -d /persistent/should-exist-in-volume",
				)
			}

			By("state1: starting")
			{
				repoDirname := "repo0"
				fixtureRelPath := "state1"
				buildReportName := "report1.json"

				By("state1: changing files in test repo")
				SuiteData.UpdateTestRepo(repoDirname, fixtureRelPath)

				By("state1: building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut, buildReport := werfProject.BuildWithReport(SuiteData.GetBuildReportPath(buildReportName))
				Expect(buildOut).To(ContainSubstring("Building stage"))

				By("state1: rebuilding same images")
				Expect(werfProject.Build()).To(And(
					ContainSubstring("Use cache image"),
					Not(ContainSubstring("Building stage")),
				))

				By(`state1: getting built "dockerfile" image metadata`)
				config := contRuntime.GetImageInspectConfig(buildReport.Images["dockerfile"].DockerImageName)

				By("state1: checking built images metadata")
				Expect(config.Env).To(ContainElement("COMPOSED_ENV=env-was_changed-state1"))
				Expect(config.Labels).To(HaveKeyWithValue("COMPOSED_LABEL", "label-was_changed-state1"))

				By("state1: checking built images content")
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["dockerfile"].DockerImageName,
					"test -f /app/added/file1",
					"test -f /app/added/file3",
					"test -f /app/copied/file1",
					"test -f /app/copied/file3",
					"! test -f /app/added/file2",
					"! test -f /app/copied/file2",
					"echo 'file1content-state1' | diff /app/added/file1 -",
					"echo 'file3content-state1' | diff /app/added/file3 -",
					"echo 'file1content-state1' | diff /app/copied/file1 -",
					"echo 'file3content-state1' | diff /app/copied/file3 -",
					"! test -f /helloworld.tgz",
					"test -f /created-by-run-state1",
				)
			}
		},
		Entry(
			"without repo using Docker",
			false, "docker",
		),
		Entry(
			"with local repo using Docker",
			true, "docker",
		),
		Entry(
			"with local repo using Native Buildah with rootless isolation",
			true, "native-rootless",
		),
		Entry(
			"with local repo using Native Buildah with chroot isolation",
			true, "native-chroot",
		),
		Entry(
			"with local repo using Docker-With-Fuse Buildah with chroot isolation",
			true, "docker-with-fuse",
		),
		// TODO: uncomment when buildah allows building without --repo flag
		// Entry("without repo using Native Rootless Buildah", false, contruntime.NativeRootlessBuildah),
		// Entry("without repo using Docker-With-Fuse Buildah", false, contruntime.DockerWithFuseBuildah),
	)
})

func setupEnv(withLocalRepo bool, buildahMode string) {
	SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", buildahMode)

	if withLocalRepo && buildahMode == "docker" {
		SuiteData.Stubs.SetEnv("WERF_REPO", strings.Join([]string{SuiteData.RegistryLocalAddress, SuiteData.ProjectName}, "/"))
	} else if withLocalRepo {
		SuiteData.Stubs.SetEnv("WERF_REPO", strings.Join([]string{SuiteData.RegistryInternalAddress, SuiteData.ProjectName}, "/"))
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_REPO")
	}

	if withLocalRepo {
		SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
		SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_INSECURE_REGISTRY")
		SuiteData.Stubs.UnsetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY")
	}
}
