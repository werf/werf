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

		Ω(Init(Options{})).Should(Succeed())
	})

	When("git repo contains broken ref", func() {
		It("does not ignore git cmd warnings in output due to bug", func() {
			ctx := context.Background()

			brokenHeadPath := filepath.Join(gitRepoPath, ".git", "refs", "heads", "foo")
			Expect(os.WriteFile(brokenHeadPath, []byte("bad"), os.ModePerm)).To(Succeed())

			output, err := runGitCmd(ctx, []string{"branch", "--list", "no-such-branch"}, gitRepoPath, runGitCmdOptions{})
			Expect(err).To(Succeed())
			Expect(output.Len()).NotTo(Equal(0))
		})
	})
})
