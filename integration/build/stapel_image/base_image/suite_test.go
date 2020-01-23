package base_image_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/prashantv/gostub"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/flant/werf/pkg/testing/utils"
	utilsDocker "github.com/flant/werf/pkg/testing/utils/docker"
)

func TestIntegration(t *testing.T) {
	if !utils.MeetsRequirements(requiredSuiteTools, requiredSuiteEnvs) {
		fmt.Println("Missing required tools")
		os.Exit(1)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Build/Stapel Image/Base Image Suite")
}

var requiredSuiteTools = []string{"docker"}
var requiredSuiteEnvs []string

var testDirPath string
var werfBinPath string
var stubs = gostub.New()
var registry, registryContainerName string
var registryProjectRepository string

var suiteImage1 = "hello-world"
var suiteImage2 = "alpine"

var _ = SynchronizedBeforeSuite(func() []byte {
	for _, suiteImage := range []string{suiteImage1, suiteImage2} {
		if !utilsDocker.IsImageExist(suiteImage) {
			Ω(utilsDocker.Pull(suiteImage)).Should(Succeed(), "docker pull")
		}
	}

	computedPathToWerf := utils.ProcessWerfBinPath()
	return []byte(computedPathToWerf)
}, func(computedPathToWerf []byte) {
	werfBinPath = string(computedPathToWerf)
	registry, registryContainerName = utilsDocker.LocalDockerRegistryRun()
})

var _ = SynchronizedAfterSuite(func() {
	utilsDocker.ContainerStopAndRemove(registryContainerName)
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	utils.BeforeEachOverrideWerfProjectName(stubs)
	registryProjectRepository = strings.Join([]string{registry, utils.ProjectName()}, "/")

	stubs.SetEnv("WERF_STAGES_STORAGE", ":local")
})

var _ = AfterEach(func() {
	utils.RunSucceedCommand(
		testDirPath,
		werfBinPath,
		"stages", "purge", "-s", ":local", "--force",
	)

	stubs.Reset()
})
