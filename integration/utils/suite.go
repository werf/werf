package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func GetTempDir() (string, error) {
	return ioutil.TempDir("", "werf-integration-tests-")
}

func ProcessWerfBinPath() string {
	path := os.Getenv("WERF_TEST_WERF_BINARY_PATH")
	if path == "" {
		var err error
		path, err = gexec.Build("github.com/flant/werf/cmd/werf")
		Ω(err).ShouldNot(HaveOccurred())
	}

	return path
}

func BeforeEachOverrideWerfProjectName() {
	projectName := "werf-integration-test-" + GetRandomString(10)
	Ω(os.Setenv("WERF_PROJECT_NAME", projectName)).ShouldNot(HaveOccurred())
}

func MeetsRequirements(requiredSuiteTools, requiredSuiteEnvs []string) bool {
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
