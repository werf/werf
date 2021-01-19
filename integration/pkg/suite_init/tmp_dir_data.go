package suite_init

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

type TmpDirData struct {
	TmpDir      string
	TestDirPath string
}

func NewTmpDirData() *TmpDirData {
	data := &TmpDirData{}
	SetupTmpDir(&data.TmpDir, &data.TestDirPath)
	return data
}

func SetupTmpDir(tmpDir, testDirPath *string) bool {
	ginkgo.BeforeEach(func() {
		*tmpDir = utils.GetTempDir()
		*testDirPath = *tmpDir
	})

	ginkgo.AfterEach(func() {
		err := os.RemoveAll(*tmpDir)
		Ω(err).ShouldNot(HaveOccurred())
	})

	return true
}

func (data *TmpDirData) GetProjectWorktree(projectName string) string {
	return filepath.Join(data.TestDirPath, fmt.Sprintf("%s.worktree", projectName))
}

func (data *TmpDirData) CommitProjectWorktree(projectName, worktreeFixtureDir, commitMessage string) {
	worktreeDir := data.GetProjectWorktree(projectName)
	repoDir := filepath.Join(data.TestDirPath, fmt.Sprintf("%s.repo", projectName))

	gomega.Expect(os.RemoveAll(worktreeDir)).To(Succeed())
	utils.CopyIn(worktreeFixtureDir, worktreeDir)
	gomega.Expect(utils.SetGitRepoState(worktreeDir, repoDir, commitMessage)).To(gomega.Succeed())
}
