package e2e_export_test

import (
	"strings"
)

type commonTestOptions struct {
	Platforms    []string
	CustomLabels []string
}

func setupEnv() {
	SuiteData.WerfRepo = strings.Join([]string{SuiteData.RegistryLocalAddress, SuiteData.ProjectName}, "/")
	SuiteData.Stubs.SetEnv("WERF_REPO", SuiteData.WerfRepo)
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
