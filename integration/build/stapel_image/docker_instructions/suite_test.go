package docker_instruction_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/prashantv/gostub"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/flant/werf/pkg/testing/utils"
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
var stubs = gostub.New()

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
	utils.BeforeEachOverrideWerfProjectName(stubs)
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
