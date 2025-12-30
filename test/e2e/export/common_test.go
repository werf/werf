package e2e_export_test

import (
	"github.com/werf/werf/v2/test/pkg/suite_init"
)

type commonTestOptions struct {
	Platforms    []string
	CustomLabels []string
}

func setupEnv() {
	SuiteData.Stubs.SetEnv("WERF_REPO", suite_init.TestRepo(SuiteData.ProjectName))
}

func getExportArgs(imageName string, opts commonTestOptions) []string {
	exportArgs := []string{
		"--tag",
		imageName,
	}
	if len(opts.Platforms) > 0 {
		for _, platform := range opts.Platforms {
			exportArgs = append(exportArgs, "--platform", platform)
		}
	}
	if len(opts.CustomLabels) > 0 {
		for _, label := range opts.CustomLabels {
			exportArgs = append(exportArgs, "--add-label", label)
		}
	}

	return exportArgs
}
