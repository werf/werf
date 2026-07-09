package e2e_build_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Content tag reuse", Label("e2e", "build", "content-tag"), func() {
	It("reuses the content tag across local builds, repo and final storages", func(ctx SpecContext) {
		By("initializing")
		setupEnv(setupEnvOptions{})

		repoDirName := "repo0"
		fixtureRelPath := "scratch/state0"

		By("preparing test repo")
		SuiteData.InitTestRepo(ctx, repoDirName, fixtureRelPath)
		werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))

		repoAddr := fmt.Sprintf("%s/%s-%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), SuiteData.ProjectName, utils.GetRandomString(6))
		finalRepoAddr := fmt.Sprintf("%s/%s-%s-final", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), SuiteData.ProjectName, utils.GetRandomString(6))

		By("[1, :local] building all stages from scratch")
		buildOut := werfProject.Build(ctx, &werf.BuildOptions{})
		Expect(buildOut).To(ContainSubstring("Building stage stapel-scratch/"))
		Expect(buildOut).NotTo(ContainSubstring("Reusing image stapel-scratch by content-based tag"))

		By("[2, :local] rebuilding reuses the image by content-based tag")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{})
		Expect(buildOut).To(ContainSubstring("Reusing image stapel-scratch by content-based tag"))
		Expect(buildOut).NotTo(ContainSubstring("Building stage stapel-scratch/"))

		By("[3, repo] building with --repo copies the content-based tag from the :local secondary")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{
					"--repo", repoAddr,
					"--insecure-registry", "--skip-tls-verify-registry",
				},
			},
		})
		Expect(buildOut).To(ContainSubstring("Copy suitable stage from secondary :local"))
		Expect(buildOut).NotTo(ContainSubstring("Building stage stapel-scratch/"))

		By("[3, repo] rebuilding with --repo reuses the image already present in the repo")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{
					"--require-built-images",
					"--repo", repoAddr,
					"--insecure-registry", "--skip-tls-verify-registry",
				},
			},
		})
		Expect(buildOut).To(ContainSubstring("Reusing image stapel-scratch by content-based tag"))
		Expect(buildOut).NotTo(ContainSubstring("Building stage stapel-scratch/"))

		By("[4, final] building with --final-repo does not rebuild any stage")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{
					"--repo", repoAddr,
					"--final-repo", finalRepoAddr,
					"--insecure-registry", "--skip-tls-verify-registry",
				},
			},
		})
		Expect(buildOut).NotTo(ContainSubstring("Building stage stapel-scratch/"))
	})

	It("resolves the content-based tag for a multi-stage image without touching individual stages", func(ctx SpecContext) {
		By("initializing")
		setupEnv(setupEnvOptions{})

		repoDirName := "repo0"
		fixtureRelPath := "content_tag/multistage/state0"

		By("preparing test repo")
		SuiteData.InitTestRepo(ctx, repoDirName, fixtureRelPath)
		werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))

		repoAddr := fmt.Sprintf("%s/%s-%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), SuiteData.ProjectName, utils.GetRandomString(6))

		By("[1, :local] building all stages from scratch")
		buildOut := werfProject.Build(ctx, &werf.BuildOptions{})
		Expect(buildOut).To(ContainSubstring("Building stage app/from"))
		Expect(buildOut).To(ContainSubstring("Building stage app/install"))
		Expect(buildOut).To(ContainSubstring("Building stage app/setup"))
		Expect(buildOut).To(ContainSubstring("Building stage app/gitLatestPatch"))

		By("[2, :local] rebuilding resolves the content-based tag without rebuilding stages")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{})
		Expect(buildOut).To(ContainSubstring("Reusing image app by content-based tag"))
		Expect(buildOut).NotTo(ContainSubstring("Building stage app/"))
		Expect(buildOut).NotTo(ContainSubstring("Use previously built image for app/from"))
		Expect(buildOut).NotTo(ContainSubstring("Use previously built image for app/install"))
		Expect(buildOut).NotTo(ContainSubstring("Use previously built image for app/setup"))

		By("[3, repo] building with --repo copies only the content-based tag from the :local secondary")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{
					"--repo", repoAddr,
					"--insecure-registry", "--skip-tls-verify-registry",
				},
			},
		})
		Expect(buildOut).To(ContainSubstring("Copy suitable stage from secondary :local"))
		Expect(buildOut).NotTo(ContainSubstring("Building stage app/"))
	})
})
