package guides_test

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
	RunSpecs(t, "Integration Guides Suite")
}

var requiredSuiteTools = []string{"git", "docker"}
var requiredSuiteEnvs []string

var tmpDir string
var testDirPath string
var werfBinPath string
var stubs = gostub.New()
var registry, registryContainerName string
var registryProjectRepository string

var _ = SynchronizedBeforeSuite(func() []byte {
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
	tmpDir = utils.GetTempDir()
	testDirPath = tmpDir

	utils.BeforeEachOverrideWerfProjectName(stubs)

	registryProjectRepository = strings.Join([]string{registry, utils.ProjectName()}, "/")
})

var _ = AfterEach(func() {
	err := os.RemoveAll(tmpDir)
	Ω(err).ShouldNot(HaveOccurred())

	stubs.Reset()
})
