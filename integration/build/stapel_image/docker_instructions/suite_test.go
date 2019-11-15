// +build integration integration_k8s

package docker_instruction

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/flant/werf/integration/utils"
)

func TestIntegration(t *testing.T) {
	if !utils.MeetsRequirements(requiredSuiteTools, requiredSuiteEnvs) {
		fmt.Println("Missing required tools")
		os.Exit(1)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Build/Stapel Image/Docker Instructions Suite")
}

var requiredSuiteTools = []string{"docker"}
var requiredSuiteEnvs []string

var testDirPath string
var werfBinPath string

var _ = SynchronizedBeforeSuite(func() []byte {
	computedPathToWerf := utils.ProcessWerfBinPath()
	return []byte(computedPathToWerf)
}, func(computedPathToWerf []byte) {
	werfBinPath = string(computedPathToWerf)
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	utils.BeforeEachOverrideWerfProjectName()
	Ω(os.Setenv("WERF_STAGES_STORAGE", ":local")).Should(Succeed())
})

var _ = AfterEach(func() {
	utils.RunSucceedCommand(
		testDirPath,
		werfBinPath,
		"stages", "purge", "-s", ":local", "--force",
	)

	utils.ResetEnviron()
})

func fixturePath(paths ...string) string {
	pathsToJoin := append([]string{"_fixtures"}, paths...)
	return filepath.Join(pathsToJoin...)
}
