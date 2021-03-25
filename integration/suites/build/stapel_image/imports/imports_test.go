package imports_test

import (
	"strings"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/integration/pkg/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func werfBuild(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(append([]string{"build"}, extraArgs...)...)...)
}

func werfRunOutput(dir string, extraArgs ...string) string {
	output, _ := utils.RunCommandWithOptions(
		dir, SuiteData.WerfBinPath,
		append([]string{"run", "--"}, extraArgs...),
		utils.RunCommandOptions{ShouldSucceed: true},
	)
	return string(output)
}

func werfHostPurge(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(append([]string{"host", "purge"}, extraArgs...)...)...)
}

var _ = Describe("Stapel imports", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("importing files and directories from artifact", func() {
		AfterEach(func() {
			werfHostPurge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{}, "--force")
		})

		It("should allow importing files and directories, optionally rename files and directories and merge directories", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "imports_app1-001", "initial commit")

			gotNoSuchFileError := false
			Expect(werfBuild(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, "/myartifact/no-such-dir") && strings.Contains(line, "No such file or directory") {
						gotNoSuchFileError = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotNoSuchFileError).To(BeTrue())

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "imports_app1-002", "add missing no-such-dir")

			gotNoSuchFileError = false
			Expect(werfBuild(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, "/myartifact/file-no-such-file") && strings.Contains(line, "No such file or directory") {
						gotNoSuchFileError = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotNoSuchFileError).To(BeTrue())

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "imports_app1-003", "add missing file-no-such-file")

			Expect(werfBuild(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/local/FILE")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/locallll")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/newlocal/file")).To(ContainSubstring("GOGOGO\n"))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/newlocal/a/b/FILE")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/share/file")).To(ContainSubstring("GOGOGO\n"))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/share/a/b/FILE")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "ls", "/usr/share/apk")).To(ContainSubstring("keys\n"))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/file2")).To(ContainSubstring("GOGOGO\n"))
			Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/file")).To(ContainSubstring("GOGOGO\n"))
		})
	})

	Context("caching by import source checksum", func() {
		AfterEach(func() {
			werfHostPurge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{}, "--force")
		})

		It("should cache image when import source checksum was not changed", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "import_metadata-001", "initial commit")

			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterFirstBuild := utils.GetBuiltImageLastStageImageName(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "import_metadata-002", "change artifact fromCacheVersion")

			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			Ω(lastStageImageNameAfterFirstBuild).Should(Equal(lastStageImageNameAfterSecondBuild))
		})

		It("should rebuild image when import source checksum was changed", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "import_metadata-001", "initial commit")

			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterFirstBuild := utils.GetBuiltImageLastStageImageName(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "import_metadata-003", "change artifact install stage")

			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			Ω(lastStageImageNameAfterFirstBuild).ShouldNot(Equal(lastStageImageNameAfterSecondBuild))
		})
	})
})
