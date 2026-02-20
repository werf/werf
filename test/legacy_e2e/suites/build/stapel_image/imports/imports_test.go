package imports_test

import (
	"context"
	"fmt"
	"path"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/utils/liveexec"
)

func werfBuild(ctx context.Context, dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(ctx, dir, SuiteData.WerfBinPath, opts, append([]string{"build"}, extraArgs...)...)
}

func werfRunOutput(ctx context.Context, dir string, extraArgs ...string) string {
	output, _ := utils.RunCommandWithOptions(ctx, dir, SuiteData.WerfBinPath, append([]string{"run", "--"}, extraArgs...), utils.RunCommandOptions{ShouldSucceed: true})
	return string(output)
}

func werfRunOutputWithSpecificImage(ctx context.Context, dir, image string, extraArgs ...string) string {
	output, _ := utils.RunCommandWithOptions(ctx, dir, SuiteData.WerfBinPath, append([]string{"run", image, "--"}, extraArgs...), utils.RunCommandOptions{ShouldSucceed: true})
	return string(output)
}

func werfHostPurge(ctx context.Context, dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(ctx, dir, SuiteData.WerfBinPath, opts, append([]string{"host", "purge"}, extraArgs...)...)
}

type pathChecks struct {
	files         []string
	dirs          []string
	excludedFiles []string
	excludedDirs  []string
}

func generatePathCheckCommands(checks pathChecks) []string {
	var commands []string

	for _, p := range checks.files {
		commands = append(commands, fmt.Sprintf("test -f %s || (echo 'FAIL: %s should exist' && false)", p, p))
	}

	for _, p := range checks.dirs {
		commands = append(commands, fmt.Sprintf("test -d %s || (echo 'FAIL: %s should exist' && false)", p, p))
	}

	for _, p := range checks.excludedFiles {
		commands = append(commands, fmt.Sprintf("! test -f %s || (echo 'FAIL: %s should NOT exist' && false)", p, p))
	}

	for _, p := range checks.excludedDirs {
		commands = append(commands, fmt.Sprintf("! test -d %s || (echo 'FAIL: %s should NOT exist' && false)", p, p))
	}

	return commands
}

