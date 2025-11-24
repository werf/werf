package e2e_stages_copy_test

import (
	"strconv"
	"strings"
)

const (
	artifactCacheVersion = "1"
	artifactData         = "1"
	archiveAddr          = "archive:copy-test-archive.tar.gz"
)

type commonTestOptions struct {
	All *bool
}

func setupEnv() {
	SuiteData.Stubs.SetEnv("WERF_SYNCHRONIZATION", ":local")
	SuiteData.Stubs.SetEnv("WERF_ENV", "test")
	SuiteData.Stubs.SetEnv("ARTIFACT_CACHE_VERSION", artifactCacheVersion)
	SuiteData.Stubs.SetEnv("ARTIFACT_DATA", artifactData)

	SuiteData.WerfFromAddr = strings.Join([]string{SuiteData.FromRegistryLocalAddress, SuiteData.ProjectName}, "/")
	SuiteData.WerfToAddr = strings.Join([]string{SuiteData.ToRegistryLocalAddress, SuiteData.ProjectName}, "/")

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

	return stagesCopyArgs
}
