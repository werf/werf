package e2e_build_test

import (
	"strings"
)

type setupEnvOptions struct {
	ContainerBackendMode        string
	WithLocalRepo               bool
	WithStagedDockerfileBuilder bool
}

func setupEnv(opts setupEnvOptions) {
	if opts.ContainerBackendMode == "docker" || strings.HasSuffix(opts.ContainerBackendMode, "-docker") {
		SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", "docker")
	} else {
		SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", opts.ContainerBackendMode)
	}

	if opts.WithLocalRepo && (opts.ContainerBackendMode == "docker" || strings.HasSuffix(opts.ContainerBackendMode, "-docker")) {
		SuiteData.WerfRepo = strings.Join([]string{SuiteData.RegistryLocalAddress, SuiteData.ProjectName}, "/")
		SuiteData.Stubs.SetEnv("WERF_REPO", SuiteData.WerfRepo)
	} else if opts.WithLocalRepo {
		SuiteData.WerfRepo = strings.Join([]string{SuiteData.RegistryInternalAddress, SuiteData.ProjectName}, "/")
		SuiteData.Stubs.SetEnv("WERF_REPO", SuiteData.WerfRepo)
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_REPO")
	}

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
}
