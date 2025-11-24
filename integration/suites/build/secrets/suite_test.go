package common_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Build/Secrets suite", suite_init.TestSuiteEntrypointFuncOptions{})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData suite_init.SuiteData

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())

	_ = SuiteData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(_ context.Context, _ []byte) {
		_ = utils.CreateTmpFileInHome("secret_file_in_home", "secret")
	})

	_ = AfterEach(func(ctx SpecContext) {
		utils.RunSucceedCommand(ctx, "", SuiteData.WerfBinPath, "host", "purge", "--force", "--project-name", SuiteData.ProjectName)
	})
)

func setupEnv(opts testOptions) {
	if opts.ContainerBackendMode == "docker" || strings.HasSuffix(opts.ContainerBackendMode, "-docker") {
		SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", "docker")
	} else {
		SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", opts.ContainerBackendMode)
	}

	if opts.ContainerBackendMode == "buildkit-docker" {
		SuiteData.Stubs.SetEnv("DOCKER_BUILDKIT", "1")
	} else {
		SuiteData.Stubs.SetEnv("DOCKER_BUILDKIT", "0")
	}

	if opts.WithStagedDockerfileBuilder {
		SuiteData.Stubs.SetEnv("WERF_FORCE_STAGED_DOCKERFILE", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_FORCE_STAGED_DOCKERFILE")
	}

	SuiteData.Stubs.SetEnv("ENV_SECRET", "WERF_BUILD_SECRET")
}
