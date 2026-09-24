package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/werf"
)

type networkTestOptions struct {
	setupEnvOptions
	ExpectError        bool
	FixturePath        string
	NetworkNone        bool
	ExpectNetworkValue string
}

func (opts networkTestOptions) env() setupEnvOptions {
	return opts.setupEnvOptions
}

var _ = Describe("Network isolation build", Label("e2e", "build", "network"), func() {
	DescribeTable("should handle network isolation correctly",
		func(ctx SpecContext, testOpts networkTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contback.SkipIfUnavailable(testOpts.ContainerBackendMode)

			repoDirname := "repo0"
			fixtureRelPath := testOpts.FixturePath

			By("preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)
			werfProject := newWerfProject(repoDirname)

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
		backendEntry("Stapel (Docker): Failure with --backend-network=none", networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/stapel",
			NetworkNone:     true,
		}, Label("stapel")),
		backendEntry("Stapel (Docker): Success without --backend-network flag", networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     false,
			FixturePath:     "network/stapel",
			NetworkNone:     false,
		}, Label("stapel")),
		backendEntry("Dockerfile (Docker): Failure with --backend-network=none", networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/dockerfile",
			NetworkNone:     true,
		}, Label("dockerfile")),
		backendEntry("Dockerfile (Docker): Success without --backend-network flag", networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     false,
			FixturePath:     "network/dockerfile",
			NetworkNone:     false,
		}, Label("dockerfile")),

		// YAML tests (verify network directive in werf.yaml)
		backendEntry("Stapel (Docker): Failure with network:none in werf.yaml", networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/stapel_yml",
			NetworkNone:     false,
		}, Label("stapel", "yml")),
		backendEntry("Stapel (Docker): Success with network:host in werf.yaml", networkTestOptions{
			setupEnvOptions:    setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:        false,
			FixturePath:        "network/stapel_yml_success",
			NetworkNone:        false,
			ExpectNetworkValue: "host",
		}, Label("stapel", "yml")),
		backendEntry("Dockerfile (Docker): Failure with network:none in werf.yaml", networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/dockerfile_yml",
			NetworkNone:     false,
		}, Label("dockerfile", "yml")),
		backendEntry("Dockerfile (Docker): Success with network:host in werf.yaml", networkTestOptions{
			setupEnvOptions:    setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:        false,
			FixturePath:        "network/dockerfile_yml_success",
			NetworkNone:        false,
			ExpectNetworkValue: "host",
		}, Label("dockerfile", "yml")),

		// Native Buildah rootless: guards that the network value reaches buildah's build options.
		// native-chroot is absent on purpose — buildah forces host networking for chroot isolation,
		// so `none` can never be honored there.
		backendEntry("Dockerfile (Native Buildah rootless): Failure with --backend-network=none", networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "native-rootless", WithLocalRepo: true},
			ExpectError:     true,
			FixturePath:     "network/dockerfile",
			NetworkNone:     true,
		}, Label("dockerfile")),
		backendEntry("Dockerfile (Native Buildah rootless): Failure with network:none in werf.yaml", networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "native-rootless", WithLocalRepo: true},
			ExpectError:     true,
			FixturePath:     "network/dockerfile_yml",
			NetworkNone:     false,
		}, Label("dockerfile", "yml")),
		backendEntry("Dockerfile (Native Buildah rootless): Success without --backend-network flag", networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "native-rootless", WithLocalRepo: true},
			ExpectError:     false,
			FixturePath:     "network/dockerfile",
			NetworkNone:     false,
		}, Label("dockerfile")),

		// CLI overriding YAML
		backendEntry("Stapel (Docker): CLI --backend-network=none overrides YAML network:host (should fail)", networkTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: false},
			ExpectError:     true,
			FixturePath:     "network/stapel_yml_success",
			NetworkNone:     true, // CLI 'none' overrides YAML 'host'
		}, Label("stapel", "override")),
	)
})
