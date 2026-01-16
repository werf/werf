package e2e_build_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type heredocTestOptions struct {
	setupEnvOptions

	FixtureRelPath string
	Verify         []string
}

var _ = Describe("Build with staged dockerfile and heredoc", Label("e2e", "build", "heredoc", "simple"), func() {
	DescribeTable("should succeed and produce expected image with heredoc content",
		func(ctx SpecContext, testOpts heredocTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if errors.Is(err, contback.ErrRuntimeUnavailable) {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("heredoc: starting")
			{
				repoDirName := "repo0"
				fixtureRelPath := testOpts.FixtureRelPath
				buildReportName := "report0.json"

				By("heredoc: preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirName, fixtureRelPath)

				By("heredoc: building image")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))
				reportProject := report.NewProjectWithReport(werfProject)

				buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)
				Expect(buildOut).To(ContainSubstring("Building stage"))
				Expect(buildOut).NotTo(ContainSubstring("Use previously built image"))

				By(fmt.Sprintf(`heredoc: checking "dockerfile" image %s content`, buildReport.Images["dockerfile"].DockerImageName))
				contRuntime.ExpectCmdsToSucceed(ctx, buildReport.Images["dockerfile"].DockerImageName, testOpts.Verify...)
			}
		},
		Entry("with simple heredoc content and local repo using Native Buildah with rootless isolation", heredocTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: true,
			},
			FixtureRelPath: "heredoc/simple",
			Verify:         []string{"test -d /etc/myapp", "test -f /etc/myapp/env", "(echo 'FOO=bar' && echo 'BAR=baz') | diff /etc/myapp/env -"},
		}),
		Entry("with simple heredoc content and local repo using Native Buildah with chroot isolation", heredocTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode:        "native-chroot",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: true,
			},
			FixtureRelPath: "heredoc/simple",
			Verify:         []string{"test -d /etc/myapp", "test -f /etc/myapp/env", "(echo 'FOO=bar' && echo 'BAR=baz') | diff /etc/myapp/env -"},
		}),
		Entry("with multiple heredoc content and local repo using Native Buildah with rootless isolation", heredocTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: true,
			},
			FixtureRelPath: "heredoc/multiple",
			Verify:         []string{"test -f /file1", "test -f /file2", "echo -e 'I am\\nfirst' | diff /file1 -", "echo -e 'I am\\nsecond' | diff /file2 -"},
		}),
		Entry("with multiple heredoc content and local repo using Native Buildah with chroot isolation", heredocTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode:        "native-chroot",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: true,
			},
			FixtureRelPath: "heredoc/multiple",
			Verify:         []string{"test -f /file1", "test -f /file2", "echo -e 'I am\\nfirst' | diff /file1 -", "echo -e 'I am\\nsecond' | diff /file2 -"},
		}),
	)

	DescribeTable("should fail and return error message with unsupported heredoc syntax",
		func(ctx SpecContext, testOpts heredocTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			_, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if errors.Is(err, contback.ErrRuntimeUnavailable) {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("heredoc: starting")
			{
				repoDirName := "repo0"
				fixtureRelPath := testOpts.FixtureRelPath

				By("heredoc: preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirName, fixtureRelPath)

				By("heredoc: building image")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))
				buildOut := werfProject.Build(ctx, &werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: true,
					},
				})

				By("heredoc: checking error messages")
				Expect(buildOut).To(ContainSubstring(testOpts.Verify[0]))
			}
		},
		Entry("with unsupported COPY heredoc and local repo using Native Buildah with rootless isolation", heredocTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: true,
			},
			FixtureRelPath: "heredoc/copy",
			Verify:         []string{"heredoc is not supported with COPY command"},
		}),
		Entry("with unsupported COPY heredoc and local repo using Native Buildah with chroot isolation", heredocTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode:        "native-chroot",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: true,
			},
			FixtureRelPath: "heredoc/copy",
			Verify:         []string{"heredoc is not supported with COPY command"},
		}),
		Entry("with unsupported ADD heredoc and local repo using Native Buildah with rootless isolation", heredocTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: true,
			},
			FixtureRelPath: "heredoc/add",
			Verify:         []string{"heredoc is not supported with ADD command"},
		}),
		Entry("with unsupported ADD heredoc and local repo using Native Buildah with chroot isolation", heredocTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode:        "native-chroot",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: true,
			},
			FixtureRelPath: "heredoc/add",
			Verify:         []string{"heredoc is not supported with ADD command"},
		}),
	)
})
