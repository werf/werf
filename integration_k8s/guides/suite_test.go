package guides_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/prashantv/gostub"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/werf/werf/pkg/testing/utils"
)

func TestIntegration(t *testing.T) {
	if !utils.MeetsRequirements(requiredSuiteTools, requiredSuiteEnvs) {
		fmt.Println("Missing required tools")
		os.Exit(1)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Guides Suite")
}

var requiredSuiteTools = []string{"git", "docker"}
var requiredSuiteEnvs []string

var tmpDir string
var testDirPath string
var werfBinPath string
var stubs = gostub.New()

var _ = SynchronizedBeforeSuite(func() []byte {
	computedPathToWerf := utils.ProcessWerfBinPath()
	return []byte(computedPathToWerf)
}, func(computedPathToWerf []byte) {
	werfBinPath = string(computedPathToWerf)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	tmpDir = utils.GetTempDir()
	testDirPath = tmpDir

	utils.BeforeEachOverrideWerfProjectName(stubs)

	imagesRepoAddress := fmt.Sprintf("%s/%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), utils.ProjectName())

	stubs.SetEnv("WERF_STAGES_STORAGE", ":local")
	stubs.SetEnv("WERF_IMAGES_REPO", imagesRepoAddress)
})

var _ = AfterEach(func() {
	err := os.RemoveAll(tmpDir)
	Ω(err).ShouldNot(HaveOccurred())

	stubs.Reset()
})
