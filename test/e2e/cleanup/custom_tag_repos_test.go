package e2e_cleanup_test

import (
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/werf"
)

var _ = ginkgo.Describe("Custom tags in several repositories", ginkgo.Label("e2e", "cleanup", "simple"), func() {
	ginkgo.It("deletes every custom tag from the repository it was added to", func(ctx ginkgo.SpecContext) {
		SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
		SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")
		SuiteData.InitTestRepo(ctx, "repo0", "sibling_custom_tags")
		repoPath := SuiteData.GetTestRepoPath("repo0")
		werfProject := werf.NewProject(SuiteData.WerfBinPath, repoPath)
		primaryRepo := SuiteData.K8sDockerRegistryRepo
		installRepo := primaryRepo + "/install"
		registryOptions := []crane.Option{crane.Insecure, crane.WithContext(ctx)}

		ginkgo.By("publishing both images under the same tag, each in its own repository")
		werfProject.Build(ctx, nil)
		werfProject.RunCommand(ctx, []string{"build", "app", "--require-built-images", "--add-custom-tag", "main"}, werf.CommonOptions{})
		werfProject.RunCommand(ctx, []string{"build", "app/install", "--require-built-images", "--final-repo", installRepo, "--add-custom-tag", "main"}, werf.CommonOptions{})

		appDigest, err := crane.Digest(primaryRepo+":main", registryOptions...)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		installDigest, err := crane.Digest(installRepo+":main", registryOptions...)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(installDigest).NotTo(gomega.Equal(appDigest), "the two images must not share a manifest, otherwise the test proves nothing")

		ginkgo.By("keeping both tags while their stages are kept")
		werfProject.RunCommand(ctx, []string{"cleanup", "--without-kube", "--save-cleanup-report"}, werf.CommonOptions{})
		report := readCleanupReport(repoPath)
		gomega.Expect(report.Kept).To(gomega.ContainElements(
			reportItem{Type: "customTag", Tag: "main"},
			reportItem{Type: "customTag", Tag: installRepo + ":main"},
		))
		gomega.Expect(crane.Digest(primaryRepo+":main", registryOptions...)).To(gomega.Equal(appDigest))
		gomega.Expect(crane.Digest(installRepo+":main", registryOptions...)).To(gomega.Equal(installDigest))

		ginkgo.By("deleting each tag from its own repository once the stages go")
		werfProject.RunCommand(ctx, []string{"cleanup", "--without-kube", "--keep-stages-built-within-last-n-hours=0", "--save-cleanup-report"}, werf.CommonOptions{})
		report = readCleanupReport(repoPath)
		gomega.Expect(report.Deleted).To(gomega.ContainElements(
			reportItem{Type: "customTag", Tag: "main"},
			reportItem{Type: "customTag", Tag: installRepo + ":main"},
		))
		_, err = crane.Digest(primaryRepo+":main", registryOptions...)
		gomega.Expect(err).To(gomega.HaveOccurred())
		_, err = crane.Digest(installRepo+":main", registryOptions...)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
