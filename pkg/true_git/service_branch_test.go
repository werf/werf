package true_git

import (
	"context"
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("SyncSourceWorktreeWithServiceBranch", func() {
	var sourceWorkTreeDir string
	var gitDir string
	var workTreeCacheDir string
	var sourceHeadCommit string
	defaultOptions := SyncSourceWorktreeWithServiceBranchOptions{ServiceBranch: "_werf-dev"}

	BeforeEach(func(ctx SpecContext) {
		sourceWorkTreeDir = filepath.Join(SuiteData.TestDirPath, "source")
		utils.MkdirAll(sourceWorkTreeDir)
		workTreeCacheDir = filepath.Join(SuiteData.TestDirPath, "worktree")
		utils.MkdirAll(workTreeCacheDir)

		utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "-c", "init.defaultBranch=main", "init")

		utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "checkout", "-b", "main")

		gitDir = filepath.Join(sourceWorkTreeDir, ".git")

		utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "commit", "--allow-empty", "-m", "Initial commit")

		sourceHeadCommit = utils.GetHeadCommit(ctx, sourceWorkTreeDir)

		Expect(werf.Init("", "")).Should(Succeed())
		Expect(Init(ctx, Options{})).Should(Succeed())
	})

	It("no changes", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)

		commit, err := SyncSourceWorktreeWithServiceBranch(
			ctx,
			gitDir,
			sourceWorkTreeDir,
			workTreeCacheDir,
			sourceHeadCommit,
			defaultOptions,
		)

		Expect(err).Should(Succeed())
		Expect(commit).Should(Equal(sourceHeadCommit))
	})

	When("tracked changes", func() {
		const trackedFileRelPath = "tracked_file"
		var trackedFilePath string

		BeforeEach(func(ctx SpecContext) {
			trackedFilePath = filepath.Join(sourceWorkTreeDir, trackedFileRelPath)
			utils.WriteFile(trackedFilePath, []byte("state"))

			utils.RunSucceedCommand(
				ctx,
				sourceWorkTreeDir,
				"git",
				"add", trackedFilePath,
			)
		})

		It("add and reproducibility", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			serviceCommit1, err := SyncSourceWorktreeWithServiceBranch(
				ctx,
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				sourceHeadCommit,
				defaultOptions,
			)

			Expect(err).Should(Succeed())
			Expect(serviceCommit1).ShouldNot(Equal(sourceHeadCommit))

			diff := utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "diff", serviceCommit1, trackedFileRelPath)

			Expect(diff).Should(BeEmpty())

			serviceCommit2, err := SyncSourceWorktreeWithServiceBranch(
				ctx,
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				sourceHeadCommit,
				defaultOptions,
			)

			Expect(err).Should(Succeed())
			Expect(serviceCommit1).Should(Equal(serviceCommit2))
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

		It("add and reproducibility", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			serviceCommit1, err := SyncSourceWorktreeWithServiceBranch(
				ctx,
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				sourceHeadCommit,
				defaultOptions,
			)

			Expect(err).Should(Succeed())
			Expect(serviceCommit1).ShouldNot(Equal(sourceHeadCommit))

			content := utils.SucceedCommandOutputString(
				ctx,
				sourceWorkTreeDir,
				"git",
				"show", serviceCommit1+":"+trackedFileRelPath,
			)

			Expect(content).Should(Equal(string(trackedFileContent1)))

			serviceCommit2, err := SyncSourceWorktreeWithServiceBranch(
				ctx,
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				sourceHeadCommit,
				defaultOptions,
			)

			Expect(err).Should(Succeed())
			Expect(serviceCommit1).Should(Equal(serviceCommit2))
		})

		When("untracked file already added", func() {
			var serviceCommitUntrackedFileAdded string
			trackedFileContent2 := []byte("state2")

			BeforeEach(func(ctx context.Context) {
				ctx = logging.WithLogger(ctx)

				serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)

				Expect(err).Should(Succeed())

				serviceCommitUntrackedFileAdded = serviceCommit
			})

			It("change", func(ctx context.Context) {
				ctx = logging.WithLogger(ctx)

				utils.WriteFile(trackedFilePath, trackedFileContent2)

				serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)

				Expect(err).Should(Succeed())
				Expect(serviceCommit).ShouldNot(Equal(serviceCommitUntrackedFileAdded))

				content := utils.SucceedCommandOutputString(
					ctx,
					sourceWorkTreeDir,
					"git",
					"show", serviceCommit+":"+trackedFileRelPath,
				)

				Expect(content).Should(Equal(string(trackedFileContent2)))
			})

			It("stage", func(ctx context.Context) {
				ctx = logging.WithLogger(ctx)

				utils.RunSucceedCommand(
					ctx,
					sourceWorkTreeDir,
					"git",
					"add", trackedFilePath,
				)

				serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)

				Expect(err).Should(Succeed())
				Expect(serviceCommit).Should(Equal(serviceCommitUntrackedFileAdded))

				content := utils.SucceedCommandOutputString(
					ctx,
					sourceWorkTreeDir,
					"git",
					"show", serviceCommit+":"+trackedFileRelPath,
				)

				Expect(content).Should(Equal(string(trackedFileContent1)))
			})

			It("delete", func(ctx context.Context) {
				ctx = logging.WithLogger(ctx)

				utils.DeleteFile(trackedFilePath)

				serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)

				Expect(err).Should(Succeed())
				Expect(serviceCommit).ShouldNot(Equal(serviceCommitUntrackedFileAdded))

				bytes, err := utils.RunCommand(
					ctx,
					sourceWorkTreeDir,
					"git",
					"show", serviceCommit+":"+trackedFileRelPath,
				)
				Expect(err).Should(HaveOccurred())

				output := string(bytes)
				Expect(output).Should(ContainSubstring(fmt.Sprintf("'%s' does not exist in '%s'", trackedFileRelPath, serviceCommit)))
			})

			It("staged and synced, then modified, staged and synced: service branch contains last main branch commit", func(ctx context.Context) {
				ctx = logging.WithLogger(ctx)

				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "add", trackedFilePath)

				_, err := SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)
				Expect(err).Should(Succeed())

				utils.WriteFile(trackedFilePath, trackedFileContent2)

				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "add", trackedFilePath)

				_, err = SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)
				Expect(err).Should(Succeed())

				utils.RunSucceedCommand(ctx, filepath.Join(workTreeCacheDir, "worktree"), "git", "merge-base", "--is-ancestor", "main", "HEAD")
			})

			It("staged and synced, then committed and synced: service branch contains last main branch commit", func(ctx context.Context) {
				ctx = logging.WithLogger(ctx)

				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "add", trackedFilePath)

				_, err := SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)
				Expect(err).Should(Succeed())

				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "commit", "-m", "1")
				sourceHeadCommit = utils.GetHeadCommit(ctx, sourceWorkTreeDir)

				_, err = SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)
				Expect(err).Should(Succeed())

				utils.RunSucceedCommand(ctx, filepath.Join(workTreeCacheDir, "worktree"), "git", "merge-base", "--is-ancestor", "main", "HEAD")
			})

			It("try to trigger a merge conflict: merge conflict not happening", func(ctx context.Context) {
				ctx = logging.WithLogger(ctx)

				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "add", ".")
				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "commit", "-m", "1")
				sourceHeadCommit = utils.GetHeadCommit(ctx, sourceWorkTreeDir)

				_, err := SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)
				Expect(err).Should(Succeed())

				trackedFilePathMoved := fmt.Sprintf("%s-moved", trackedFilePath)

				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "mv", trackedFilePath, trackedFilePathMoved)
				utils.WriteFile(trackedFilePathMoved, trackedFileContent2)

				_, err = SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)
				Expect(err).Should(Succeed())

				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "reset", "--hard")
				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "clean", "-f")

				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "mv", trackedFilePath, trackedFilePathMoved)
				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "add", ".")
				utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "commit", "-m", "2")
				sourceHeadCommit = utils.GetHeadCommit(ctx, sourceWorkTreeDir)

				_, err = SyncSourceWorktreeWithServiceBranch(
					ctx,
					gitDir,
					sourceWorkTreeDir,
					workTreeCacheDir,
					sourceHeadCommit,
					defaultOptions,
				)
				Expect(err).Should(Succeed())
			})
		})
	})

	When("glob exclude specified", func() {
		const untrackedFileRelPath = "untracked_file.ext"
		var untrackedFilePath string
		var serviceCommitUntrackedFileAdded string

		BeforeEach(func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			untrackedFilePath = filepath.Join(sourceWorkTreeDir, untrackedFileRelPath)
			utils.WriteFile(untrackedFilePath, []byte("any"))

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
				ctx,
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				sourceHeadCommit,
				defaultOptions,
			)

			Expect(err).Should(Succeed())

			serviceCommitUntrackedFileAdded = serviceCommit
		})

		It("not ignore", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			defaultOptions.GlobExcludeList = []string{"file"}

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
				ctx,
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				sourceHeadCommit,
				defaultOptions,
			)

			Expect(err).Should(Succeed())
			Expect(serviceCommit).Should(Equal(serviceCommitUntrackedFileAdded))
		})

		It("ignore", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			defaultOptions.GlobExcludeList = []string{"*.ext"}

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(
				ctx,
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				sourceHeadCommit,
				defaultOptions,
			)

			Expect(err).Should(Succeed())
			Expect(serviceCommit).ShouldNot(Equal(serviceCommitUntrackedFileAdded))

			bytes, err := utils.RunCommand(
				ctx,
				sourceWorkTreeDir,
				"git",
				"show", serviceCommit+":"+untrackedFileRelPath,
			)
			Expect(err).Should(HaveOccurred())

			output := string(bytes)
			Expect(output).Should(ContainSubstring(fmt.Sprintf("'%s' exists on disk, but not in '%s'", untrackedFileRelPath, serviceCommit)))
		})
	})
})
