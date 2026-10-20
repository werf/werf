package e2e_build_test

import (
	"errors"
	"os/exec"
	"slices"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/suite_init"
	"github.com/werf/werf/v3/test/pkg/utils"
	"github.com/werf/werf/v3/test/pkg/werf"
)

type checkBuiltImagesTestOptions struct {
	setupEnvOptions
	Args    []string
	Env     map[string]string
	Require bool
}

var checkBuiltImagesEnv = setupEnvOptions{
	ContainerBackendMode: "docker",
	WithLocalRepo:        true,
}

var _ = Describe("Check built images", Label("e2e", "build", "simple"), func() {
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
			checkOutBytes, checkErr := utils.RunCommandWithOptions(ctx, werfProject.GitRepoPath, werfProject.WerfBinPath, slices.Concat([]string{"build"}, opts.Args), utils.RunCommandOptions{})
			var checkExitErr *exec.ExitError
			Expect(errors.As(checkErr, &checkExitErr)).To(BeTrue(), "expected werf build to exit with a non-zero status, got: %v", checkErr)
			checkOut := string(checkOutBytes)
			if opts.Require {
				Expect(checkOut).To(ContainSubstring("build images first without --require-built-images"))
			} else {
				Expect(checkOut).To(ContainSubstring("stages required"))
			}
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
			Require:         true,
		}),
		backendEntry("with --require-built-images for Stapel only", checkBuiltImagesTestOptions{
			setupEnvOptions: checkBuiltImagesEnv,
			Args:            []string{"stapel-shell", "--require-built-images"},
			Require:         true,
		}),
		backendEntry("with --require-built-images for Stapel only using Buildah", checkBuiltImagesTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "native-rootless", WithLocalRepo: true},
			Args:            []string{"stapel-shell", "--require-built-images"},
			Require:         true,
		}),
		backendEntry("with -Z", checkBuiltImagesTestOptions{
			setupEnvOptions: checkBuiltImagesEnv,
			Args:            []string{"-Z"},
			Require:         true,
		}),
		backendEntry("with WERF_CHECK_BUILT_IMAGES", checkBuiltImagesTestOptions{
			setupEnvOptions: checkBuiltImagesEnv,
			Env:             map[string]string{"WERF_CHECK_BUILT_IMAGES": "1"},
		}),
		backendEntry("with WERF_REQUIRE_BUILT_IMAGES", checkBuiltImagesTestOptions{
			setupEnvOptions: checkBuiltImagesEnv,
			Env:             map[string]string{"WERF_REQUIRE_BUILT_IMAGES": "1"},
			Require:         true,
		}),
	)
})

var _ = Describe("Require built images publication", Label("e2e", "build", "simple"), func() {
	DescribeTable("copies secondary stages and publishes all final repositories without building", func(ctx SpecContext, opts setupEnvOptions) {
		setupEnv(opts)
		SuiteData.InitTestRepo(ctx, "repo0", "simple/state0")
		project := newWerfProject("repo0")
		project.Build(ctx, &werf.BuildOptions{CommonOptions: werf.CommonOptions{ExtraArgs: []string{"dockerfile"}}})

		source := suite_init.TestRepo(SuiteData.ProjectName)
		primary := suite_init.TestRepo(SuiteData.ProjectName + "-required")
		finals := []string{suite_init.TestRepo(SuiteData.ProjectName + "-final-one"), suite_init.TestRepo(SuiteData.ProjectName + "-final-two")}
		args := []string{"dockerfile", "--repo", primary, "--secondary-repo", source, "--add-custom-tag", "release"}
		for _, repo := range finals {
			args = append(args, "--final-repo", repo)
		}

		checkOutput := project.Build(ctx, &werf.BuildOptions{CommonOptions: werf.CommonOptions{
			ShouldFail: true,
			ExtraArgs:  slices.Concat(args, []string{"--check-built-images"}),
		}})
		Expect(checkOutput).To(ContainSubstring("stages required"))

		output := project.Build(ctx, &werf.BuildOptions{CommonOptions: werf.CommonOptions{
			ExtraArgs: slices.Concat(args, []string{"--require-built-images"}),
		}})
		Expect(output).NotTo(ContainSubstring("Building stage"))
		var digest string
		for _, repo := range finals {
			tag, err := name.NewTag(repo+":release", name.Insecure)
			Expect(err).NotTo(HaveOccurred())
			manifest, err := remote.Get(tag, remote.WithContext(ctx))
			Expect(err).NotTo(HaveOccurred())
			if digest == "" {
				digest = manifest.Digest.String()
			}
			Expect(manifest.Digest.String()).To(Equal(digest))
		}
		project.Build(ctx, &werf.BuildOptions{CommonOptions: werf.CommonOptions{
			ExtraArgs: slices.Concat(args, []string{"--check-built-images"}),
		}})
	},
		backendEntry("Docker", setupEnvOptions{ContainerBackendMode: "docker", WithLocalRepo: true}),
		backendEntry("Buildah", setupEnvOptions{ContainerBackendMode: "native-rootless", WithLocalRepo: true}),
	)
})
