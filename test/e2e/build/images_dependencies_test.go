package e2e_build_test

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
)

func werfBuild(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(append([]string{"build"}, extraArgs...)...)...)
}

func werfRun(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(append([]string{"run"}, extraArgs...)...)...)
}

func werfStageImage(dir string, imageName string) (string, string) {
	res := utils.SucceedCommandOutputString(
		dir,
		SuiteData.WerfBinPath,
		"stage", "image", imageName,
	)

	return image.ParseRepositoryAndTag(strings.TrimSpace(res))
}

func werfRunOutput(dir string, imageName string, shellCommand string) string {
	handlingOutput := false

	var output []string

	Expect(werfRun(dir, liveexec.ExecCommandOptions{
		OutputLineHandler: func(line string) {
			if strings.HasPrefix(line, "START_OUTPUT") {
				handlingOutput = true
				return
			}

			if handlingOutput {
				output = append(output, line)
			}
		},
	}, imageName, "--", "sh", "-c", fmt.Sprintf("echo START_OUTPUT && %s", shellCommand))).To(Succeed())

	return strings.Join(output, "\n")
}

func getImageID(ctx context.Context, ref string, containerRuntime container_runtime.ContainerRuntime) string {
	info, err := containerRuntime.GetImageInfo(ctx, ref, container_runtime.GetImageInfoOpts{})
	Expect(err).To(Succeed())
	return info.ID
}

var _ = Describe("Images dependencies", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	When("werf.yaml contains stapel and dockerfile images which used as dependencies in another stapel and dockerfile images", func() {
		It("should successfully build images using specified dependencies", func() {
			ctx := context.Background()
			containerRuntime := container_runtime.NewDockerServerRuntime()

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "_fixtures/images_dependencies-001", "initial commit")
			Expect(werfBuild(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /install/BASE_STAPEL_IMAGE_NAME")).To(BeEmpty())
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /install/BASE_STAPEL_IMAGE_ID")).To(BeEmpty())
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /install/BASE_STAPEL_IMAGE_REPO")).To(BeEmpty())
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /install/BASE_STAPEL_IMAGE_TAG")).To(BeEmpty())

			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /install/BASE_DOCKERFILE_IMAGE_NAME")).To(BeEmpty())
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /install/BASE_DOCKERFILE_IMAGE_ID")).To(BeEmpty())
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /install/BASE_DOCKERFILE_IMAGE_REPO")).To(BeEmpty())
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /install/BASE_DOCKERFILE_IMAGE_TAG")).To(BeEmpty())

			baseStapelRepo, baseStapelTag := werfStageImage(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "base-stapel")
			baseStapelName := fmt.Sprintf("%s:%s", baseStapelRepo, baseStapelTag)
			baseStapelID := getImageID(ctx, baseStapelName, containerRuntime)

			baseDockerfileRepo, baseDockerfileTag := werfStageImage(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "base-dockerfile")
			baseDockerfileName := fmt.Sprintf("%s:%s", baseDockerfileRepo, baseDockerfileTag)
			baseDockerfileID := getImageID(ctx, baseDockerfileName, containerRuntime)

			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /setup/BASE_STAPEL_IMAGE_NAME")).To(Equal(baseStapelName))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /setup/BASE_STAPEL_IMAGE_ID")).To(Equal(baseStapelID))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /setup/BASE_STAPEL_IMAGE_REPO")).To(Equal(baseStapelRepo))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /setup/BASE_STAPEL_IMAGE_TAG")).To(Equal(baseStapelTag))

			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /setup/BASE_DOCKERFILE_IMAGE_NAME")).To(Equal(baseDockerfileName))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /setup/BASE_DOCKERFILE_IMAGE_ID")).To(Equal(baseDockerfileID))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /setup/BASE_DOCKERFILE_IMAGE_REPO")).To(Equal(baseDockerfileRepo))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "stapel", "cat /setup/BASE_DOCKERFILE_IMAGE_TAG")).To(Equal(baseDockerfileTag))

			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "dockerfile", "cat /BASE_STAPEL_IMAGE_NAME")).To(Equal(baseStapelName))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "dockerfile", "cat /BASE_STAPEL_IMAGE_ID")).To(Equal(baseStapelID))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "dockerfile", "cat /BASE_STAPEL_IMAGE_REPO")).To(Equal(baseStapelRepo))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "dockerfile", "cat /BASE_STAPEL_IMAGE_TAG")).To(Equal(baseStapelTag))

			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "dockerfile", "cat /BASE_DOCKERFILE_IMAGE_NAME")).To(Equal(baseDockerfileName))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "dockerfile", "cat /BASE_DOCKERFILE_IMAGE_ID")).To(Equal(baseDockerfileID))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "dockerfile", "cat /BASE_DOCKERFILE_IMAGE_REPO")).To(Equal(baseDockerfileRepo))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "dockerfile", "cat /BASE_DOCKERFILE_IMAGE_TAG")).To(Equal(baseDockerfileTag))
		})
	})
})
