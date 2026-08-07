package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type networkTestOptions struct {
	setupEnvOptions
	ExpectError        bool
	FixturePath        string
	NetworkNone        bool
	ExpectNetworkValue string
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
				if testOpts.ExpectNetworkValue != "" {
					Expect(buildOut).To(ContainSubstring("network: " + testOpts.ExpectNetworkValue))
				} else {
					Expect(buildOut).To(ContainSubstring("network: default"))
				}
			}
		},

		// CLI tests (verify CLI works when YAML is empty)
		Entry("Stapel (Docker): Failure with --backend-network=none", Label("stapel"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/stapel",
			NetworkNone:     true,
		}),
		Entry("Stapel (Docker): Success without --backend-network flag", Label("stapel"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     false,
			FixturePath:     "network/stapel",
			NetworkNone:     false,
		}),
		Entry("Dockerfile (Docker): Failure with --backend-network=none", Label("dockerfile"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/dockerfile",
			NetworkNone:     true,
		}),
		Entry("Dockerfile (Docker): Success without --backend-network flag", Label("dockerfile"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     false,
			FixturePath:     "network/dockerfile",
			NetworkNone:     false,
		}),

		// YAML tests (verify network directive in werf.yaml)
		Entry("Stapel (Docker): Failure with network:none in werf.yaml", Label("stapel", "yml"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/stapel_yml",
			NetworkNone:     false,
		}),
		Entry("Stapel (Docker): Success with network:host in werf.yaml", Label("stapel", "yml"), networkTestOptions{
			setupEnvOptions:    setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:        false,
			FixturePath:        "network/stapel_yml_success",
			NetworkNone:        false,
			ExpectNetworkValue: "host",
		}),
		Entry("Dockerfile (Docker): Failure with network:none in werf.yaml", Label("dockerfile", "yml"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/dockerfile_yml",
			NetworkNone:     false,
		}),
		Entry("Dockerfile (Docker): Success with network:host in werf.yaml", Label("dockerfile", "yml"), networkTestOptions{
			setupEnvOptions:    setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:        false,
			FixturePath:        "network/dockerfile_yml_success",
			NetworkNone:        false,
			ExpectNetworkValue: "host",
		}),

		// CLI overriding YAML
		Entry("Stapel (Docker): CLI --backend-network=none overrides YAML network:host (should fail)", Label("stapel", "override"), networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/stapel_yml_success",
			NetworkNone:     true, // CLI 'none' overrides YAML 'host'
		}),
	)
})
