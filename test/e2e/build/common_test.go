package e2e_build_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v2/test/pkg/suite_init"
)

type setupEnvOptions struct {
	ContainerBackendMode        string
	WithLocalRepo               bool
	WithFinalRepo               bool
	WithStagedDockerfileBuilder bool
	State                       string
}

func buildkitHostOrSkip() string {
	buildkitHost := os.Getenv("WERF_TEST_BUILDKIT_HOST")
	if buildkitHost == "" {
		Skip("WERF_TEST_BUILDKIT_HOST is not set")
	}
	return buildkitHost
}

func setupEnv(opts setupEnvOptions) {
	if opts.ContainerBackendMode == "buildkit" {
		SuiteData.Stubs.SetEnv("WERF_BUILDKIT_HOST", buildkitHostOrSkip())
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_BUILDKIT_HOST")
	}

	if opts.WithLocalRepo {
		SuiteData.Stubs.SetEnv(
			"WERF_REPO",
			suite_init.TestRepo(SuiteData.ProjectName),
		)
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_REPO")
	}

	if opts.WithFinalRepo {
		SuiteData.Stubs.SetEnv(
			"WERF_FINAL_REPO",
			suite_init.TestRepo(SuiteData.ProjectName+"-final"),
		)
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_FINAL_REPO")
	}

	if opts.WithLocalRepo || opts.WithFinalRepo {
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
