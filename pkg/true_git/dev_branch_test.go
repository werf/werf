package true_git

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/pkg/werf"
)

var _ = Describe("SyncSourceWorktreeWithServiceWorktreeBranch", func() {
	var sourceWorkTreeDir string
	var gitDir string
	var workTreeCacheDir string
	var initialCommit string
	defaultOptions := SyncSourceWorktreeWithServiceWorktreeBranchOptions{ServiceBranchPrefix: "dev-"}

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

		initialCommit = utils.GetHeadCommit(sourceWorkTreeDir)

		Ω(werf.Init("", "")).Should(Succeed())
	})

	It("no changes", func() {
		commit, err := SyncSourceWorktreeWithServiceWorktreeBranch(
			context.Background(),
			gitDir,
			sourceWorkTreeDir,
			workTreeCacheDir,
			initialCommit,
			defaultOptions,
		)

		Ω(err).Should(Succeed())
		Ω(commit).Should(Equal(initialCommit))
	})

	When("tracked changes", func() {
		const trackedFileRelPath = "tracked_file"
		var trackedFilePath string

		BeforeEach(func() {
			trackedFilePath = filepath.Join(sourceWorkTreeDir, trackedFileRelPath)
			utils.CreateFile(trackedFilePath, []byte("state"))

			utils.RunSucceedCommand(
				sourceWorkTreeDir,
				"git",
				"add", trackedFilePath,
			)
		})

		It("add", func() {
			serviceCommit, err := SyncSourceWorktreeWithServiceWorktreeBranch(
				context.Background(),
				gitDir,
				sourceWorkTreeDir,
				workTreeCacheDir,
				initialCommit,
				defaultOptions,
			)

			Ω(err).Should(Succeed())
			Ω(serviceCommit).ShouldNot(Equal(initialCommit))

			diff := utils.SucceedCommandOutputString(
				sourceWorkTreeDir,
				"git",
				"diff", serviceCommit, trackedFileRelPath,
			)

			Ω(diff).Should(BeEmpty())
		})
	})
})
