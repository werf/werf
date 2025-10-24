package e2e_build_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/werf"
)

type importTestOptions struct {
	setupEnvOptions
	stateName  string
	shouldFail bool
}

var _ = DescribeTable(
	"Images import", Label("e2e", "build", "extra"),
	func(ctx SpecContext, testOpts importTestOptions) {
		By("initializing")
		setupEnv(testOpts.setupEnvOptions)

		By("starting")
		repoDirname := "repo0"
		fixtureRelPath := filepath.Join("images_import", testOpts.stateName)

		By("preparing test repo")
		SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

		By("building images")
		werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
		buildOut := werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{ShouldFail: testOpts.shouldFail},
		})
		Expect(buildOut).To(ContainSubstring("Building stage"))
		Expect(buildOut).To(ContainSubstring("nothing to import"))
	},
	Entry(
		"should fail when import does nothing", importTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode:        "vanilla-docker",
				WithLocalRepo:               false,
				WithStagedDockerfileBuilder: false,
			},
			stateName:  "nothing",
			shouldFail: true,
		},
	),
)
