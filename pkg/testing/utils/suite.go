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

	"github.com/prashantv/gostub"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func GetTempDir() string {
	dir, err := ioutil.TempDir("", "werf-integration-tests-")
	Ω(err).ShouldNot(HaveOccurred())

	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		dir, err = filepath.EvalSymlinks(dir)
		Ω(err).ShouldNot(HaveOccurred(), fmt.Sprintf("eval symlinks of path %s failed: %s", dir, err))
	}

	return dir
}

func ProcessWerfBinPath() string {
	werfBinPath := os.Getenv("WERF_TEST_BINARY_PATH")
	if werfBinPath == "" {
		var err error
		werfBinPath, err = gexec.Build("github.com/werf/werf/cmd/werf")
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

func isWerfTestBinaryPath(path string) bool {
	werfTestBinaryPath := os.Getenv("WERF_TEST_BINARY_PATH")
	return werfTestBinaryPath != "" && werfTestBinaryPath == path
}

func BeforeEachOverrideWerfProjectName(stubs *gostub.Stubs) {
	var packageId string
	filename := filepath.Base(os.Args[0])
	if strings.HasPrefix(filename, "___") { // ide
		packageId = "none"
	} else {
		packageId = strings.Split(filename, ".")[0] // .test .test.exe
	}

	projectName := strings.Join([]string{
		"werf-test",
		packageId,
		strconv.Itoa(os.Getpid()),
		GetRandomString(10),
	}, "-")

	stubs.SetEnv("WERF_PROJECT_NAME", projectName)
}

func ProjectName() string {
	val := os.Getenv("WERF_PROJECT_NAME")
	Ω(val).NotTo(BeEmpty())
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

func FixturePath(paths ...string) string {
	absFixturesPath, err := filepath.Abs("_fixtures")
	if err != nil {
		panic(err)
	}
	pathsToJoin := append([]string{absFixturesPath}, paths...)
	return filepath.Join(pathsToJoin...)
}
