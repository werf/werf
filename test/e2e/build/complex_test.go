package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/contback"
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
	"github.com/werf/werf/test/pkg/werf"
)

type complexTestOptions struct {
	BuildahMode                 string
	WithLocalRepo               bool
	WithStagedDockerfileBuilder bool
}

var _ = Describe("Complex build", Label("e2e", "build", "complex"), func() {
	DescribeTable("should succeed and produce expected image",
		func(testOpts complexTestOptions) {
			By("initializing")
			setupEnv(setupEnvOptions{
				BuildahMode:               testOpts.BuildahMode,
				WithLocalRepo:             testOpts.WithLocalRepo,
				WithForceStagedDockerfile: testOpts.WithStagedDockerfileBuilder,
			})
			contRuntime, err := contback.NewContainerBackend(testOpts.BuildahMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("state0: starting")
			{
				repoDirname := "repo0"
				fixtureRelPath := "complex/state0"
				buildReportName := "report0.json"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By("state0: building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut, buildReport := werfProject.BuildWithReport(SuiteData.GetBuildReportPath(buildReportName), nil)
				Expect(buildOut).To(ContainSubstring("Building stage"))
				Expect(buildOut).NotTo(ContainSubstring("Use cache image"))

				By("state0: rebuilding same images")
				Expect(werfProject.Build(nil)).To(And(
					ContainSubstring("Use cache image"),
					Not(ContainSubstring("Building stage")),
				))

				By(`state0: getting built images metadata`)
				dockerfileImgCfg := contRuntime.GetImageInspectConfig(buildReport.Images["dockerfile"].DockerImageName)
				stapelShellImgCfg := contRuntime.GetImageInspectConfig(buildReport.Images["stapel-shell"].DockerImageName)

				By(`state0: checking "dockerfile" image metadata`)
				Expect(dockerfileImgCfg.Env).To(ContainElement("COMPOSED_ENV=env-was_changed"))
				Expect(dockerfileImgCfg.Labels).To(HaveKeyWithValue("COMPOSED_LABEL", "label-was_changed"))
				Expect(dockerfileImgCfg.Shell).To(ContainElements("/bin/sh", "-c"))
				Expect(dockerfileImgCfg.User).To(Equal("0:0"))
				Expect(dockerfileImgCfg.WorkingDir).To(Equal("/"))
				Expect(dockerfileImgCfg.Entrypoint).To(ContainElements("sh", "-ec"))
				Expect(dockerfileImgCfg.Cmd).To(ContainElement("tail -f /dev/null"))
				Expect(dockerfileImgCfg.Volumes).To(HaveKey("/volume10"))
				Expect(dockerfileImgCfg.OnBuild).To(ContainElement("RUN echo onbuild"))
				Expect(dockerfileImgCfg.StopSignal).To(Equal("SIGTERM"))
				Expect(dockerfileImgCfg.ExposedPorts).To(HaveKey(manifest.Schema2Port("8000/tcp")))
				Expect(dockerfileImgCfg.Healthcheck.Test).To(ContainElements("CMD-SHELL", "echo healthcheck10"))

				By(`state0: checking "stapel-shell" image metadata`)
				Expect(stapelShellImgCfg.Env).To(ContainElement("ENV_NAME=env-value"))
				Expect(stapelShellImgCfg.Labels).To(HaveKeyWithValue("LABEL_NAME", "label-value"))
				Expect(stapelShellImgCfg.User).To(Equal("0:0"))
				Expect(stapelShellImgCfg.WorkingDir).To(Equal("/app"))
				Expect(stapelShellImgCfg.Entrypoint).To(ContainElements("sh", "-ec"))
				Expect(stapelShellImgCfg.Cmd).To(ContainElement("sleep infinity"))
				Expect(stapelShellImgCfg.Volumes).To(HaveKey("/volume20"))
				Expect(stapelShellImgCfg.ExposedPorts).To(HaveKey(manifest.Schema2Port("8010/tcp")))
				Expect(stapelShellImgCfg.Healthcheck.Test).To(ContainElements("CMD-SHELL", "echo healthcheck20"))

				By(`state0: checking "dockerfile" image content`)
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["dockerfile"].DockerImageName,
					"test -f /app/added/file1",
					"echo 'file1content' | diff /app/added/file1 -",

					"test -f /app/added/file2",
					"echo 'file2content' | diff /app/added/file2 -",

					"test -f /app/copied/file1",
					"echo 'file1content' | diff /app/copied/file1 -",

					"test -f /app/copied/file2",
					"echo 'file2content' | diff /app/copied/file2 -",

					"test -f /helloworld.tgz",
					"tar xOf /helloworld.tgz | grep -qF 'Hello World!'",

					"test -f /created-by-run-state0",

					"test -d /volume10/should-exist-in-volume",

					"! test -e /tmpfs",

					"! test -e /bind",

					"! test -e /bind-from-builder",
				)

				By(`state0: checking "stapel-shell" image content`)
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["stapel-shell"].DockerImageName,
					"test -f /app/README.md",
					"stat -c %u:%g /app/README.md | diff <(echo 1050:1051) -",
					"grep -qF 'https://cloud.google.com/appengine/docs/go/#Go_tools' /app/README.md",

					"test -f /app/static/index.html",
					"stat -c %u:%g /app/static/index.html | diff <(echo 1050:1051) -",
					"grep -qF '<title>Hello, world</title>' /app/static/index.html",

					"test -f /app/static/style.css",
					"stat -c %u:%g /app/static/style.css | diff <(echo 1050:1051) -",
					"grep -qF 'text-align: center;' /app/static/style.css",

					"! test -e /app/app.go",

					"! test -e /app/static/script.js",

					"test -f /triggered-stages",
					"stat -c %u:%g /triggered-stages | diff <(echo 0:0) -",
					"echo 'beforeInstall\ninstall\nbeforeSetup\nsetup' | diff /triggered-stages -",

					"! test -e /tmp_dir/file",

					"test -f /basedir/file",
					"stat -c %u:%g /basedir/file | diff <(echo 0:0) -",
					"echo 'content' | diff /basedir/file -",

					"test -f /basedir-imported/file",
					"stat -c %u:%g /basedir-imported/file | diff <(echo 1060:1061) -",
					"echo 'content' | diff /basedir-imported/file -",
				)
			}

			By("state1: starting")
			{
				repoDirname := "repo0"
				fixtureRelPath := "complex/state1"
				buildReportName := "report1.json"

				By("state1: changing files in test repo")
				SuiteData.UpdateTestRepo(repoDirname, fixtureRelPath)

				By("state1: building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut, buildReport := werfProject.BuildWithReport(SuiteData.GetBuildReportPath(buildReportName), nil)
				Expect(buildOut).To(ContainSubstring("Building stage"))

				By("state1: rebuilding same images")
				Expect(werfProject.Build(nil)).To(And(
					ContainSubstring("Use cache image"),
					Not(ContainSubstring("Building stage")),
				))

				By(`state1: getting built images metadata`)
				dockerfileImgCfg := contRuntime.GetImageInspectConfig(buildReport.Images["dockerfile"].DockerImageName)
				stapelShellImgCfg := contRuntime.GetImageInspectConfig(buildReport.Images["stapel-shell"].DockerImageName)

				By(`state1: checking "dockerfile" image metadata`)
				Expect(dockerfileImgCfg.Env).To(ContainElement("COMPOSED_ENV=env-was_changed-state1"))
				Expect(dockerfileImgCfg.Labels).To(HaveKeyWithValue("COMPOSED_LABEL", "label-was_changed-state1"))

				By(`state1: checking "stapel-shell" image metadata`)
				Expect(stapelShellImgCfg.Volumes).ToNot(HaveKey("/volume20"))

				By(`state1: checking "dockerfile" image content`)
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["dockerfile"].DockerImageName,
					"test -f /app/added/file1",
					"echo 'file1content-state1' | diff /app/added/file1 -",

					"! test -e /app/added/file2",

					"test -f /app/added/file3",
					"echo 'file3content-state1' | diff /app/added/file3 -",

					"test -f /app/copied/file1",
					"echo 'file1content-state1' | diff /app/copied/file1 -",

					"! test -e /app/copied/file2",

					"test -f /app/copied/file3",
					"echo 'file3content-state1' | diff /app/copied/file3 -",

					"! test -e /helloworld.tgz",

					"test -f /created-by-run-state1",

					"! test -e /tmpfs",

					"! test -e /bind",

					"! test -e /bind-from-builder",
				)

				By(`state1: checking "stapel-shell" image content`)
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["stapel-shell"].DockerImageName,
					"test -f /app/README.md",
					"stat -c %u:%g /app/README.md | diff <(echo 1050:1051) -",
					"grep -qF 'https://cloud.google.com/sdk/' /app/README.md",

					"test -f /app/static/index.html",
					"stat -c %u:%g /app/static/index.html | diff <(echo 1050:1051) -",
					"grep -qF '<title>Hello, world</title>' /app/static/index.html",

					"! test -e /app/static/style.css",

					"test -f /app/app.go",
					"stat -c %u:%g /app/app.go | diff <(echo 1050:1051) -",
					"grep -qF 'package hello' /app/app.go",

					"! test -e /app/static/script.js",

					"test -f /triggered-stages",
					"stat -c %u:%g /triggered-stages | diff <(echo 0:0) -",
					"echo 'beforeInstall\ninstall\nbeforeSetup\nsetup' | diff /triggered-stages -",

					"! test -e /tmp_dir/file",

					"test -f /basedir/file",
					"stat -c %u:%g /basedir/file | diff <(echo 0:0) -",
					"echo 'content' | diff /basedir/file -",

					"test -f /basedir-imported/file",
					"stat -c %u:%g /basedir-imported/file | diff <(echo 1060:1061) -",
					"echo 'content' | diff /basedir-imported/file -",
				)
			}
		},
		Entry("without repo using Vanilla Docker", complexTestOptions{
			BuildahMode:                 "vanilla-docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}),
		Entry("with local repo using Vanilla Docker", complexTestOptions{
			BuildahMode:                 "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
		Entry("without repo using BuildKit Docker", complexTestOptions{
			BuildahMode:                 "buildkit-docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}),
		Entry("with local repo using BuildKit Docker", complexTestOptions{
			BuildahMode:                 "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
		Entry("with local repo using Native Buildah with rootless isolation", complexTestOptions{
			BuildahMode:                 "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
		Entry("with local repo using Native Buildah with chroot isolation", complexTestOptions{
			BuildahMode:                 "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
		// TODO(ilya-lesikov): uncomment after Staged Dockerfile builder finished
		// // TODO(1.3): after Full Dockerfile Builder removed and Staged Dockerfile Builder enabled by default this test no longer needed
		// Entry("with local repo using Native Buildah and Staged Dockerfile builder with rootless isolation", complexTestOptions{
		// 	BuildahMode:                 "native-rootless",
		// 	WithLocalRepo:               true,
		// 	WithStagedDockerfileBuilder: true,
		// }),
		// TODO(ilya-lesikov): uncomment after Staged Dockerfile builder finished
		// // TODO(1.3): after Full Dockerfile Builder removed and Staged Dockerfile Builder enabled by default this test no longer needed
		// Entry("with local repo using Native Buildah and Staged Dockerfile builder with chroot isolation", complexTestOptions{
		// 	BuildahMode:                 "native-chroot",
		// 	WithLocalRepo:               true,
		// 	WithStagedDockerfileBuilder: true,
		// }),
	)
})
