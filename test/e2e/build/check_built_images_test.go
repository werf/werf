package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/werf"
)

type checkBuiltImagesTestOptions struct {
	setupEnvOptions
	Args []string
	Env  map[string]string
}

var checkBuiltImagesEnv = setupEnvOptions{
	ContainerBackendMode: "docker",
	WithLocalRepo:        true,
}

var _ = Describe("Check built images", Label("e2e", "build", "simple"), entryLabels(checkBuiltImagesEnv), func() {
	DescribeTable("should check built images instead of building them",
		func(ctx SpecContext, opts checkBuiltImagesTestOptions) {
			By("initializing")
			setupEnv(opts.setupEnvOptions)
			setCheckEnv := func() {
				for name, value := range opts.Env {
					SuiteData.Stubs.SetEnv(name, value)
				}
			}
			unsetCheckEnv := func() {
				for name := range opts.Env {
					SuiteData.Stubs.UnsetEnv(name)
				}
			}

			By("state0: preparing test repo")
			repoDirname := "repo0"
			SuiteData.InitTestRepo(ctx, repoDirname, "simple/state0")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			By("state0: checking built images against an empty repo")
			setCheckEnv()
			checkOut := werfProject.Build(ctx, &werf.BuildOptions{
				CommonOptions: werf.CommonOptions{
					ShouldFail: true,
					ExtraArgs:  opts.Args,
				},
			})
			Expect(checkOut).To(ContainSubstring("stages required"))
			Expect(checkOut).NotTo(ContainSubstring("Building stage"))

			By("state0: building images")
			unsetCheckEnv()
			buildOut := werfProject.Build(ctx, nil)
			Expect(buildOut).To(ContainSubstring("Building stage"))
			Expect(buildOut).NotTo(ContainSubstring("Use previously built image"))

			By("state0: checking built images against a populated repo")
			setCheckEnv()
			checkOut = werfProject.Build(ctx, &werf.BuildOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: opts.Args,
				},
			})
			Expect(checkOut).To(ContainSubstring("Use previously built image"))
			Expect(checkOut).NotTo(ContainSubstring("Building stage"))
		},
		backendEntry("with --check-built-images", checkBuiltImagesTestOptions{
			setupEnvOptions: checkBuiltImagesEnv,
			Args:            []string{"--check-built-images"},
		}),
		backendEntry("with --require-built-images", checkBuiltImagesTestOptions{
			setupEnvOptions: checkBuiltImagesEnv,
			Args:            []string{"--require-built-images"},
		}),
		backendEntry("with -Z", checkBuiltImagesTestOptions{
			setupEnvOptions: checkBuiltImagesEnv,
			Args:            []string{"-Z"},
		}),
		backendEntry("with WERF_CHECK_BUILT_IMAGES", checkBuiltImagesTestOptions{
			setupEnvOptions: checkBuiltImagesEnv,
			Env:             map[string]string{"WERF_CHECK_BUILT_IMAGES": "1"},
		}),
	)

	It("should build images regardless of WERF_REQUIRE_BUILT_IMAGES", func(ctx SpecContext) {
		By("initializing")
		setupEnv(checkBuiltImagesEnv)
		SuiteData.Stubs.SetEnv("WERF_REQUIRE_BUILT_IMAGES", "1")

		By("state0: preparing test repo")
		repoDirname := "repo0"
		SuiteData.InitTestRepo(ctx, repoDirname, "simple/state0")
		werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

		By("state0: building images against an empty repo")
		buildOut := werfProject.Build(ctx, nil)
		Expect(buildOut).To(ContainSubstring("Building stage"))
		Expect(buildOut).NotTo(ContainSubstring("stages required"))
	})
})
