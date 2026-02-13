package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type networkTestOptions struct {
	setupEnvOptions
	ExpectError bool
	FixturePath string
}

var _ = Describe("Network isolation build", Label("e2e", "build", "network"), func() {
	DescribeTable("should handle network isolation correctly",
		func(ctx SpecContext, testOpts networkTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			_, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			repoDirname := "repo0"
			fixtureRelPath := testOpts.FixturePath

			By("preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			By("building images with --network=none")
			opts := &werf.BuildOptions{
				CommonOptions: werf.CommonOptions{
					ShouldFail: testOpts.ExpectError,
					ExtraArgs:  []string{"--network", "none"},
				},
			}
			buildOut := werfProject.Build(ctx, opts)

			if !testOpts.ExpectError {
				Expect(buildOut).To(ContainSubstring("Building stage"))
			}
		},
		Entry("Scenario 1 (Vanilla): Failure when network access is required in RUN/shell with --network=none", Label("scenario1"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode: "vanilla-docker",
				WithLocalRepo:        false,
			},
			ExpectError: true,
			FixturePath: "network/state_no_network_error",
		}),
		Entry("Scenario 1 (BuildKit): Failure when network access is required in RUN/shell with --network=none", Label("scenario1", "buildkit"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode: "buildkit-docker",
				WithLocalRepo:        true,
			},
			ExpectError: true,
			FixturePath: "network/state_no_network_error",
		}),
		Entry("Scenario 2 (Vanilla): Success when NO network access is required in RUN/shell with --network=none", Label("scenario2"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode: "vanilla-docker",
				WithLocalRepo:        false,
			},
			ExpectError: false,
			FixturePath: "network/state_no_network_success",
		}),
		Entry("Scenario 2 (BuildKit): Success when NO network access is required in RUN/shell with --network=none", Label("scenario2", "buildkit"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{
				ContainerBackendMode: "buildkit-docker",
				WithLocalRepo:        true,
			},
			ExpectError: false,
			FixturePath: "network/state_no_network_success",
		}),
	)
})
