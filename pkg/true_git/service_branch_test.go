package true_git

import (
	"context"
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/pkg/werf"
)

var _ = Describe("SyncSourceWorktreeWithServiceBranch", func() {
	var sourceWorkTreeDir string
	var gitDir string
	var workTreeCacheDir string
	var headCommit string
	defaultOptions := SyncSourceWorktreeWithServiceBranchOptions{ServiceBranchPrefix: "dev-"}

	BeforeEach(func() {
		sourceWorkTreeDir = filepath.Join(SuiteData.TestDirPath, "source")
		utils.MkdirAll(sourceWorkTreeDir)
		workTreeCacheDir = filepath.Join(SuiteData.TestDirPath, "worktree")
		utils.MkdirAll(workTreeCacheDir)

		utils.RunSucceedCommand(
			sourceWorkTreeDir,
			"git",
			"init",
		)
		gitDir = filepath.Join(sourceWorkTreeDir, ".git")

		utils.RunSucceedCommand(
			sourceWorkTreeDir,
			"git",
			"commit", "--allow-empty", "-m", "Initial commit",
		)

		headCommit = utils.GetHeadCommit(sourceWorkTreeDir)

		Ω(werf.Init("", "")).Should(Succeed())
		Ω(Init(context.Background(), Options{})).Should(Succeed())
	})

	It("no changes", func() {
		commit, err := SyncSourceWorktreeWithServiceBranch(
			context.Background(),
			gitDir,
			sourceWorkTreeDir,
			workTreeCacheDir,
			headCommit,
			defaultOptions,
		)

		Ω(err).Should(Succeed())
		Ω(commit).Should(Equal(headCommit))
	})

	When("tracked changes", func() {
		const trackedFileRelPath = "tracked_file"
		var trackedFilePath string

		BeforeEach(func() {
			trackedFilePath = filepath.Join(sourceWorkTreeDir, trackedFileRelPath)
			utils.WriteFile(trackedFilePath, []byte("state"))

			utils.RunSucceedCommand(
				sourceWorkTreeDir,
				"git",
				"add", trackedFilePath,
			)
		})

		It("add and reproducibility", func() {
			serviceCommit1, err := SyncSourceWorktreeWithServiceBranch(
				context.Background(),
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				headCommit,
				defaultOptions,
			)

			Ω(err).Should(Succeed())
			Ω(serviceCommit1).ShouldNot(Equal(headCommit))

			diff := utils.SucceedCommandOutputString(
				sourceWorkTreeDir,
				"git",
				"diff", serviceCommit1, trackedFileRelPath,
			)

			Ω(diff).Should(BeEmpty())

			serviceCommit2, err := SyncSourceWorktreeWithServiceBranch(
				context.Background(),
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				headCommit,
				defaultOptions,
			)

			Ω(err).Should(Succeed())
			Ω(serviceCommit1).Should(Equal(serviceCommit2))
		})
	})

	When("untracked changes", func() {
		const trackedFileRelPath = "untracked_file"
		var trackedFilePath string
		trackedFileContent1 := []byte("state1")

		BeforeEach(func() {
			trackedFilePath = filepath.Join(sourceWorkTreeDir, trackedFileRelPath)
			utils.WriteFile(trackedFilePath, trackedFileContent1)
		})

		It("add and reproducibility", func() {
			serviceCommit1, err := SyncSourceWorktreeWithServiceBranch(
				context.Background(),
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				headCommit,
				defaultOptions,
			)

			Ω(err).Should(Succeed())
			Ω(serviceCommit1).ShouldNot(Equal(headCommit))

			content := utils.SucceedCommandOutputString(
				sourceWorkTreeDir,
				"git",
				"show", serviceCommit1+":"+trackedFileRelPath,
			)

			Ω(content).Should(Equal(string(trackedFileContent1)))

			serviceCommit2, err := SyncSourceWorktreeWithServiceBranch(
				context.Background(),
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				headCommit,
				defaultOptions,
			)

			Ω(err).Should(Succeed())
			Ω(serviceCommit1).Should(Equal(serviceCommit2))
		})

		When("untracked file already added", func() {
			var serviceCommitUntrackedFileAdded string
			trackedFileContent2 := []byte("state2")

			BeforeEach(func() {
				serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
					context.Background(),
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					headCommit,
					defaultOptions,
				)

				Ω(err).Should(Succeed())

				serviceCommitUntrackedFileAdded = serviceCommit
			})

			It("change", func() {
				utils.WriteFile(trackedFilePath, trackedFileContent2)

				serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
					context.Background(),
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					headCommit,
					defaultOptions,
				)

				Ω(err).Should(Succeed())
				Ω(serviceCommit).ShouldNot(Equal(serviceCommitUntrackedFileAdded))

				content := utils.SucceedCommandOutputString(
					sourceWorkTreeDir,
					"git",
					"show", serviceCommit+":"+trackedFileRelPath,
				)

				Ω(content).Should(Equal(string(trackedFileContent2)))
			})

			It("stage", func() {
				utils.RunSucceedCommand(
					sourceWorkTreeDir,
					"git",
					"add", trackedFilePath,
				)

				serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
					context.Background(),
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					headCommit,
					defaultOptions,
				)

				Ω(err).Should(Succeed())
				Ω(serviceCommit).Should(Equal(serviceCommitUntrackedFileAdded))

				content := utils.SucceedCommandOutputString(
					sourceWorkTreeDir,
					"git",
					"show", serviceCommit+":"+trackedFileRelPath,
				)

				Ω(content).Should(Equal(string(trackedFileContent1)))
			})

			It("delete", func() {
				utils.DeleteFile(trackedFilePath)

				serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
					context.Background(),
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					headCommit,
					defaultOptions,
				)

				Ω(err).Should(Succeed())
				Ω(serviceCommit).ShouldNot(Equal(serviceCommitUntrackedFileAdded))

				bytes, err := utils.RunCommand(
					sourceWorkTreeDir,
					"git",
					"show", serviceCommit+":"+trackedFileRelPath,
				)
				Ω(err).Should(HaveOccurred())

				output := string(bytes)
				Ω(output).Should(ContainSubstring(fmt.Sprintf("'%s' does not exist in '%s'", trackedFileRelPath, serviceCommit)))
			})
		})
	})

	When("glob exclude specified", func() {
		const untrackedFileRelPath = "untracked_file.ext"
		var untrackedFilePath string
		var serviceCommitUntrackedFileAdded string

		BeforeEach(func() {
			untrackedFilePath = filepath.Join(sourceWorkTreeDir, untrackedFileRelPath)
			utils.WriteFile(untrackedFilePath, []byte("any"))

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
				context.Background(),
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				headCommit,
				defaultOptions,
			)

			Ω(err).Should(Succeed())

			serviceCommitUntrackedFileAdded = serviceCommit
		})

		It("not ignore", func() {
			defaultOptions.GlobExcludeList = []string{"file"}

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
				context.Background(),
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				headCommit,
				defaultOptions,
			)

			Ω(err).Should(Succeed())
			Ω(serviceCommit).Should(Equal(serviceCommitUntrackedFileAdded))
		})

		It("ignore", func() {
			defaultOptions.GlobExcludeList = []string{"*.ext"}

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
				context.Background(),
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				headCommit,
				defaultOptions,
			)

			Ω(err).Should(Succeed())
			Ω(serviceCommit).ShouldNot(Equal(serviceCommitUntrackedFileAdded))

			bytes, err := utils.RunCommand(
				sourceWorkTreeDir,
				"git",
				"show", serviceCommit+":"+untrackedFileRelPath,
			)
			Ω(err).Should(HaveOccurred())

			output := string(bytes)
			Ω(output).Should(ContainSubstring(fmt.Sprintf("'%s' exists on disk, but not in '%s'", untrackedFileRelPath, serviceCommit)))
		})
	})
})