var _ = Describe("Stapel imports", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("importing files and directories from artifact", func() {
		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{}, "--force")
		})

		It("should allow importing files and directories, optionally rename files and directories and merge directories", func(ctx SpecContext) {
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("imports_app_1", "001"), "initial commit")

			gotNoSuchFileError := false
			Expect(werfBuild(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, "/myartifact/no-such-dir") && strings.Contains(line, "No such file or directory") {
						gotNoSuchFileError = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotNoSuchFileError).To(BeTrue())

			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("imports_app_1", "002"), "add missing no-such-dir")

			gotNoSuchFileError = false
			Expect(werfBuild(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, "/myartifact/file-no-such-file") && strings.Contains(line, "No such file or directory") {
						gotNoSuchFileError = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotNoSuchFileError).To(BeTrue())

			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("imports_app_1", "003"), "add missing file-no-such-file")

			Expect(werfBuild(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())
			Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/local/FILE")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/locallll")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/newlocal/file")).To(ContainSubstring("GOGOGO\n"))
			Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/newlocal/a/b/FILE")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/share/file")).To(ContainSubstring("GOGOGO\n"))
			Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/usr/share/a/b/FILE")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "ls", "/usr/share/apk")).To(ContainSubstring("keys\n"))
			Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/file2")).To(ContainSubstring("GOGOGO\n"))
			Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", "/file")).To(ContainSubstring("GOGOGO\n"))
		})

		// There are three imports to different destination directories "/dest{1,3}".
		// All directories are expected to contain the same files "/dest{1,3}/added_file{1,2}".
		// The test consists of two steps, each step expects imported files with different content "VERSION_{1,2}\n".
		It("should import artifacts and reimport them at the next build (tests include and exclude paths)", func(ctx SpecContext) {
			for stepInd := 1; stepInd <= 2; stepInd++ {
				stepID := fmt.Sprintf("00%d", stepInd)
				stepDir := utils.FixturePath("imports_app_2", stepID)
				SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, stepDir, stepID)

				Expect(werfBuild(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

				expectedStepFileContent := fmt.Sprintf("VERSION_%d\n", stepInd)
				for destInd := 1; destInd <= 3; destInd++ {
					destDir := fmt.Sprintf("/dest%d", destInd)
					for fileInd := 1; fileInd <= 2; fileInd++ {
						destFilePath := path.Join(destDir, fmt.Sprintf("added_file%d", fileInd))
						Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "cat", destFilePath)).To(ContainSubstring(expectedStepFileContent))
					}

					for fileInd := 3; fileInd <= 4; fileInd++ {
						destFilePath := path.Join(destDir, fmt.Sprintf("not_added_file%d", fileInd))
						checkFileNotExistCommand := fmt.Sprintf("test -f %s || echo 'not exist'", destFilePath)
						Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "sh", "-c", checkFileNotExistCommand)).To(ContainSubstring("not exist"))
					}
				}
			}
		})

		It("should import symlink file", func(ctx SpecContext) {
			By("Check the symlink is added")
			var lastStageImageNameAfterFirstBuild string
			{
				SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("imports_app_4", "001"), "001")
				Expect(werfBuild(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())
				Expect(werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), "readlink", "/file")).To(ContainSubstring("/app/file"))
				lastStageImageNameAfterFirstBuild = utils.GetBuiltImageLastStageImageName(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "app")
			}

			By("Check nothing happens when the file to which the symlink points is changed")
			{
				Expect(werfBuild(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())
				lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "app")
				Expect(lastStageImageNameAfterFirstBuild).To(BeEquivalentTo(lastStageImageNameAfterSecondBuild))
			}
		})

		It("should import file from external image", func(ctx SpecContext) {
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("import_from_external", "001"), "initial commit")
			Expect(werfBuild(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())
			for i := 0; i <= 3; i++ {
				output := werfRunOutputWithSpecificImage(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), fmt.Sprintf("test-import-%d", i), "ls", "/etc/test/busybox")
				Expect(output).To(ContainSubstring(`/etc/test/busybox`))
			}
		})

		It("should import file with correct owner and group", func(ctx SpecContext) {
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("import_app_owner_group", "001"), "initial commit")

			Expect(werfBuild(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			output := werfRunOutput(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				"sh", "-c", "stat -c %u:%g /etc/test/busybox")

			Expect(output).To(ContainSubstring("11111:11111"))
		})

		It("should handle empty directories properly", func(ctx SpecContext) {
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("import_app_empty_dirs", "001"), "initial commit")

			Expect(werfBuild(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			By("Test 1: File glob (**/*.txt) should keep only directories with .txt files")
			test1Commands := []string{
				"echo '=== Test 1: File glob **/*.txt ===' && tree /test-tree || true",
			}
			test1Checks := pathChecks{
				files: []string{
					"/test-tree/add-dir/add-file.txt",
					"/test-tree/add-dir/sub/add-file.txt",
				},
				dirs: []string{
					"/test-tree/add-dir",
					"/test-tree/add-dir/sub",
				},
				excludedFiles: []string{
					"/test-tree/add-dir/not-add-file.log",
				},
				excludedDirs: []string{
					"/test-tree/not-add-dir",
					"/test-tree/app",
				},
			}
			test1Commands = append(test1Commands, generatePathCheckCommands(test1Checks)...)

			werfRunOutputWithSpecificImage(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				"final-file-glob", "sh", "-c", strings.Join(test1Commands, " && "))

			By("Test 2: Directory glob (app/**/add-dir) should keep matching empty directories")
			test2Commands := []string{
				"echo '=== Test 2: Directory glob app/**/add-dir ===' && tree /test-tree || true",
			}
			test2Checks := pathChecks{
				dirs: []string{
					"/test-tree/app/add-dir",
					"/test-tree/app/foo/bar/add-dir",
				},
				excludedDirs: []string{
					"/test-tree/app/not-add-dir",
					"/test-tree/app/foo/not-add-dir",
					"/test-tree/not-add-dir",
				},
			}
			test2Commands = append(test2Commands, generatePathCheckCommands(test2Checks)...)

			werfRunOutputWithSpecificImage(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				"final-dir-glob", "sh", "-c", strings.Join(test2Commands, " && "))
		})
	})

	Context("caching by import source checksum", func() {
		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{}, "--force")
		})

		It("should cache image when import source checksum was not changed", func(ctx SpecContext) {
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("import_metadata", "001"), "initial commit")

			utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build")

			lastStageImageNameAfterFirstBuild := utils.GetBuiltImageLastStageImageName(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("import_metadata", "002"), "change artifact fromCacheVersion")

			utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build")

			lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			Expect(lastStageImageNameAfterFirstBuild).Should(Equal(lastStageImageNameAfterSecondBuild))
		})

		It("should rebuild image when import source checksum was changed under experimental flag", func(ctx SpecContext) {
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("import_metadata", "001"), "initial commit")

			utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build")

			lastStageImageNameAfterFirstBuild := utils.GetBuiltImageLastStageImageName(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("import_metadata", "003"), "change artifact install stage")

			utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build")

			lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "image")

			Expect(lastStageImageNameAfterFirstBuild).ShouldNot(Equal(lastStageImageNameAfterSecondBuild))
		})

		It("should rebuild image when import source checksum and permissions were changed", func(ctx SpecContext) {
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("imports_app_5", "001"), "initial commit")

			utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build")

			lastStageImageNameAfterFirstBuild := utils.GetBuiltImageLastStageImageName(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "final")

			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("imports_app_5", "002"), "change permissions")

			SuiteData.Stubs.SetEnv("WERF_EXPERIMENTAL_STAPEL_IMPORT_PERMISSIONS", "true")
			_, _ = utils.RunCommandWithOptions(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, []string{"build"}, utils.RunCommandOptions{
				ShouldSucceed: true,
			})

			lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "final")

			Expect(lastStageImageNameAfterFirstBuild).ShouldNot(Equal(lastStageImageNameAfterSecondBuild))
		})

		It("should account symlinks when calculating import source checksum", func(ctx SpecContext) {
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("import_app_symlinks", "001"), "initial commit")

			utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build")

			lastStageImageNameAfterFirstBuild := utils.GetBuiltImageLastStageImageName(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "final")

			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("import_app_symlinks", "002"), "change symlink target")

			utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build")

			lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "final")

			Expect(lastStageImageNameAfterFirstBuild).ShouldNot(Equal(lastStageImageNameAfterSecondBuild))
		})
	})
})
