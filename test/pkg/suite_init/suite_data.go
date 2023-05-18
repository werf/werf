package suite_init

import (
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

type SuiteData struct {
	*StubsData
	*SynchronizedSuiteCallbacksData
	*WerfBinaryData
	*ProjectNameData
	*K8sDockerRegistryData
	*TmpDirData
	*ContainerRegistryPerImplementationData
}

func (data *SuiteData) SetupStubs(setupData *StubsData) bool {
	data.StubsData = setupData
	return true
}

func (data *SuiteData) SetupSynchronizedSuiteCallbacks(setupData *SynchronizedSuiteCallbacksData) bool {
	data.SynchronizedSuiteCallbacksData = setupData
	return true
}

func (data *SuiteData) SetupWerfBinary(setupData *WerfBinaryData) bool {
	data.WerfBinaryData = setupData
	return true
}

func (data *SuiteData) SetupProjectName(setupData *ProjectNameData) bool {
	data.ProjectNameData = setupData
	return true
}

func (data *SuiteData) SetupK8sDockerRegistry(setupData *K8sDockerRegistryData) bool {
	data.K8sDockerRegistryData = setupData
	return true
}

func (data *SuiteData) SetupTmp(setupData *TmpDirData) bool {
	data.TmpDirData = setupData
	return true
}

func (data *SuiteData) SetupContainerRegistryPerImplementation(setupData *ContainerRegistryPerImplementationData) bool {
	data.ContainerRegistryPerImplementationData = setupData
	return true
}

func (data *SuiteData) GetTestRepoPath(dirname string) string {
	testReposDir := filepath.Join(data.TestDirPath, "repos")
	Expect(os.MkdirAll(testReposDir, os.ModePerm)).To(Succeed())
	return filepath.Join(testReposDir, dirname)
}

func (data *SuiteData) GetBuildReportPath(filename string) string {
	buildReportsDir := filepath.Join(data.TestDirPath, "build-reports")
	Expect(os.MkdirAll(buildReportsDir, os.ModePerm)).To(Succeed())
	return filepath.Join(buildReportsDir, filename)
}

func (data *SuiteData) GetDeployReportPath(filename string) string {
	deployReportsDir := filepath.Join(data.TestDirPath, "deploy-reports")
	Expect(os.MkdirAll(deployReportsDir, os.ModePerm)).To(Succeed())
	return filepath.Join(deployReportsDir, filename)
}

func (data *SuiteData) InitTestRepo(dirname, fixtureRelPath string) {
	testRepoPath := data.GetTestRepoPath(dirname)
	utils.CopyIn(utils.FixturePath(fixtureRelPath), testRepoPath)
	utils.RunSucceedCommand(testRepoPath, "git", "init")
	utils.RunSucceedCommand(testRepoPath, "git", "add", ".")
	utils.RunSucceedCommand(testRepoPath, "git", "commit", "-m", "initial")
}

func (data *SuiteData) UpdateTestRepo(dirname, fixtureRelPath string) {
	testRepoPath := data.GetTestRepoPath(dirname)
	utils.RunSucceedCommand(testRepoPath, "git", "rm", "--ignore-unmatch", "-rf", ".")
	utils.CopyIn(utils.FixturePath(fixtureRelPath), testRepoPath)
	utils.RunSucceedCommand(testRepoPath, "git", "add", ".")
	utils.RunSucceedCommand(testRepoPath, "git", "commit", "-m", "updated")
}
