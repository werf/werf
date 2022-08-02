package imports_test

import (
	"fmt"
	"path"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
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
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("imports_app_1", "001"), "initial commit")

			gotNoSuchFileError := false
			Expect(werfBuild(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, "/myartifact/no-such-dir") && strings.Contains(line, "No such file or directory") {
						gotNoSuchFileError = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotNoSuchFileError).To(BeTrue())

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("imports_app_1", "002"), "add missing no-such-dir")

			gotNoSuchFileError = false
			Expect(werfBuild(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, "/myartifact/file-no-such-file") && strings.Contains(line, "No such file or directory") {
						gotNoSuchFileError = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotNoSuchFileError).To(BeTrue())

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("imports_app_1", "003"), "add missing file-no-such-file")

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

		// There are three imports to different destination directories "/dest{1,3}".
		// All directories are expected to contain the same files "/dest{1,3}/added_file{1,2}".
		// The test consists of two steps, each step expects imported files with different content "VERSION_{1,2}\n".
		It("should import artifacts and reimport them at the next build (tests include and exclude paths)", func() {
			for stepInd := 1; stepInd <= 2; stepInd++ {
				stepID := fmt.Sprintf("00%d", stepInd)
				stepDir := utils.FixturePath("imports_app_2", stepID)
				SuiteData.CommitProjectWorktree(SuiteData.ProjectName, stepDir, stepID)

				Expect(werfBuild(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

				expectedStepFileContent := fmt.Sprintf("VERSION_%d\n", stepInd)
				for destInd := 1; destInd <= 3; destInd++ {
					destDir := fmt.Sprintf("/dest%d", destInd)
					for fileInd := 1; fileInd <= 2; fileInd++ {
						destFilePath := path.Join(destDir, fmt.Sprintf("added_file%d", fileInd))
						Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", destFilePath)).To(ContainSubstring(expectedStepFileContent))
					}

					for fileInd := 3; fileInd <= 4; fileInd++ {
						destFilePath := path.Join(destDir, fmt.Sprintf("not_added_file%d", fileInd))
						checkFileNotExistCommand := fmt.Sprintf("test -f %s || echo 'not exist'", destFilePath)
						Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "sh", "-c", checkFileNotExistCommand)).To(ContainSubstring("not exist"))
					}
				}
			}
		})

		It("should import artifacts from directory to which the symlink points", func() {
			for stepInd := 1; stepInd <= 2; stepInd++ {
				stepID := fmt.Sprintf("00%d", stepInd)
				stepDir := utils.FixturePath("imports_app_3", stepID)
				SuiteData.CommitProjectWorktree(SuiteData.ProjectName, stepDir, stepID)

				Expect(werfBuild(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

				expectedStepFileContent := fmt.Sprintf("VERSION_%d\n", stepInd)
				destFilePath := "/dir/file"
				Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", destFilePath)).To(ContainSubstring(expectedStepFileContent))
			}
		})

		It("should import symlink file", func() {
			By("Check the symlink is added")
			var lastStageImageNameAfterFirstBuild string
			{
				SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("imports_app_4", "001"), "001")
				Expect(werfBuild(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())
				Expect(werfRunOutput(SuiteData.GetProjectWorktree(SuiteData.ProjectName), "readlink", "/file")).To(ContainSubstring("/app/file"))
				lastStageImageNameAfterFirstBuild = utils.GetBuiltImageLastStageImageName(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "app")
			}

			By("Check nothing happens when the file to which the symlink points is changed")
			{
				Expect(werfBuild(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())
				lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "app")
				Expect(lastStageImageNameAfterFirstBuild).To(BeEquivalentTo(lastStageImageNameAfterSecondBuild))
			}
		})
	})

	Context("caching by import source checksum", func() {
		AfterEach(func() {
			werfHostPurge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{}, "--force")
		})

		It("should cache image when import source checksum was not changed", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("import_metadata", "001"), "initial commit")

			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterFirstBuild := utils.GetBuiltImageLastStageImageName(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("import_metadata", "002"), "change artifact fromCacheVersion")

			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			Ω(lastStageImageNameAfterFirstBuild).Should(Equal(lastStageImageNameAfterSecondBuild))
		})

		It("should rebuild image when import source checksum was changed", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("import_metadata", "001"), "initial commit")

			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterFirstBuild := utils.GetBuiltImageLastStageImageName(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("import_metadata", "003"), "change artifact install stage")

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
