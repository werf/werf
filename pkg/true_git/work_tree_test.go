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

	"github.com/werf/common-go/pkg/graceful"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("Work tree helpers", func() {
	BeforeEach(func(ctx SpecContext) {
		Expect(Init(ctx, Options{})).To(Succeed())
	})

	DescribeTable("lists worktrees with Git before 2.36",
		func(ctx SpecContext, mainName, linkedName, access string, wantError bool) {
			mainDir := filepath.Join(SuiteData.TestDirPath, mainName)
			linkedDir := filepath.Join(SuiteData.TestDirPath, linkedName)
			Expect(os.MkdirAll(mainDir, 0o755)).To(Succeed())
			utils.RunSucceedCommand(ctx, mainDir, "git", "init")
			gitCommitSucceed(ctx, mainDir, "--allow-empty", "-m", "Initial commit")
			utils.RunSucceedCommand(ctx, mainDir, "git", "worktree", "add", "--detach", linkedDir)
			repoDir := mainDir
			if access == "symlink" {
				repoDir = filepath.Join(SuiteData.TestDirPath, "source-link")
				Expect(os.Symlink(mainDir, repoDir)).To(Succeed())
			} else if access == "relative" {
				originalDir, err := os.Getwd()
				Expect(err).To(Succeed())
				DeferCleanup(func() { Expect(os.Chdir(originalDir)).To(Succeed()) })
				Expect(os.Chdir(mainDir)).To(Succeed())
				repoDir = "."
			}

			originalVersion := gitVersion
			DeferCleanup(func() { gitVersion = originalVersion })
			gitVersion = semver.MustParse("2.35.0")

			list, err := GetWorkTreeList(ctx, repoDir)
			if wantError {
				Expect(err).To(MatchError(ContainSubstring("Git >= 2.36 required for worktree paths containing newlines")))
				return
			}
			Expect(err).To(Succeed())
			Expect(list).To(ConsistOf(
				HaveField("Path", mainDir),
				SatisfyAll(HaveField("Path", linkedDir), HaveField("Detached", BeTrue())),
			))
		},
		Entry("ordinary paths", "main", "linked", "direct", false),
		Entry("ordinary paths reached through a symlink", "main", "linked", "symlink", false),
		Entry("ordinary paths reached from the current directory", "main", "linked", "relative", false),
		Entry("newline in main path", "main\nworktree", "linked", "direct", true),
		Entry("newline in main path reached through a symlink", "main\nworktree", "linked", "symlink", true),
		Entry("newline in main path reached from the current directory", "main\nworktree", "linked", "relative", true),
		Entry("newline in linked path", "main", "linked\nworktree", "direct", true),
	)

	Describe("resolveDotGitFile", func() {
		It("parses correctly formatted dot git link file", func(ctx SpecContext) {
			linkFile := filepath.Join(SuiteData.TestDirPath, ".git")

			targetPath := "/path/to/target/git"

			Expect(os.WriteFile(linkFile, []byte(fmt.Sprintf("gitdir: %s\n", targetPath)), 0o644)).To(Succeed())

			resPath, err := resolveDotGitFile(ctx, linkFile)
			Expect(err).To(Succeed())

			Expect(resPath).To(Equal(targetPath))
		})

		It("fails to parse invalid dot git link file", func(ctx SpecContext) {
			linkFile := filepath.Join(SuiteData.TestDirPath, ".git")

			Expect(os.WriteFile(linkFile, []byte("invalid"), 0o644)).To(Succeed())

			_, err := resolveDotGitFile(ctx, linkFile)
			Expect(err).To(Equal(ErrInvalidDotGit))
		})
	})

	When("side worktree was previously added and locked", func() {
		When("no submodules are used", func() {
			var mainWtDir, sideWtDir string

			BeforeEach(func(ctx SpecContext) {
				mainWtDir = filepath.Join(SuiteData.TestDirPath, "main-wt")
				sideWtDir = filepath.Join(SuiteData.TestDirPath, "side-wt")

				Expect(os.MkdirAll(mainWtDir, os.ModePerm)).To(Succeed())

				utils.RunSucceedCommand(ctx, mainWtDir, "git", "-c", "init.defaultBranch=main", "init")

				utils.RunSucceedCommand(ctx, mainWtDir, "git", "checkout", "-b", "main")

				gitCommitSucceed(ctx, mainWtDir, "--allow-empty", "-m", "Initial commit")

				utils.RunSucceedCommand(ctx, mainWtDir, "git", "worktree", "add", sideWtDir)

				utils.RunSucceedCommand(ctx, mainWtDir, "git", "worktree", "lock", sideWtDir)

				err := os.RemoveAll(sideWtDir)
				Expect(err).To(Succeed())
			})

			It("should replace worktree without errors", func(ctx SpecContext) {
				commit := getHeadCommit(ctx, mainWtDir)

				Expect(switchWorkTree(ctx, mainWtDir, sideWtDir, commit, false)).To(Succeed())
			})
		})
	})

	When("a stale index.lock is left in a cached worktree", func() {
		var mainWtDir, sideWtDir string

		BeforeEach(func(ctx SpecContext) {
			mainWtDir = filepath.Join(SuiteData.TestDirPath, "main-wt")
			sideWtDir = filepath.Join(SuiteData.TestDirPath, "side-wt")

			Expect(os.MkdirAll(mainWtDir, os.ModePerm)).To(Succeed())
			utils.RunSucceedCommand(ctx, mainWtDir, "git", "-c", "init.defaultBranch=main", "init")
			utils.RunSucceedCommand(ctx, mainWtDir, "git", "checkout", "-b", "main")
			gitCommitSucceed(ctx, mainWtDir, "--allow-empty", "-m", "Initial commit")
			utils.RunSucceedCommand(ctx, mainWtDir, "git", "worktree", "add", "--detach", sideWtDir)
		})

		It("self-heals and switches the worktree", func(ctx SpecContext) {
			commit := getHeadCommit(ctx, mainWtDir)

			lockPath := strings.TrimSpace(utils.SucceedCommandOutputString(ctx, sideWtDir, "git", "rev-parse", "--git-path", "index.lock"))
			if !filepath.IsAbs(lockPath) {
				lockPath = filepath.Join(sideWtDir, lockPath)
			}
			Expect(os.WriteFile(lockPath, []byte("stale"), 0o644)).To(Succeed())

			Expect(switchWorkTree(ctx, mainWtDir, sideWtDir, commit, false)).To(Succeed())
		})
	})

	When("a cached worktree is broken beyond git's own checks", func() {
		var mainWtDir, workTreeCacheDir string

		BeforeEach(func(ctx SpecContext) {
			mainWtDir = filepath.Join(SuiteData.TestDirPath, "main-wt")
			workTreeCacheDir = filepath.Join(SuiteData.TestDirPath, "wt-cache")

			Expect(os.MkdirAll(mainWtDir, os.ModePerm)).To(Succeed())
			utils.RunSucceedCommand(ctx, mainWtDir, "git", "-c", "init.defaultBranch=main", "init")
			utils.RunSucceedCommand(ctx, mainWtDir, "git", "checkout", "-b", "main")
			gitCommitSucceed(ctx, mainWtDir, "--allow-empty", "-m", "Initial commit")
		})

		It("rebuilds the cached worktree from scratch and switches it to the commit", func(ctx SpecContext) {
			firstCommit := getHeadCommit(ctx, mainWtDir)

			workTreeDir, err := prepareWorkTree(ctx, mainWtDir, workTreeCacheDir, firstCommit, false)
			Expect(err).To(Succeed())

			gitCommitSucceed(ctx, mainWtDir, "--allow-empty", "-m", "Second commit")
			secondCommit := getHeadCommit(ctx, mainWtDir)
			Expect(secondCommit).NotTo(Equal(firstCommit))

			indexPath := strings.TrimSpace(utils.SucceedCommandOutputString(ctx, workTreeDir, "git", "rev-parse", "--git-path", "index"))
			if !filepath.IsAbs(indexPath) {
				indexPath = filepath.Join(workTreeDir, indexPath)
			}
			Expect(os.WriteFile(indexPath, []byte("garbage"), 0o644)).To(Succeed())

			consistent, err := verifyWorkTreeConsistency(ctx, mainWtDir, workTreeDir)
			Expect(err).To(Succeed())
			Expect(consistent).To(BeTrue())
			Expect(switchWorkTree(ctx, mainWtDir, workTreeDir, secondCommit, false)).NotTo(Succeed())

			healedWorkTreeDir, err := prepareWorkTree(ctx, mainWtDir, workTreeCacheDir, secondCommit, false)
			Expect(err).To(Succeed())
			Expect(healedWorkTreeDir).To(Equal(workTreeDir))
			Expect(getHeadCommit(ctx, workTreeDir)).To(Equal(secondCommit))
		})

		It("keeps a healthy cached worktree intact when the context is canceled mid-switch", func(ctx SpecContext) {
			firstCommit := getHeadCommit(ctx, mainWtDir)

			workTreeDir, err := prepareWorkTree(ctx, mainWtDir, workTreeCacheDir, firstCommit, false)
			Expect(err).To(Succeed())

			hookStartedPath := filepath.Join(SuiteData.TestDirPath, "hook-started")
			hookProceedPath := filepath.Join(SuiteData.TestDirPath, "hook-proceed")
			hookPath := filepath.Join(SuiteData.TestDirPath, "blocking-smudge.sh")
			hookScript := fmt.Sprintf("#!/bin/sh\ntouch %q\nfor i in $(seq 1 600); do [ -f %q ] && break; sleep 0.05; done\ncat\n", hookStartedPath, hookProceedPath)
			Expect(os.WriteFile(hookPath, []byte(hookScript), 0o755)).To(Succeed())

			utils.RunSucceedCommand(ctx, mainWtDir, "git", "config", "filter.block.smudge", hookPath)
			Expect(os.WriteFile(filepath.Join(mainWtDir, ".gitattributes"), []byte("blocked.txt filter=block\n"), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(mainWtDir, "blocked.txt"), []byte("v2"), 0o644)).To(Succeed())
			utils.RunSucceedCommand(ctx, mainWtDir, "git", "add", ".gitattributes", "blocked.txt")
			gitCommitSucceed(ctx, mainWtDir, "-m", "Second commit")
			secondCommit := getHeadCommit(ctx, mainWtDir)

			terminationCtx := graceful.WithTermination(ctx)
			helperDone := make(chan struct{})
			go func() {
				defer close(helperDone)
				for i := 0; i < 600; i++ {
					if _, err := os.Stat(hookStartedPath); err == nil {
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
				graceful.Terminate(terminationCtx, fmt.Errorf("sibling task failed"), 1)
				<-terminationCtx.Done()
				Expect(os.WriteFile(hookProceedPath, []byte("go"), 0o644)).To(Succeed())
			}()

			_, err = prepareWorkTree(terminationCtx, mainWtDir, workTreeCacheDir, secondCommit, false)
			Eventually(helperDone, "35s").Should(BeClosed())
			Expect(err).NotTo(Succeed())
			Expect(hookStartedPath).To(BeAnExistingFile(), "cancellation must have happened mid-switch")

			Expect(workTreeDir).To(BeADirectory())
			Expect(filepath.Join(workTreeDir, ".git")).To(BeAnExistingFile())
		})
	})

	When("the repository has a foreign worktree git reports as prunable", func() {
		var mainWtDir, workTreeCacheDir, foreignWtDir string

		BeforeEach(func(ctx SpecContext) {
			mainWtDir = filepath.Join(SuiteData.TestDirPath, "main-wt")
			workTreeCacheDir = filepath.Join(SuiteData.TestDirPath, "wt-cache")
			foreignWtDir = filepath.Join(SuiteData.TestDirPath, "foreign-wt")

			Expect(os.MkdirAll(mainWtDir, os.ModePerm)).To(Succeed())
			utils.RunSucceedCommand(ctx, mainWtDir, "git", "-c", "init.defaultBranch=main", "init")
			utils.RunSucceedCommand(ctx, mainWtDir, "git", "checkout", "-b", "main")
			gitCommitSucceed(ctx, mainWtDir, "--allow-empty", "-m", "Initial commit")
			utils.RunSucceedCommand(ctx, mainWtDir, "git", "worktree", "add", "--detach", foreignWtDir)

			_, err := prepareWorkTree(ctx, mainWtDir, workTreeCacheDir, getHeadCommit(ctx, mainWtDir), false)
			Expect(err).To(Succeed())
		})

		DescribeTable("keeps the foreign worktree registered",
			func(ctx SpecContext, makePrunable, restore func(dir string)) {
				makePrunable(foreignWtDir)
				defer restore(foreignWtDir)

				list, err := GetWorkTreeList(ctx, mainWtDir)
				Expect(err).To(Succeed())
				Expect(list).To(ContainElement(SatisfyAll(
					HaveField("Path", foreignWtDir),
					HaveField("Prunable", BeTrue()),
				)), "precondition: git must report the foreign worktree as prunable")

				_, err = prepareWorkTree(ctx, mainWtDir, workTreeCacheDir, getHeadCommit(ctx, mainWtDir), false)
				Expect(err).To(Succeed())

				list, err = GetWorkTreeList(ctx, mainWtDir)
				Expect(err).To(Succeed())
				Expect(list).To(ContainElement(HaveField("Path", foreignWtDir)))
			},
			Entry("when its directory was deleted", func(dir string) {
				Expect(os.RemoveAll(dir)).To(Succeed())
			}, func(string) {}),
			Entry("when its directory is inaccessible", func(dir string) {
				Expect(os.Chmod(dir, 0o000)).To(Succeed())
			}, func(dir string) {
				Expect(os.Chmod(dir, 0o755)).To(Succeed())
			}),
		)
	})

	Describe("verifyWorkTreeConsistency", func() {
		var mainWtDir, sideWtDir string
		BeforeEach(func(ctx SpecContext) {
			mainWtDir = filepath.Join(SuiteData.TestDirPath, "main-wt")
			sideWtDir = filepath.Join(SuiteData.TestDirPath, "side-wt")

			Expect(os.MkdirAll(mainWtDir, os.ModePerm)).To(Succeed())

			utils.RunSucceedCommand(ctx, mainWtDir, "git", "-c", "init.defaultBranch=main", "init")

			utils.RunSucceedCommand(ctx, mainWtDir, "git", "checkout", "-b", "main")

			gitCommitSucceed(ctx, mainWtDir, "--allow-empty", "-m", "Initial commit")

			utils.RunSucceedCommand(ctx, mainWtDir, "git", "worktree", "add", sideWtDir)
		})

		It("passes correct work tree", func(ctx SpecContext) {
			valid, err := verifyWorkTreeConsistency(ctx, mainWtDir, sideWtDir)
			Expect(err).To(Succeed())
			Expect(valid).To(BeTrue())
		})

		It("should return false,nil if got git file is removed in working tree dir", func(ctx SpecContext) {
			Expect(os.RemoveAll(filepath.Join(sideWtDir, ".git"))).To(Succeed())

			valid, err := verifyWorkTreeConsistency(ctx, mainWtDir, sideWtDir)
			Expect(err).To(Succeed())
			Expect(valid).To(BeFalse())
		})

		It("detects side work tree with incorrect back dot git link", func(ctx SpecContext) {
			Expect(os.WriteFile(filepath.Join(sideWtDir, ".git"), []byte(fmt.Sprintf("gitdir: %s\n", filepath.Join(mainWtDir, ".git", "worktrees", "no-such-worktree"))), os.ModePerm)).To(Succeed())

			valid, err := verifyWorkTreeConsistency(ctx, mainWtDir, sideWtDir)
			Expect(err).To(Succeed())
			Expect(valid).To(BeFalse())
		})
	})
})

func getHeadCommit(ctx context.Context, repoDir string) string {
	refs, err := ShowRef(ctx, repoDir)
	Expect(err).To(Succeed())

	for _, ref := range refs.Refs {
		if ref.IsHEAD {
			return ref.Commit
		}
	}

	Expect(fmt.Errorf("head commit not found")).NotTo(HaveOccurred())
	return ""
}
