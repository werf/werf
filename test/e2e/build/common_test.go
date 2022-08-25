package e2e_build_test

import (
	"strings"
)

func setupEnv(withLocalRepo bool, buildahMode string) {
	SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", buildahMode)

	if withLocalRepo && buildahMode == "docker" {
		SuiteData.Stubs.SetEnv("WERF_REPO", strings.Join([]string{SuiteData.RegistryLocalAddress, SuiteData.ProjectName}, "/"))
	} else if withLocalRepo {
		SuiteData.Stubs.SetEnv("WERF_REPO", strings.Join([]string{SuiteData.RegistryInternalAddress, SuiteData.ProjectName}, "/"))
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_REPO")
	}

	if withLocalRepo {
		SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
		SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_INSECURE_REGISTRY")
		SuiteData.Stubs.UnsetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY")
	}
}
