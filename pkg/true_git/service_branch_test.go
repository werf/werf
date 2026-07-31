package true_git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

func cleanFilterInvocationCount(counterPath string) int {
	data, err := os.ReadFile(counterPath)
	if err != nil {
		return 0
	}
	return len(data)
}

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

		gitCommitSucceed(ctx, sourceWorkTreeDir, "--allow-empty", "-m", "Initial commit")

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

				gitCommitSucceed(ctx, sourceWorkTreeDir, "-m", "1")
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
				gitCommitSucceed(ctx, sourceWorkTreeDir, "-m", "1")
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
				gitCommitSucceed(ctx, sourceWorkTreeDir, "-m", "2")
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

	When("a clean filter is configured", func() {
		const bigFileRelPath = "bigfile"
		var counterPath string
		syncOptions := SyncSourceWorktreeWithServiceBranchOptions{ServiceBranch: "_werf-dev"}

		BeforeEach(func(ctx SpecContext) {
			counterPath = filepath.Join(SuiteData.TestDirPath, "clean_count")
			utils.WriteFile(filepath.Join(sourceWorkTreeDir, ".gitattributes"), []byte(bigFileRelPath+" filter=count\n"))
			bigFilePath := filepath.Join(sourceWorkTreeDir, bigFileRelPath)
			utils.WriteFile(bigFilePath, []byte("payload"))
			// Push mtime into the past so the first sync's index is not "racily clean":
			// git re-hashes entries whose mtime is not older than the index timestamp, which
			// would otherwise re-invoke the clean filter on the unchanged second sync.
			past := time.Unix(1, 0)
			Expect(os.Chtimes(bigFilePath, past, past)).Should(Succeed())
			utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "config", "filter.count.clean", fmt.Sprintf("printf x >> '%s'; cat", counterPath))
		})

		It("does not re-run the clean filter on an unchanged second sync (#4754)", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			serviceCommit1, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			countAfter1 := cleanFilterInvocationCount(counterPath)
			Expect(countAfter1).Should(BeNumerically(">", 0))

			serviceCommit2, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(serviceCommit2).Should(Equal(serviceCommit1))
			Expect(cleanFilterInvocationCount(counterPath)).Should(Equal(countAfter1))

			utils.WriteFile(filepath.Join(sourceWorkTreeDir, bigFileRelPath), []byte("payload-modified"))

			serviceCommit3, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(serviceCommit3).ShouldNot(Equal(serviceCommit1))
			Expect(cleanFilterInvocationCount(counterPath)).Should(BeNumerically(">", countAfter1))
		})

		It("reseeds when an untracked .gitattributes newly filters a stat-unchanged file", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			// Filter config pre-exists; only an UNTRACKED .gitattributes appears between syncs, so
			// the reseed must come from hashing worktree (not just index-tracked) .gitattributes.
			utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "config", "filter.up.clean", "tr 'a-z' 'A-Z'")

			const fRel = "cased_file"
			fPath := filepath.Join(sourceWorkTreeDir, fRel)
			utils.WriteFile(fPath, []byte("abc"))
			utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "add", fRel)
			gitCommitSucceed(ctx, sourceWorkTreeDir, "-m", "add cased_file")
			sourceHeadCommit = utils.GetHeadCommit(ctx, sourceWorkTreeDir)
			past := time.Unix(1, 0)
			Expect(os.Chtimes(fPath, past, past)).Should(Succeed())

			serviceCommit1, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "show", serviceCommit1+":"+fRel)).Should(Equal("abc"))

			// Only create the untracked .gitattributes; do NOT touch fPath or git config.
			utils.WriteFile(filepath.Join(sourceWorkTreeDir, ".gitattributes"), []byte(bigFileRelPath+" filter=count\n"+fRel+" filter=up\n"))

			serviceCommit2, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "show", serviceCommit2+":"+fRel)).Should(Equal("ABC"))
		})

		It("reseeds when the global gitattributes file newly filters a stat-unchanged file", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			// git var GIT_ATTR_GLOBAL honors XDG_CONFIG_HOME (git >= 2.43); on older git the
			// signature falls back to the same XDG default, so both paths are exercised here.
			// Isolate from the host's real global config, which could set core.attributesFile and
			// redirect GIT_ATTR_GLOBAL away from our XDG file.
			xdg := filepath.Join(SuiteData.TestDirPath, "xdg")
			globalAttrs := filepath.Join(xdg, "git", "attributes")
			utils.MkdirAll(filepath.Dir(globalAttrs))
			utils.WriteFile(globalAttrs, []byte(""))
			GinkgoT().Setenv("XDG_CONFIG_HOME", xdg)
			emptyGlobalConfig := filepath.Join(SuiteData.TestDirPath, "empty_global_gitconfig")
			utils.WriteFile(emptyGlobalConfig, []byte(""))
			GinkgoT().Setenv("GIT_CONFIG_GLOBAL", emptyGlobalConfig)

			utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "config", "filter.up.clean", "tr 'a-z' 'A-Z'")

			const fRel = "global_cased"
			fPath := filepath.Join(sourceWorkTreeDir, fRel)
			utils.WriteFile(fPath, []byte("abc"))
			utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "add", fRel)
			gitCommitSucceed(ctx, sourceWorkTreeDir, "-m", "add global_cased")
			sourceHeadCommit = utils.GetHeadCommit(ctx, sourceWorkTreeDir)
			past := time.Unix(1, 0)
			Expect(os.Chtimes(fPath, past, past)).Should(Succeed())

			serviceCommit1, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "show", serviceCommit1+":"+fRel)).Should(Equal("abc"))

			// Change ONLY the global attributes file content; repo config and file stat unchanged.
			utils.WriteFile(globalAttrs, []byte(fRel+" filter=up\n"))

			serviceCommit2, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "show", serviceCommit2+":"+fRel)).Should(Equal("ABC"))
		})

		It("includes the system gitattributes path in the conversion signature (git >= 2.43)", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)
			if gitVersion.LessThan(semver.MustParse("2.43.0")) {
				Skip("git var GIT_ATTR_SYSTEM requires git >= 2.43")
			}

			systemPath := strings.TrimSpace(utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "var", "GIT_ATTR_SYSTEM"))
			if systemPath == "" {
				Skip("system gitattributes disabled (GIT_ATTR_NOSYSTEM)")
			}
			Expect(globalAndSystemAttributesFilePaths(ctx, sourceWorkTreeDir)).Should(ContainElement(systemPath))
		})
	})

	When("the persistent dev-index is corrupted", func() {
		syncOptions := SyncSourceWorktreeWithServiceBranchOptions{ServiceBranch: "_werf-dev"}

		It("self-heals and produces the same commit", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)
			utils.WriteFile(filepath.Join(sourceWorkTreeDir, "tracked_file"), []byte("state"))

			serviceCommit1, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())

			utils.WriteFile(filepath.Join(workTreeCacheDir, "dev_index"), []byte("corrupt-not-an-index"))

			serviceCommit2, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(serviceCommit2).Should(Equal(serviceCommit1))
		})

		It("self-heals past a stale dev_index.lock left by a killed run", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)
			utils.WriteFile(filepath.Join(sourceWorkTreeDir, "tracked_file"), []byte("state"))

			serviceCommit1, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())

			// A SIGKILLed git leaves this behind; every later index op would fail to lock.
			utils.WriteFile(filepath.Join(workTreeCacheDir, "dev_index.lock"), []byte("stale"))

			serviceCommit2, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(serviceCommit2).Should(Equal(serviceCommit1))
		})

		It("self-heals past a stale dev_index.lock left by a killed COLD run", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)
			utils.WriteFile(filepath.Join(sourceWorkTreeDir, "tracked_file"), []byte("state"))

			// No prior successful sync: marker absent ⇒ next sync takes the cold seed path, where
			// the warm retry-with-rebuild does not run. A stale lock here must still be cleared.
			utils.WriteFile(filepath.Join(workTreeCacheDir, "dev_index.lock"), []byte("stale"))

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(serviceCommit).ShouldNot(Equal(sourceHeadCommit))
		})
	})

	When("a failing repository pre-commit hook is installed", func() {
		syncOptions := SyncSourceWorktreeWithServiceBranchOptions{ServiceBranch: "_werf-dev"}

		It("does not run the hook during sync", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)
			utils.MkdirAll(filepath.Join(gitDir, "hooks"))
			hookPath := filepath.Join(gitDir, "hooks", "pre-commit")
			utils.WriteFile(hookPath, []byte("#!/bin/sh\nexit 1\n"))
			Expect(os.Chmod(hookPath, 0o755)).Should(Succeed())

			utils.WriteFile(filepath.Join(sourceWorkTreeDir, "tracked_file"), []byte("state"))

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(serviceCommit).ShouldNot(Equal(sourceHeadCommit))
		})
	})

	When("paths contain spaces and unicode", func() {
		syncOptions := SyncSourceWorktreeWithServiceBranchOptions{ServiceBranch: "_werf-dev"}

		It("captures them into the service commit", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)
			const rel = "dir with space/файл.txt"
			utils.MkdirAll(filepath.Join(sourceWorkTreeDir, "dir with space"))
			utils.WriteFile(filepath.Join(sourceWorkTreeDir, rel), []byte("content"))

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			Expect(utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "show", serviceCommit+":"+rel)).Should(Equal("content"))
		})
	})

	When("the source worktree has a submodule", func() {
		syncOptions := SyncSourceWorktreeWithServiceBranchOptions{ServiceBranch: "_werf-dev"}

		It("records the submodule gitlink in the service commit", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			subRepoDir := filepath.Join(SuiteData.TestDirPath, "submodule-src")
			utils.MkdirAll(subRepoDir)
			utils.RunSucceedCommand(ctx, subRepoDir, "git", "-c", "init.defaultBranch=main", "init")
			utils.WriteFile(filepath.Join(subRepoDir, "subfile"), []byte("sub"))
			utils.RunSucceedCommand(ctx, subRepoDir, "git", "add", ".")
			gitCommitSucceed(ctx, subRepoDir, "-m", "sub init")
			subHead := utils.GetHeadCommit(ctx, subRepoDir)

			// git blocks file:// submodule transport by default (CVE-2022-39253). werf's internal
			// `git submodule update` runs without -c overrides but inherits our env, so allow it via
			// GIT_CONFIG_* which git applies process-wide including to spawned git.
			GinkgoT().Setenv("GIT_CONFIG_COUNT", "1")
			GinkgoT().Setenv("GIT_CONFIG_KEY_0", "protocol.file.allow")
			GinkgoT().Setenv("GIT_CONFIG_VALUE_0", "always")

			utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "submodule", "add", subRepoDir, "sub")
			gitCommitSucceed(ctx, sourceWorkTreeDir, "-m", "add submodule")
			sourceHeadCommit = utils.GetHeadCommit(ctx, sourceWorkTreeDir)

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())

			lsTree := utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "ls-tree", serviceCommit, "sub")
			Expect(lsTree).Should(ContainSubstring("commit " + subHead))
		})
	})

	When("validating the service tree against a full git-add snapshot", func() {
		syncOptions := SyncSourceWorktreeWithServiceBranchOptions{ServiceBranch: "_werf-dev"}

		It("equals an independent full snapshot even with staged and unstaged changes", func(ctx context.Context) {
			ctx = logging.WithLogger(ctx)

			utils.WriteFile(filepath.Join(sourceWorkTreeDir, "staged_file"), []byte("staged"))
			utils.RunSucceedCommand(ctx, sourceWorkTreeDir, "git", "add", "staged_file")
			utils.WriteFile(filepath.Join(sourceWorkTreeDir, "unstaged_file"), []byte("unstaged"))

			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(ctx, gitDir, sourceWorkTreeDir, workTreeCacheDir, sourceHeadCommit, syncOptions)
			Expect(err).Should(Succeed())
			serviceTree := utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "rev-parse", serviceCommit+"^{tree}")

			// Build the oracle with a disposable index so the user's real index is untouched.
			oracleEnv := []string{"GIT_INDEX_FILE=" + filepath.Join(SuiteData.TestDirPath, "oracle_index")}
			runOracle := func(args ...string) string {
				res, err := utils.RunCommandWithOptions(ctx, sourceWorkTreeDir, "git", args, utils.RunCommandOptions{ExtraEnv: oracleEnv, ShouldSucceed: true, NoStderr: true})
				Expect(err).Should(Succeed())
				return string(res)
			}
			runOracle("read-tree", sourceHeadCommit)
			runOracle("add", "--", ":.")
			oracleTree := runOracle("write-tree")

			Expect(strings.TrimSpace(serviceTree)).Should(Equal(strings.TrimSpace(oracleTree)))

			// staged content is in the snapshot; the user index still shows staged_file staged.
			Expect(utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "show", serviceCommit+":staged_file")).Should(Equal("staged"))
			Expect(utils.SucceedCommandOutputString(ctx, sourceWorkTreeDir, "git", "show", serviceCommit+":unstaged_file")).Should(Equal("unstaged"))
		})
	})
})
