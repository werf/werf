package e2e_converge_test

import (
	"os"
	"strings"

	"github.com/werf/werf/pkg/util"
)

func setupEnv() {
	SuiteData.Stubs.SetEnv("WERF_REPO", strings.Join([]string{os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), SuiteData.ProjectName}, "/"))
	SuiteData.Stubs.SetEnv("WERF_ENV", "test")

	if util.GetBoolEnvironmentDefaultFalse("WERF_TEST_K8S_DOCKER_REGISTRY_INSECURE") {
		SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
		SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_INSECURE_REGISTRY")
		SuiteData.Stubs.UnsetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY")
	}
}
