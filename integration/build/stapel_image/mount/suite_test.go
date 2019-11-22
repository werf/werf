// +build integration integration_k8s

package mount_test

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
	RunSpecs(t, "Integration Build/Stapel Image/Mount Suite")
}

var requiredSuiteTools = []string{"docker"}
var requiredSuiteEnvs []string

var tmpDir string
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
	var err error
	tmpDir, err = utils.GetTempDir()
	Ω(err).ShouldNot(HaveOccurred())

	testDirPath = tmpPath()

	utils.BeforeEachOverrideWerfProjectName()
	Ω(os.Setenv("WERF_STAGES_STORAGE", ":local")).Should(Succeed())
})

var _ = AfterEach(func() {
	utils.RunSucceedCommand(
		testDirPath,
		werfBinPath,
		"stages", "purge", "-s", ":local", "--force",
	)

	err := os.RemoveAll(tmpDir)
	Ω(err).ShouldNot(HaveOccurred())

	utils.ResetEnviron()
})

func tmpPath(paths ...string) string {
	pathsToJoin := append([]string{tmpDir}, paths...)
	return filepath.Join(pathsToJoin...)
}

func fixturePath(paths ...string) string {
	absFixturesPath, err := filepath.Abs("_fixtures")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "getting absolute fixtures path failed: %s", err)
		os.Exit(1)
	}
	pathsToJoin := append([]string{absFixturesPath}, paths...)
	return filepath.Join(pathsToJoin...)
}
