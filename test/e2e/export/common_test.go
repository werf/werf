package e2e_export_test

import (
	"fmt"

	"github.com/werf/werf/v3/test/pkg/suite_init"
	"github.com/werf/werf/v3/test/pkg/utils"
)

type commonTestOptions struct {
	Platforms       []string
	CustomLabels    []string
	BuildReportPath string
}

func setupEnv() {
	SuiteData.Stubs.SetEnv("WERF_REPO", suite_init.TestRepo(SuiteData.ProjectName))
}

// newExportRepo returns a fresh export repository and registers it so that the
// local tag werf creates during export is covered by project cleanup.
func newExportRepo(name string) string {
	repo := suite_init.TestRepo(name)
	SuiteData.CleanupRepositories = append(SuiteData.CleanupRepositories, repo)
	return repo
}

func newRandomExportRepo() string {
	return newExportRepo(fmt.Sprintf("werf-export-%s", utils.GetRandomString(10)))
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
	if opts.BuildReportPath != "" {
		exportArgs = append(exportArgs, "--use-build-report", "--build-report-path", opts.BuildReportPath)
	}

	return exportArgs
}
