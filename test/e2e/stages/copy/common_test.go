package e2e_stages_copy_test

import (
	"fmt"
	"strconv"

	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
)

const (
	artifactCacheVersion = "1"
	artifactData         = "1"
	archiveAddr          = "archive:copy-test-archive.tar.gz"
)

type commonTestOptions struct {
	All       *bool
	ExtraArgs []string
}

func setupEnv() {
	SuiteData.Stubs.SetEnv("WERF_ENV", "test")
	SuiteData.Stubs.SetEnv("ARTIFACT_CACHE_VERSION", artifactCacheVersion)
	SuiteData.Stubs.SetEnv("ARTIFACT_DATA", artifactData)

	SuiteData.Stubs.UnsetEnv("WERF_REPO")

	SuiteData.WerfFromAddr = suite_init.TestRepo(fmt.Sprintf("%s-%s", SuiteData.ProjectName, utils.GetRandomString(6)))
	SuiteData.WerfToAddr = suite_init.TestRepo(fmt.Sprintf("%s-%s", SuiteData.ProjectName, utils.GetRandomString(6)))

	SuiteData.WerfArchiveAddr = archiveAddr
}

func getStagesCopyArgs(fromStorage, toStorage string, opts commonTestOptions) []string {
	stagesCopyArgs := []string{
		"--from",
		fromStorage,
		"--to",
		toStorage,
	}

	if opts.All != nil {
		stagesCopyArgs = append(stagesCopyArgs, "--all", strconv.FormatBool(*opts.All))
	}

	stagesCopyArgs = append(stagesCopyArgs, opts.ExtraArgs...)

	return stagesCopyArgs
}
