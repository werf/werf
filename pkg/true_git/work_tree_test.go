package true_git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("Work tree helpers", func() {
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

				utils.RunSucceedCommand(ctx, mainWtDir, "git", "commit", "--allow-empty", "-m", "Initial commit")

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

	Describe("verifyWorkTreeConsistency", func() {
		var mainWtDir, sideWtDir string
		BeforeEach(func(ctx SpecContext) {
			mainWtDir = filepath.Join(SuiteData.TestDirPath, "main-wt")
			sideWtDir = filepath.Join(SuiteData.TestDirPath, "side-wt")

			Expect(os.MkdirAll(mainWtDir, os.ModePerm)).To(Succeed())

			utils.RunSucceedCommand(ctx, mainWtDir, "git", "-c", "init.defaultBranch=main", "init")

			utils.RunSucceedCommand(ctx, mainWtDir, "git", "checkout", "-b", "main")

			utils.RunSucceedCommand(ctx, mainWtDir, "git", "commit", "--allow-empty", "-m", "Initial commit")

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
