package e2e_build_test

import (
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = ginkgo.Describe("Legacy stage script", ginkgo.Label("e2e", "build", "legacy-stage-script"), func() {
	ginkgo.DescribeTable("executes the complete script with closed command stdin", func(ctx ginkgo.SpecContext, script string, exitCode int) {
		gomega.Expect(werf.Init(SuiteData.TmpDir, "")).To(gomega.Succeed())
		gomega.Expect(docker.Init(ctx, docker.InitOptions{})).To(gomega.Succeed())
		buildCtx := utils.WithDependencies(ctx)
		backend := container_backend.NewDockerServerBackend(werf.HostLocker().Locker())
		const baseRef = "registry.werf.io/base/ubuntu:22.04"
		utils.RunSucceedCommand(ctx, "", "docker", "pull", "--platform=linux/amd64", baseRef)
		base := container_backend.NewLegacyStageImage(nil, baseRef, backend, "linux/amd64")
		stage := container_backend.NewLegacyStageImage(base, SuiteData.ProjectName, backend, "linux/amd64")
		stage.BuilderContainer().AddRunCommands(script, "printf complete > /stage-complete")

		err := stage.Build(buildCtx, container_backend.BuildOptions{})
		ginkgo.DeferCleanup(func(ctx ginkgo.SpecContext) {
			if stage.BuiltID() != "" {
				utils.RunSucceedCommand(ctx, "", "docker", "rmi", stage.BuiltID())
			}
		})
		if exitCode != 0 {
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring(fmt.Sprintf("Code: %d", exitCode)))
			gomega.Expect(stage.BuiltID()).To(gomega.BeEmpty())
			return
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		out, err := docker.CliRun_RecordedOutput(buildCtx, "--rm", "--platform=linux/amd64", "--entrypoint=/bin/cat", stage.BuiltID(), "/stage-complete")
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), out)
		gomega.Expect(out).To(gomega.Equal("complete"))
	},
		ginkgo.Entry("stdin consumers cannot eat later commands", "cat > /tmp/eaten\n! read -r line\ntest ! -s /tmp/eaten", 0),
		ginkgo.Entry("a script exceeding the argument-size limit", strings.Repeat(": import-padding\n", 200_000)+":", 0),
		ginkgo.Entry("a failing command prevents a committed image", "exit 23", 23),
		ginkgo.Entry("a syntax error prevents a committed image", "printf early\n)", 2),
	)
})
