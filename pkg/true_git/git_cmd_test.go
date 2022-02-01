package true_git

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("Git command", func() {
	var gitRepoPath string

	BeforeEach(func() {
		gitRepoPath = filepath.Join(SuiteData.TestDirPath, "repo")
		utils.MkdirAll(gitRepoPath)

		utils.RunSucceedCommand(
			gitRepoPath,
			"git",
			"init",
		)

		utils.RunSucceedCommand(
			gitRepoPath,
			"git",
			"commit", "--allow-empty", "-m", "Initial commit",
		)

		Ω(Init(context.Background(), Options{})).Should(Succeed())
	})

	When("looking for existent ref", func() {
		It("succeeds, populates stdout/err/outerr buffers correctly", func() {
			ctx := context.Background()

			brokenHeadPath := filepath.Join(gitRepoPath, ".git", "refs", "heads", "broken")
			Expect(os.WriteFile(brokenHeadPath, []byte("invalid"), os.ModePerm)).To(Succeed())

			branchCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: gitRepoPath}, "branch", "--list", "master")
			err := branchCmd.Run(ctx)
			Expect(err).To(Succeed())
			Expect(branchCmd.OutBuf.String()).To(Equal("* master\n"))
			Expect(branchCmd.ErrBuf.String()).To(Equal("warning: ignoring broken ref refs/heads/broken\n"))
			Expect(branchCmd.OutErrBuf.String()).To(Equal("warning: ignoring broken ref refs/heads/broken\n* master\n"))
		})
	})
})
