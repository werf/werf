package base_image_test

import (
	"strings"
	"testing"

	"github.com/werf/werf/integration/utils"

	"github.com/werf/werf/integration/suite_init"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	utilsDocker "github.com/werf/werf/integration/utils/docker"
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
		"purge", "--force",
	)
})

var _ = SuiteData.StubsData.Setup()
var _ = SuiteData.SynchronizedSuiteCallbacksData.Setup()
var _ = SuiteData.WerfBinaryData.Setup(&SuiteData.SynchronizedSuiteCallbacksData)
var _ = SuiteData.ProjectNameData.Setup(&SuiteData.StubsData)
var _ = SuiteData.TmpDirData.Setup()

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
