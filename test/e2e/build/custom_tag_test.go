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
	Platforms  []string
	CustomTags []string
}

var _ = Describe("Custom tag build", Label("e2e", "build", "simple"), func() {
	DescribeTable("should succeed and produce expected image",
		func(ctx SpecContext, opts customTagTestOptions) {
			By("initializing")
			setupEnv(opts.setupEnvOptions)

			By("state0: starting")
			repoDirname := "repo0"
			fixtureRelPath := "custom_tag/state0"

			By("state0: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("state0: building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			platforms := lo.Map(opts.Platforms, func(p string, _ int) string {
				return fmt.Sprintf("--platform=%s", p)
			})

			customTags := lo.Map(opts.CustomTags, func(t string, _ int) string {
				return fmt.Sprintf("--add-custom-tag=%s", t)
			})

			buildOut := werfProject.Build(ctx, &werf.BuildOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: slices.Concat(platforms, customTags, []string{"dockerfile"}),
				},
			})
			Expect(buildOut).To(ContainSubstring("Building stage base-stapel/from"))
			Expect(buildOut).To(ContainSubstring("Building stage base-stapel/install"))
			Expect(buildOut).To(ContainSubstring("Building stage base-dockerfile/dockerfile"))
			Expect(buildOut).To(ContainSubstring("Building stage dockerfile/dockerfile"))

			Expect(strings.Count(buildOut, "Adding custom tags")).To(Equal(len(opts.CustomTags) * 2))

			for _, tag := range opts.CustomTags {
				tagRef := strings.Join([]string{os.Getenv("WERF_REPO"), tag}, ":")
				Expect(buildOut).To(ContainSubstring(tagRef))
			}
		},
		Entry(
			"with repo using Vanilla Docker",
			customTagTestOptions{
				setupEnvOptions: setupEnvOptions{
					ContainerBackendMode:        "vanilla-docker",
					WithLocalRepo:               true,
					WithStagedDockerfileBuilder: false,
				},
				Platforms: []string{
					"linux/amd64",
				},
				CustomTags: []string{
					"my-tag",
				},
			},
		),
		// TODO: add multi-platform test
	)
})
