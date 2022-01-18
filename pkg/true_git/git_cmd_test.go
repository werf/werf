package true_git

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
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
		It("succeeds, returning branch name", func() {
			ctx := context.Background()

			output, err := runGitCmd(ctx, []string{"branch", "--list", "master"}, gitRepoPath, runGitCmdOptions{})
			Expect(err).To(Succeed())
			Expect(output.String()).To(ContainSubstring("master"))
		})
	})

	When("looking for non-existent ref", func() {
		It("succeeds, ignoring stderr in git output and returning only (empty) stdout", func() {
			ctx := context.Background()

			brokenHeadPath := filepath.Join(gitRepoPath, ".git", "refs", "heads", "foo")
			Expect(os.WriteFile(brokenHeadPath, []byte("bad"), os.ModePerm)).To(Succeed())

			output, err := runGitCmd(ctx, []string{"branch", "--list", "no-such-branch"}, gitRepoPath, runGitCmdOptions{})
			Expect(err).To(Succeed())
			Expect(output.Len()).To(Equal(0))
		})
	})
})
