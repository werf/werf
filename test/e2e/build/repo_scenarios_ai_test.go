package e2e_build_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

// Covers the registry-model flags (--repo, --images-repo, --final-repo,
// --cache-from, --cache-to, --meta-repo) individually and in every
// user-facing combination, per the registry-model-rework Definition of Done.
var _ = Describe("Registry model repo flag scenarios", Label("e2e", "build", "registry-model"), func() {
	newRepoAddr := func(suffix string) string {
		return fmt.Sprintf("%s/%s-%s-%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), SuiteData.ProjectName, suffix, utils.GetRandomString(6))
	}

	insecureArgs := []string{"--insecure-registry", "--skip-tls-verify-registry"}

	initProject := func(ctx SpecContext, repoDirName string) *werf.Project {
		setupEnv(setupEnvOptions{})
		SuiteData.InitTestRepo(ctx, repoDirName, "simple/state0")
		return werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))
	}

	It("builds without any repo flags, defaulting stages/images-repo/meta-repo/cache-from to :local", func(ctx SpecContext) {
		werfProject := initProject(ctx, "repo0")

		By("building from scratch")
		buildOut := werfProject.Build(ctx, &werf.BuildOptions{})
		Expect(buildOut).To(ContainSubstring("stages:      :local"))
		Expect(buildOut).To(ContainSubstring("meta-repo:   :local"))
		Expect(buildOut).To(ContainSubstring("images-repo: :local"))
		Expect(buildOut).To(ContainSubstring("cache-from:  :local"))
		Expect(buildOut).To(ContainSubstring("Building stage"))
		Expect(buildOut).NotTo(ContainSubstring("invalid reference format"))

		By("rebuilding reuses local cache and content tags")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{})
		Expect(buildOut).To(ContainSubstring("Use previously built image"))
		Expect(buildOut).NotTo(ContainSubstring("Building stage"))
	})

	It("builds with --repo alone, fanning stages/meta-repo/images-repo out to the preset address", func(ctx SpecContext) {
		werfProject := initProject(ctx, "repo0")
		repoAddr := newRepoAddr("repo")

		buildOut := werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: append([]string{"--repo", repoAddr}, insecureArgs...),
			},
		})
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("stages:      %s", repoAddr)))
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("meta-repo:   %s", repoAddr)))
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("images-repo: %s", repoAddr)))
		Expect(buildOut).NotTo(ContainSubstring("mutually exclusive"))
	})

	It("builds with --images-repo alone (remote), keeping raw stages local", func(ctx SpecContext) {
		werfProject := initProject(ctx, "repo0")
		imagesRepoAddr := newRepoAddr("images")

		By("building from scratch")
		buildOut := werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: append([]string{"--images-repo", imagesRepoAddr}, insecureArgs...),
			},
		})
		Expect(buildOut).To(ContainSubstring("stages:      :local"))
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("images-repo: %s", imagesRepoAddr)))
		Expect(buildOut).NotTo(ContainSubstring("invalid reference format"))
		Expect(buildOut).NotTo(ContainSubstring("UNAUTHORIZED"))

		By("rebuilding reuses the content tag published to images-repo, even though stages stayed local")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: append([]string{"--images-repo", imagesRepoAddr}, insecureArgs...),
			},
		})
		Expect(buildOut).To(ContainSubstring("Use previously built image"))
		Expect(buildOut).NotTo(ContainSubstring("Building stage"))
	})

	It("builds with --final-repo alone as a fully independent flag, without deprecation warnings", func(ctx SpecContext) {
		werfProject := initProject(ctx, "repo0")
		finalRepoAddr := newRepoAddr("final")

		buildOut := werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: append([]string{"--final-repo", finalRepoAddr}, insecureArgs...),
			},
		})
		Expect(buildOut).NotTo(ContainSubstring("DEPRECATED"))
		Expect(buildOut).To(ContainSubstring("stages:      :local"))
		Expect(buildOut).To(ContainSubstring("into the final repo"))
	})

	It("builds with --images-repo and --final-repo combined, publishing content tags to images-repo and copying into final-repo", func(ctx SpecContext) {
		werfProject := initProject(ctx, "repo0")
		imagesRepoAddr := newRepoAddr("images")
		finalRepoAddr := newRepoAddr("final")

		buildOut := werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: append([]string{"--images-repo", imagesRepoAddr, "--final-repo", finalRepoAddr}, insecureArgs...),
			},
		})
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("images-repo: %s", imagesRepoAddr)))
		Expect(buildOut).To(ContainSubstring("into the final repo"))
		Expect(buildOut).NotTo(ContainSubstring("DEPRECATED"))
		Expect(buildOut).NotTo(ContainSubstring("mutually exclusive"))
	})

	It("builds with --repo and an explicit --images-repo override, the explicit value wins without error", func(ctx SpecContext) {
		werfProject := initProject(ctx, "repo0")
		repoAddr := newRepoAddr("repo")
		imagesRepoAddr := newRepoAddr("images")

		buildOut := werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: append([]string{"--repo", repoAddr, "--images-repo", imagesRepoAddr}, insecureArgs...),
			},
		})
		Expect(buildOut).NotTo(ContainSubstring("mutually exclusive"))
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("stages:      %s", repoAddr)))
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("images-repo: %s", imagesRepoAddr)))
	})

	It("rejects --repo combined with an explicit --cache-from", func(ctx SpecContext) {
		werfProject := initProject(ctx, "repo0")
		repoAddr := newRepoAddr("repo")
		cacheFromAddr := newRepoAddr("cache-from")

		buildOut := werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs:  append([]string{"--repo", repoAddr, "--cache-from", cacheFromAddr}, insecureArgs...),
				ShouldFail: true,
			},
		})
		Expect(buildOut).To(ContainSubstring("mutually exclusive"))
	})

	It("rejects --repo combined with an explicit --cache-to", func(ctx SpecContext) {
		werfProject := initProject(ctx, "repo0")
		repoAddr := newRepoAddr("repo")
		cacheToAddr := newRepoAddr("cache-to")

		buildOut := werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs:  append([]string{"--repo", repoAddr, "--cache-to", cacheToAddr}, insecureArgs...),
				ShouldFail: true,
			},
		})
		Expect(buildOut).To(ContainSubstring("mutually exclusive"))
	})

	It("rejects --repo combined with an explicit --meta-repo", func(ctx SpecContext) {
		werfProject := initProject(ctx, "repo0")
		repoAddr := newRepoAddr("repo")
		metaRepoAddr := newRepoAddr("meta")

		buildOut := werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs:  append([]string{"--repo", repoAddr, "--meta-repo", metaRepoAddr}, insecureArgs...),
				ShouldFail: true,
			},
		})
		Expect(buildOut).To(ContainSubstring("mutually exclusive"))
	})

	It("builds with granular flags only (--cache-from, --cache-to, --images-repo, --meta-repo), no --repo", func(ctx SpecContext) {
		werfProject := initProject(ctx, "repo0")
		cacheFromAddr := newRepoAddr("cache-from")
		cacheToAddr := newRepoAddr("cache-to")
		imagesRepoAddr := newRepoAddr("images")
		metaRepoAddr := newRepoAddr("meta")

		buildOut := werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: append([]string{
					"--cache-from", cacheFromAddr,
					"--cache-to", cacheToAddr,
					"--images-repo", imagesRepoAddr,
					"--meta-repo", metaRepoAddr,
				}, insecureArgs...),
			},
		})
		Expect(buildOut).NotTo(ContainSubstring("mutually exclusive"))
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("cache-from:  %s", cacheFromAddr)))
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("cache-to:    %s", cacheToAddr)))
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("images-repo: %s", imagesRepoAddr)))
		Expect(buildOut).To(ContainSubstring(fmt.Sprintf("meta-repo:   %s", metaRepoAddr)))
	})
})
