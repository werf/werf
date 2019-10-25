// +build integration integration_k8s

package guides_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
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
	RunSpecs(t, "Integration Guides Suite")
}

var requiredSuiteTools = []string{"git", "docker"}
var requiredSuiteEnvs []string

var tmpDir string
var werfBinPath string

var _ = SynchronizedBeforeSuite(func() []byte {
	pathToWerf := utils.ProcessWerfBinPath()
	return []byte(pathToWerf)
}, func(computedPathToWerf []byte) {
	werfBinPath = string(computedPathToWerf)
})

var _ = BeforeEach(func() {
	var err error
	tmpDir, err = utils.GetTempDir()
	Ω(err).ShouldNot(HaveOccurred())

	utils.BeforeEachOverrideWerfProjectName()
})

var _ = AfterEach(func() {
	err := os.RemoveAll(tmpDir)
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

func tmpPath(paths ...string) string {
	pathsToJoin := append([]string{tmpDir}, paths...)
	return filepath.Join(pathsToJoin...)
}

func fixturePath(paths ...string) string {
	pathsToJoin := append([]string{"_fixtures"}, paths...)
	return filepath.Join(pathsToJoin...)
}

func waitTillHostReadyAndCheckResponseBody(url string, maxAttempts int, bodySubstring string) {
	utils.WaitTillHostReadyToRespond(url, maxAttempts)

	resp, err := http.Get(url)
	Ω(err).ShouldNot(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()

	Ω(resp.StatusCode).Should(Equal(200))

	body, err := ioutil.ReadAll(resp.Body)
	Ω(err).ShouldNot(HaveOccurred())
	Ω(string(body)).Should(ContainSubstring(bodySubstring))
}
