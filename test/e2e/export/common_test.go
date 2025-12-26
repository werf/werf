package e2e_export_test

import (
	"fmt"
	"os"
)

type commonTestOptions struct {
	Platforms    []string
	CustomLabels []string
}

func setupEnv() {
	SuiteData.Stubs.SetEnv("WERF_REPO", fmt.Sprintf("%s/%s",
		os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"),
		SuiteData.ProjectName,
	))
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
