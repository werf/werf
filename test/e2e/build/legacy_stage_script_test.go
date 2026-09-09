package e2e_build_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/docker/cli/cli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/utils"
	werftest "github.com/werf/werf/v2/test/pkg/werf"
)

const (
	legacyStageScriptBaseImage = "registry.werf.io/base/ubuntu:22.04"

	// The build must fail, whatever exit code the shell reports for the script.
	legacyStageScriptAnyExitCode = -1
)

var _ = Describe("Legacy stage script", Label("e2e", "build", "legacy-stage-script"), func() {
	It("builds a werf.yaml with many imports", func(ctx SpecContext) {
		setupEnv(setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: true})
		contRuntime, err := contback.NewContainerBackend("docker")
		if err == contback.ErrRuntimeUnavailable {
			Skip(err.Error())
		} else if err != nil {
			Fail(err.Error())
		}

		SuiteData.InitTestRepo(ctx, "repo0", "legacy-stage-script/state0")
		werfProject := werftest.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath("repo0"))
		_, buildReport := report.NewProjectWithReport(werfProject).BuildWithReport(ctx, SuiteData.GetBuildReportPath("report.json"), &werftest.WithReportOptions{
			CommonOptions: werftest.CommonOptions{ExtraArgs: []string{"--platform=linux/amd64"}},
		})

		contRuntime.ExpectCmdsToSucceed(ctx, buildReport.Images["target"].DockerImageName,
			`for i in $(seq 0 255); do test "$(cat /imports/$i)" = payload || exit 1; done`,
			"test ! -s /tmp/eaten",
			"test -f /stage-complete",
		)
	})

	Describe("stage script transport", func() {
		BeforeEach(func(ctx SpecContext) {
			Expect(werf.Init(SuiteData.TmpDir, "")).To(Succeed())
			Expect(docker.Init(ctx, docker.InitOptions{})).To(Succeed())
			utils.RunSucceedCommand(ctx, "", "docker", "pull", "--platform=linux/amd64", legacyStageScriptBaseImage)
		})

		DescribeTable("executes the complete script with closed command stdin", func(ctx SpecContext, script string, exitCode int) {
			buildCtx := utils.WithDependencies(ctx)
			backend := container_backend.NewDockerServerBackend(werf.HostLocker().Locker())
			base := container_backend.NewLegacyStageImage(nil, legacyStageScriptBaseImage, backend, "linux/amd64")
			stage := container_backend.NewLegacyStageImage(base, SuiteData.ProjectName, backend, "linux/amd64")
			stage.BuilderContainer().AddRunCommands(script, "printf complete > /stage-complete")

			err := stage.Build(buildCtx, container_backend.BuildOptions{})
			DeferCleanup(func(ctx SpecContext) {
				if stage.BuiltID() != "" {
					utils.RunSucceedCommand(ctx, "", "docker", "rmi", stage.BuiltID())
				}
			})
			if exitCode != 0 {
				var statusErr cli.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue(), fmt.Sprintf("%v", err))
				if exitCode == legacyStageScriptAnyExitCode {
					Expect(statusErr.StatusCode).NotTo(BeZero())
				} else {
					Expect(statusErr.StatusCode).To(Equal(exitCode))
				}
				Expect(stage.BuiltID()).To(BeEmpty())
				return
			}
			Expect(err).NotTo(HaveOccurred())
			out, err := docker.CliRun_RecordedOutput(buildCtx, "--rm", "--platform=linux/amd64", "--entrypoint=/bin/cat", stage.BuiltID(), "/stage-complete")
			Expect(err).NotTo(HaveOccurred(), out)
			Expect(out).To(ContainSubstring("complete"))
		},
			Entry("stdin consumers cannot eat later commands", "cat > /tmp/eaten\n! read -r line\ntest ! -s /tmp/eaten", 0),
			Entry("a script exceeding the argument-size limit", strings.Repeat(": import-padding\n", 200_000)+":", 0),
			Entry("a failing command prevents a committed image", "exit 23", 23),
			Entry("a syntax error prevents a committed image", "printf early\n)", legacyStageScriptAnyExitCode),
		)
	})
})
