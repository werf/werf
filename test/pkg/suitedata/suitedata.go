package suitedata

import (
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/suite_init"
	"github.com/werf/werf/integration/pkg/utils"
)

type SuiteData struct {
	suite_init.SuiteData
}

func (s *SuiteData) GetTestRepoPath(dirname string) string {
	testReposDir := filepath.Join(s.TestDirPath, "repos")
	Expect(os.MkdirAll(testReposDir, os.ModePerm)).To(Succeed())
	return filepath.Join(testReposDir, dirname)
}

func (s *SuiteData) GetBuildReportPath(filename string) string {
	buildReportsDir := filepath.Join(s.TestDirPath, "build-reports")
	Expect(os.MkdirAll(buildReportsDir, os.ModePerm)).To(Succeed())
	return filepath.Join(buildReportsDir, filename)
}

func (s *SuiteData) InitTestRepo(dirname, fixtureRelPath string) {
	testRepoPath := s.GetTestRepoPath(dirname)
	utils.CopyIn(utils.FixturePath(fixtureRelPath), testRepoPath)
	utils.RunSucceedCommand(testRepoPath, "git", "init")
	utils.RunSucceedCommand(testRepoPath, "git", "add", ".")
	utils.RunSucceedCommand(testRepoPath, "git", "commit", "-m", "initial")
}

func (s *SuiteData) UpdateTestRepo(dirname, fixtureRelPath string) {
	testRepoPath := s.GetTestRepoPath(dirname)
	utils.RunSucceedCommand(testRepoPath, "git", "rm", "--ignore-unmatch", "-rf", ".")
	utils.CopyIn(utils.FixturePath(fixtureRelPath), testRepoPath)
	utils.RunSucceedCommand(testRepoPath, "git", "add", ".")
	utils.RunSucceedCommand(testRepoPath, "git", "commit", "-m", "updated")
}
