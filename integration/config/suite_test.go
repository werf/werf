// +build integration integration_k8s

package config_test

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/flant/werf/integration/utils"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Config Suite")
}

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
})

var _ = AfterEach(func() {
	utils.ResetEnviron()
})

func fixturePath(paths ...string) string {
	pathsToJoin := append([]string{"_fixtures"}, paths...)
	return filepath.Join(pathsToJoin...)
}
