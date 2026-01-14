package e2e_build_test

import (
	"fmt"
	"os"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type customTagTestOptions struct {
	setupEnvOptions
	BuildImages        []string
	CustomTags         []string
	ExpectedCustomTags []string
}

var _ = Describe("Custom tag build", Label("e2e", "build", "simple"), func() {
	DescribeTable("should build images with custom tags",
		func(ctx SpecContext, opts customTagTestOptions) {
			By("initializing")
			setupEnv(opts.setupEnvOptions)

			By("state0: starting")

			By("state0: preparing test repo")
			const repoDirname = "repo0"
			const buildReportName = "report-custom-tag.json"
			fixtureRelPath := "custom_tag/state0"

			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("state0: building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)

			customTags := lo.Map(opts.CustomTags, func(t string, _ int) string {
				return fmt.Sprintf("--add-custom-tag=%s", t)
			})

			buildArgs := slices.Concat(customTags, opts.BuildImages)

			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), &werf.WithReportOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: buildArgs,
				},
			})

			Expect(strings.Count(buildOut, "Adding custom tags") / 2).To(Equal(len(opts.ExpectedCustomTags)))

			// validate custom-tag refs
			for _, expectedCustomTag := range opts.ExpectedCustomTags {
				Expect(buildOut).To(ContainSubstring(expectedCustomTag))
			}
		},
		Entry(
			"with repo, vanilla-docker, select multiplatform image, "+
				"and add the custom tag for multiplatform image",
			customTagTestOptions{
				setupEnvOptions: setupEnvOptions{
					ContainerBackendMode:        "vanilla-docker",
					WithLocalRepo:               true,
					WithStagedDockerfileBuilder: false,
				},
				BuildImages: []string{
					"dockerfile",
				},
				CustomTags: []string{
					"my-tag",
				},
				ExpectedCustomTags: []string{
					os.Getenv("WERF_REPO") + ":my-tag",
				},
			},
		),
		Entry(
			"with repo, vanilla-docker, doesn't select any image, "+
				"but add custom tag for multiplatform image and one single platform final image",
			customTagTestOptions{
				setupEnvOptions: setupEnvOptions{
					ContainerBackendMode:        "vanilla-docker",
					WithLocalRepo:               true,
					WithStagedDockerfileBuilder: false,
				},
				BuildImages: []string{},
				CustomTags: []string{
					"%image%-my-tag",
				},
				ExpectedCustomTags: []string{
					os.Getenv("WERF_REPO") + ":dockerfile-my-tag",
					os.Getenv("WERF_REPO") + ":stapel-my-tag",
				},
			},
		),
	)
})
