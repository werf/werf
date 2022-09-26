package e2e_converge_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/test/pkg/suite_init"
	"github.com/werf/werf/test/pkg/utils"
)

func TestSuite(t *testing.T) {
	suite_init.MakeTestSuiteEntrypointFunc("E2E converge suite", suite_init.TestSuiteEntrypointFuncOptions{
		RequiredSuiteTools: []string{"docker", "git"},
		RequiredSuiteEnvs: []string{
			"WERF_TEST_K8S_DOCKER_REGISTRY",
		},
	})(t)
}

var SuiteData suite_init.SuiteData

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())

	_ = SuiteData.SetupK8sDockerRegistry(suite_init.NewK8sDockerRegistryData(SuiteData.ProjectNameData, SuiteData.StubsData))

	_ = BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	_ = AfterEach(func() {
		utils.RunSucceedCommand("", SuiteData.WerfBinPath, "host", "purge", "--force", "--project-name", SuiteData.ProjectName)
	})
)
