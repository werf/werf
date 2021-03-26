package base_image_test

import (
	"strings"
	"testing"

	"github.com/werf/werf/integration/pkg/utils"

	"github.com/werf/werf/integration/pkg/suite_init"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	utilsDocker "github.com/werf/werf/integration/pkg/utils/docker"
)

var suiteImage1 = "flant/werf-test:base-image-suite-image1"
var suiteImage2 = "flant/werf-test:base-image-suite-image2"

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Ansible suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"docker"},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData = struct {
	suite_init.SuiteData

	Registry                  string
	RegistryContainerName     string
	RegistryProjectRepository string
}{}

var _ = AfterEach(func() {
	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		SuiteData.WerfBinPath,
		"host", "purge", "--force",
	)
})

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
var _ = SuiteData.SetupTmp(suite_init.NewTmpDirData())

var _ = SuiteData.AppendSynchronizedBeforeSuiteNode1Func(func() {
	for _, suiteImage := range []string{suiteImage1, suiteImage2} {
		if !utilsDocker.IsImageExist(suiteImage) {
			Ω(utilsDocker.Pull(suiteImage)).Should(Succeed(), "docker pull")
		}
	}
})

var _ = SuiteData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(_ []byte) {
	SuiteData.Registry, SuiteData.RegistryContainerName = utilsDocker.LocalDockerRegistryRun()
})

var _ = SuiteData.AppendSynchronizedAfterSuiteAllNodesFunc(func() {
	utilsDocker.ContainerStopAndRemove(SuiteData.RegistryContainerName)
})

var _ = BeforeEach(func() {
	SuiteData.RegistryProjectRepository = strings.Join([]string{SuiteData.Registry, SuiteData.ProjectName}, "/")
})
