package e2e_verify_test

import (
	"strings"

	"github.com/werf/werf/v2/test/pkg/suite_init"
)

type setupEnvOptions struct {
	ContainerBackendMode        string
	WithLocalRepo               bool
	WithStagedDockerfileBuilder bool
	State                       string
}

func setupEnv(opts setupEnvOptions) {
	if opts.ContainerBackendMode == "docker" || strings.HasSuffix(opts.ContainerBackendMode, "-docker") {
		SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", "docker")
	} else {
		SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", opts.ContainerBackendMode)
	}

	SuiteData.Stubs.SetEnv("WERF_REPO", suite_init.TestRepo(SuiteData.ProjectName))

	if opts.ContainerBackendMode == "buildkit-docker" {
		SuiteData.Stubs.SetEnv("DOCKER_BUILDKIT", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("DOCKER_BUILDKIT")
	}

	if opts.WithLocalRepo {
		SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
		SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_INSECURE_REGISTRY")
		SuiteData.Stubs.UnsetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY")
	}

	if opts.WithStagedDockerfileBuilder {
		SuiteData.Stubs.SetEnv("WERF_FORCE_STAGED_DOCKERFILE", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_FORCE_STAGED_DOCKERFILE")
	}

	SuiteData.Stubs.SetEnv("ENV_SECRET", "WERF_BUILD_SECRET")
}
