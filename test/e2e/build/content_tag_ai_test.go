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

		By("[1, :local] building all stages and the content tag from scratch")
		buildOut := werfProject.Build(ctx, &werf.BuildOptions{})
		Expect(buildOut).To(ContainSubstring("Building stage"))
		Expect(buildOut).To(ContainSubstring("Building stapel-scratch/content-tag"))
		Expect(buildOut).NotTo(ContainSubstring("Use previously built image for stapel-scratch/content-tag"))

		By("[2, :local] rebuilding reuses the content tag from the local cache")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{})
		Expect(buildOut).To(ContainSubstring("Use previously built image for stapel-scratch/content-tag"))
		Expect(buildOut).NotTo(ContainSubstring("Building stapel-scratch/content-tag"))

		By("[3, repo] building with --repo copies only the content tag from the :local secondary")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{
					"--repo", repoAddr,
					"--insecure-registry", "--skip-tls-verify-registry",
				},
			},
		})
		Expect(buildOut).To(ContainSubstring("Copy suitable stapel-scratch/content-tag from secondary :local"))
		Expect(buildOut).NotTo(ContainSubstring("Building stapel-scratch/content-tag"))

		By("[3, repo] rebuilding with --repo reuses the content tag already present in the repo")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{
					"--require-built-images",
					"--repo", repoAddr,
					"--insecure-registry", "--skip-tls-verify-registry",
				},
			},
		})
		Expect(buildOut).To(ContainSubstring("Use previously built image for stapel-scratch/content-tag"))
		Expect(buildOut).NotTo(ContainSubstring("Building stapel-scratch/content-tag"))

		By("[4, final] building with --final-repo copies only the content tag into the final repo")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{
					"--repo", repoAddr,
					"--final-repo", finalRepoAddr,
					"--insecure-registry", "--skip-tls-verify-registry",
				},
			},
		})
		Expect(buildOut).NotTo(ContainSubstring("Building stapel-scratch/content-tag"))
	})

	It("resolves only the content tag for a multi-stage image without touching individual stages", func(ctx SpecContext) {
		By("initializing")
		setupEnv(setupEnvOptions{})

		repoDirName := "repo0"
		fixtureRelPath := "content_tag/multistage/state0"

		By("preparing test repo")
		SuiteData.InitTestRepo(ctx, repoDirName, fixtureRelPath)
		werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))

		repoAddr := fmt.Sprintf("%s/%s-%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), SuiteData.ProjectName, utils.GetRandomString(6))

		By("[1, :local] building all stages and the content tag from scratch")
		buildOut := werfProject.Build(ctx, &werf.BuildOptions{})
		Expect(buildOut).To(ContainSubstring("Building stage app/from"))
		Expect(buildOut).To(ContainSubstring("Building stage app/install"))
		Expect(buildOut).To(ContainSubstring("Building stage app/setup"))
		Expect(buildOut).To(ContainSubstring("Building app/content-tag"))

		By("[2, :local] rebuilding resolves only the content tag, no individual stages")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{})
		Expect(buildOut).To(ContainSubstring("Use previously built image for app/content-tag"))
		Expect(buildOut).NotTo(ContainSubstring("Use previously built image for app/from"))
		Expect(buildOut).NotTo(ContainSubstring("Use previously built image for app/install"))
		Expect(buildOut).NotTo(ContainSubstring("Use previously built image for app/setup"))

		By("[3, repo] building with --repo copies only the content tag from the :local secondary")
		buildOut = werfProject.Build(ctx, &werf.BuildOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{
					"--repo", repoAddr,
					"--insecure-registry", "--skip-tls-verify-registry",
				},
			},
		})
		Expect(buildOut).To(ContainSubstring("Copy suitable app/content-tag from secondary :local"))
		Expect(buildOut).NotTo(ContainSubstring("Copy suitable stage from secondary"))
		Expect(buildOut).NotTo(ContainSubstring("Building stage app/"))
	})
})
