package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func GetTempDir() (string, error) {
	dir, err := ioutil.TempDir("", "werf-integration-tests-")
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "darwin" {
		dir, err = filepath.EvalSymlinks(dir)
		if err != nil {
			return "", fmt.Errorf("eval symlinks of path %s failed: %s", dir, err)
		}
	}

	return dir, nil
}

var werfBinPath string

func ProcessWerfBinPath() string {
	werfBinPath = os.Getenv("WERF_TEST_BINARY_PATH")
	if werfBinPath == "" {
		var err error
		werfBinPath, err = gexec.Build("github.com/flant/werf/cmd/werf")
		Ω(err).ShouldNot(HaveOccurred())
	}

	return werfBinPath
}

func WerfBinArgs(userArgs ...string) []string {
	var args []string
	if os.Getenv("WERF_TEST_BINARY_PATH") != "" && os.Getenv("WERF_TEST_COVERAGE_DIR") != "" {
		coverageFilePath := filepath.Join(
			os.Getenv("WERF_TEST_COVERAGE_DIR"),
			fmt.Sprintf("%s-%s.out", strconv.FormatInt(time.Now().UTC().UnixNano(), 10), GetRandomString(10)),
		)
		args = append(args, fmt.Sprintf("-test.coverprofile=%s", coverageFilePath))
	}

	args = append(args, userArgs...)

	return args
}

func BeforeEachOverrideWerfProjectName() {
	projectName := "werf-integration-test-" + strconv.Itoa(os.Getpid()) + "-" + GetRandomString(10)
	Ω(os.Setenv("WERF_PROJECT_NAME", projectName)).ShouldNot(HaveOccurred())
}

func ProjectName() string {
	val := os.Getenv("WERF_PROJECT_NAME")
	Expect(val).NotTo(BeEmpty())
	return val
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

var environ = os.Environ()

func ResetEnviron() {
	os.Clearenv()
	for _, env := range environ {
		// ignore dynamic variables (e.g. "=ExitCode" windows variable)
		if strings.HasPrefix(env, "=") {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		envName := parts[0]
		envValue := parts[1]

		Ω(os.Setenv(envName, envValue)).Should(Succeed(), env)
	}
}
