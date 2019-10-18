// +build integration integration_k8s

package guides_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/flant/werf/integration/utils"
)

func TestIntegration(t *testing.T) {
	if !meetsRequirements() {
		fmt.Println("Missing required tools")
		os.Exit(1)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Guides Suite")
}

var requiredSuiteTools = []string{"git", "docker"}
var requiredSuiteEnvs []string

func meetsRequirements() bool {
	hasRequirements := true
	for _, tool := range requiredSuiteTools {
		_, err := exec.LookPath(tool)
		if err != nil {
			fmt.Printf("You must have %s installed on your PATH\n", tool)
			hasRequirements = false
		}
	}

	for _, env := range requiredSuiteEnvs {
		_, exist := os.LookupEnv(env)
		if !exist {
			fmt.Printf("You must have defined %s environment variable\n", env)
			hasRequirements = false
		}
	}

	return hasRequirements
}

var tmpDir string
var werfBinPath string

var _ = SynchronizedBeforeSuite(func() []byte {
	pathToWerf := os.Getenv("WERF_TEST_WERF_BINARY_PATH")
	if pathToWerf == "" {
		var err error
		pathToWerf, err = gexec.Build("github.com/flant/werf/cmd/werf")
		Ω(err).ShouldNot(HaveOccurred())
	}

	return []byte(pathToWerf)
}, func(computedPathToWerf []byte) {
	werfBinPath = string(computedPathToWerf)
})

var _ = BeforeEach(func() {
	var err error
	tmpDir, err = ioutil.TempDir("", "werf-integration-tests")
	Ω(err).ShouldNot(HaveOccurred())
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
