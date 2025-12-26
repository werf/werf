package base_image_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
	utilsDocker "github.com/werf/werf/v2/test/pkg/utils/docker"
)

var (
	suiteImage1 = "flant/werf-test:base-image-suite-image1"
	suiteImage2 = "flant/werf-test:base-image-suite-image2"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Ansible suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"docker"},
	RequiredSuiteEnvs: []string{
		"WERF_TEST_K8S_DOCKER_REGISTRY",
	},
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

var _ = AfterEach(func(ctx SpecContext) {
	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "host", "purge", "--force")
})

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
)

var _ = SuiteData.AppendSynchronizedBeforeSuiteNode1Func(func(ctx context.Context) {
	for _, suiteImage := range []string{suiteImage1, suiteImage2} {
		if !utilsDocker.IsImageExist(ctx, suiteImage) {
			Expect(utilsDocker.Pull(ctx, suiteImage)).Should(Succeed(), "docker pull")
		}
	}
})

var _ = BeforeEach(func() {
	SuiteData.Stubs.SetEnv("WERF_REPO", fmt.Sprintf("%s/%s",
		os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"),
		SuiteData.ProjectName,
	))
})
