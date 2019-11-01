// +build integration integration_k8s

package base_image

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/flant/werf/integration/utils"
	utilsDocker "github.com/flant/werf/integration/utils/docker"
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

var tmpDir string
var werfBinPath string
var registry, registryContainerName string
var registryProjectRepository string

var _ = SynchronizedBeforeSuite(func() []byte {
	computedPathToWerf := utils.ProcessWerfBinPath()
	return []byte(computedPathToWerf)
}, func(computedPathToWerf []byte) {
	werfBinPath = string(computedPathToWerf)
	registry, registryContainerName = utilsDocker.LocalDockerRegistryRun()
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
	utilsDocker.ContainerStopAndRemove(registryContainerName)
})

var _ = BeforeEach(func() {
	var err error
	tmpDir, err = utils.GetTempDir()
	Ω(err).ShouldNot(HaveOccurred())

	utils.BeforeEachOverrideWerfProjectName()

	registryProjectRepository = strings.Join([]string{registry, utils.ProjectName()}, "/")
})

var _ = AfterEach(func() {
	err := os.RemoveAll(tmpDir)
	Ω(err).ShouldNot(HaveOccurred())

	utils.ResetEnviron()
})

func tmpPath(paths ...string) string {
	pathsToJoin := append([]string{tmpDir}, paths...)
	return filepath.Join(pathsToJoin...)
}

func fixturePath(paths ...string) string {
	pathsToJoin := append([]string{"_fixtures"}, paths...)
	return filepath.Join(pathsToJoin...)
}
