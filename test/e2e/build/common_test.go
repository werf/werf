package e2e_build_test

import (
	"os"

	"github.com/werf/werf/v2/test/pkg/suite_init"
)

type setupEnvOptions struct {
	WithFinalRepo               bool
	WithStagedDockerfileBuilder bool
	State                       string
}

func setupEnv(opts setupEnvOptions) {
	if host := os.Getenv("WERF_TEST_BUILDKIT_HOST"); host != "" {
		SuiteData.Stubs.SetEnv("WERF_BUILDKIT_HOST", host)
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_BUILDKIT_HOST")
	}

	SuiteData.Stubs.SetEnv(
		"WERF_REPO",
		suite_init.TestRepo(SuiteData.ProjectName),
	)

	if opts.WithFinalRepo {
		SuiteData.Stubs.SetEnv(
			"WERF_FINAL_REPO",
			suite_init.TestRepo(SuiteData.ProjectName+"-final"),
		)
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_FINAL_REPO")
	}

	SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
	SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")

	if opts.WithStagedDockerfileBuilder {
		SuiteData.Stubs.SetEnv("WERF_FORCE_STAGED_DOCKERFILE", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_FORCE_STAGED_DOCKERFILE")
	}

	SuiteData.Stubs.SetEnv("ENV_SECRET", "WERF_BUILD_SECRET")
}
