package e2e_stages_copy_test

import (
	"strconv"
	"strings"
)

type commonTestOptions struct {
	All *bool
}

func setupEnv() {
	SuiteData.Stubs.SetEnv("WERF_SYNCHRONIZATION", ":local")
	SuiteData.Stubs.SetEnv("WERF_ENV", "test")

	SuiteData.WerfFromRepo = strings.Join([]string{SuiteData.FromRegistryLocalAddress, SuiteData.ProjectName}, "/")
	SuiteData.WerfToRepo = strings.Join([]string{SuiteData.ToRegistryLocalAddress, SuiteData.ProjectName}, "/")
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
