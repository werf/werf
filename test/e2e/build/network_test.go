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
	NetworkNone bool
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

			var extraArgs []string
			if testOpts.NetworkNone {
				extraArgs = append(extraArgs, "--backend-network", "none")
			}

			By("building images")
			opts := &werf.BuildOptions{
				CommonOptions: werf.CommonOptions{
					ShouldFail: testOpts.ExpectError,
					ExtraArgs:  extraArgs,
				},
			}
			buildOut := werfProject.Build(ctx, opts)

			if !testOpts.ExpectError {
				Expect(buildOut).To(ContainSubstring("Building stage"))
				Expect(buildOut).To(ContainSubstring("network:"))
			}
		},

		// Stapel tests
		Entry("Stapel (Vanilla): Failure with --backend-network=none", Label("stapel"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "vanilla-docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/stapel",
			NetworkNone:     true,
		}),
		Entry("Stapel (Vanilla): Success without --backend-network flag", Label("stapel"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "vanilla-docker", WithLocalRepo: false},
			ExpectError:     false,
			FixturePath:     "network/stapel",
			NetworkNone:     false,
		}),

		// Dockerfile (Vanilla) tests
		Entry("Dockerfile (Vanilla): Failure with --backend-network=none", Label("dockerfile"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "vanilla-docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/dockerfile",
			NetworkNone:     true,
		}),
		Entry("Dockerfile (Vanilla): Success without --backend-network flag", Label("dockerfile"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "vanilla-docker", WithLocalRepo: false},
			ExpectError:     false,
			FixturePath:     "network/dockerfile",
			NetworkNone:     false,
		}),

		// Dockerfile (BuildKit) tests
		Entry("Dockerfile (BuildKit): Failure with --backend-network=none", Label("dockerfile", "buildkit"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "buildkit-docker", WithLocalRepo: true},
			ExpectError:     true,
			FixturePath:     "network/dockerfile",
			NetworkNone:     true,
		}),
		Entry("Dockerfile (BuildKit): Success without --backend-network flag", Label("dockerfile", "buildkit"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "buildkit-docker", WithLocalRepo: true},
			ExpectError:     false,
			FixturePath:     "network/dockerfile",
			NetworkNone:     false,
		}),
	)
})
