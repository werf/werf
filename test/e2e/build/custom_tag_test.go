package e2e_build_test

import (
	"fmt"
	"os"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

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
			fixtureRelPath := "custom_tag/state0"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("state0: building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			customTags := lo.Map(opts.CustomTags, func(t string, _ int) string {
				return fmt.Sprintf("--add-custom-tag=%s", t)
			})

			buildArgs := slices.Concat(customTags, opts.BuildImages)

			buildOut := werfProject.Build(ctx, &werf.BuildOptions{
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
			"with repo, vanilla-docker, image selection and a custom tag",
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
			"with repo, vanilla-docker, no image selection and a custom tag",
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
