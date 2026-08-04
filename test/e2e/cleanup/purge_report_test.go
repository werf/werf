package e2e_cleanup_test

import (
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/werf/werf/v2/test/pkg/werf"
)

type reportItem struct {
	Type string `json:"type"`
	Tag  string `json:"tag"`
}

type cleanupReport struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Command    string       `json:"command"`
	DryRun     bool         `json:"dryRun"`
	Repo       string       `json:"repo"`
	Deleted    []reportItem `json:"deleted"`
}

var _ = Describe("Cleanup report", Label("e2e", "cleanup", "simple"), func() {
	It("should describe what werf purge deleted from the repo", func(ctx SpecContext) {
		By("initializing")
		SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
		SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")

		const repoDirname = "repo0"
		SuiteData.InitTestRepo(ctx, repoDirname, "purge_report")
		repoPath := SuiteData.GetTestRepoPath(repoDirname)
		werfProject := werf.NewProject(SuiteData.WerfBinPath, repoPath)

		By("building images")
		werfProject.Build(ctx, nil)

		By("purging with a report")
		purgeOut := werfProject.RunCommand(ctx, []string{"purge", "--save-cleanup-report"}, werf.CommonOptions{})

		By("reading the report")
		data, err := os.ReadFile(filepath.Join(repoPath, ".werf-cleanup-report.json"))
		Expect(err).ShouldNot(HaveOccurred())

		var report cleanupReport
		Expect(json.Unmarshal(data, &report)).To(Succeed())

		Expect(report.APIVersion).To(Equal("v1"))
		Expect(report.Kind).To(Equal("CleanupReport"))
		Expect(report.Command).To(Equal("purge"))
		Expect(report.DryRun).To(BeFalse())
		Expect(report.Repo).To(Equal(SuiteData.K8sDockerRegistryRepo))

		By("checking the deleted stages against the log")
		stageTags := lo.FilterMap(report.Deleted, func(item reportItem, _ int) (string, bool) {
			return item.Tag, item.Type == "stage"
		})
		Expect(stageTags).ShouldNot(BeEmpty())
		for _, tag := range stageTags {
			Expect(purgeOut).To(ContainSubstring(tag))
		}

		Expect(lo.Map(report.Deleted, func(item reportItem, _ int) string {
			return item.Type
		})).To(ContainElements("managedImage", "imageMetadata"))
	})
})
