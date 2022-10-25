package e2e_build_test

import (
	"strings"
)

type setupEnvOptions struct {
	BuildahMode               string
	WithLocalRepo             bool
	WithForceStagedDockerfile bool
}

func setupEnv(opts setupEnvOptions) {
	SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", opts.BuildahMode)

	if opts.WithLocalRepo && opts.BuildahMode == "docker" {
		SuiteData.Stubs.SetEnv("WERF_REPO", strings.Join([]string{SuiteData.RegistryLocalAddress, SuiteData.ProjectName}, "/"))
	} else if opts.WithLocalRepo {
		SuiteData.Stubs.SetEnv("WERF_REPO", strings.Join([]string{SuiteData.RegistryInternalAddress, SuiteData.ProjectName}, "/"))
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_REPO")
	}

	if opts.WithLocalRepo {
		SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
		SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_INSECURE_REGISTRY")
		SuiteData.Stubs.UnsetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY")
	}

	if opts.WithForceStagedDockerfile {
		SuiteData.Stubs.SetEnv("WERF_FORCE_STAGED_DOCKERFILE", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_FORCE_STAGED_DOCKERFILE")
	}
}
