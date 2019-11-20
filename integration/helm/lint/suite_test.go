// +build integration integration_k8s

package lint_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/flant/werf/integration/utils"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Helm/Lint Suite")
}

var testDirPath string
var tmpDir string
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
