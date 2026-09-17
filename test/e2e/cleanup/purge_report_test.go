package e2e_cleanup_test

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type reportItem struct {
	Type   string `json:"type"`
	Tag    string `json:"tag"`
	Reason string `json:"reason"`
}

type cleanupReport struct {
	Command   string       `json:"command"`
	DryRun    bool         `json:"dryRun"`
	Repo      string       `json:"repo"`
	FinalRepo string       `json:"finalRepo"`
	Kept      []reportItem `json:"kept"`
	Deleted   []reportItem `json:"deleted"`
}

var _ = ginkgo.Describe("Cleanup report", ginkgo.Label("e2e", "cleanup", "simple"), func() {
	ginkgo.It("should describe what werf purge deleted from the repo", func(ctx ginkgo.SpecContext) {
		ginkgo.By("initializing")
		SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
		SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")

		const repoDirname = "repo0"
		SuiteData.InitTestRepo(ctx, repoDirname, "purge_report")
		repoPath := SuiteData.GetTestRepoPath(repoDirname)
		werfProject := werf.NewProject(SuiteData.WerfBinPath, repoPath)

		ginkgo.By("building images")
		werfProject.Build(ctx, nil)

		ginkgo.By("purging with a report")
		purgeOut := werfProject.RunCommand(ctx, []string{"purge", "--save-cleanup-report"}, werf.CommonOptions{})

		ginkgo.By("reading the report")
		data, err := os.ReadFile(filepath.Join(repoPath, ".werf-cleanup-report.json"))
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		var report cleanupReport
		gomega.Expect(json.Unmarshal(data, &report)).To(gomega.Succeed())

		gomega.Expect(report.Command).To(gomega.Equal("purge"))
		gomega.Expect(report.DryRun).To(gomega.BeFalse())
		gomega.Expect(report.Repo).To(gomega.Equal(SuiteData.K8sDockerRegistryRepo))

		ginkgo.By("checking the deleted stages against the log")
		stageTags := lo.FilterMap(report.Deleted, func(item reportItem, _ int) (string, bool) {
			return item.Tag, item.Type == "stage"
		})
		gomega.Expect(stageTags).ShouldNot(gomega.BeEmpty())
		for _, tag := range stageTags {
			gomega.Expect(purgeOut).To(gomega.ContainSubstring(tag))
		}

		gomega.Expect(lo.Map(report.Deleted, func(item reportItem, _ int) string {
			return item.Type
		})).To(gomega.ContainElements("managedImage", "imageMetadata"))
	})

	ginkgo.DescribeTable("cleans final images according to surviving primary stages",
		func(ctx ginkgo.SpecContext, removePrimaryBeforeCleanup bool) {
			SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
			SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")
			SuiteData.InitTestRepo(ctx, "repo0", "final_repo")
			repoPath := SuiteData.GetTestRepoPath("repo0")
			utils.RunSucceedCommand(ctx, repoPath, "git", "remote", "add", "origin", repoPath)
			utils.RunSucceedCommand(ctx, repoPath, "git", "fetch", "origin")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, repoPath)
			primaryRepo := SuiteData.K8sDockerRegistryRepo
			finalRepo := primaryRepo + "-final"
			registryOptions := []crane.Option{crane.Insecure, crane.WithContext(ctx)}

			ginkgo.By("publishing the same stage in primary and final repositories")
			werfProject.RunCommand(ctx, []string{"build", "--final-repo", finalRepo}, werf.CommonOptions{})
			finalTags, err := crane.ListTags(finalRepo, registryOptions...)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(finalTags).To(gomega.HaveLen(1))
			stageTag := finalTags[0]
			primaryImage := primaryRepo + ":" + stageTag
			finalImage := finalRepo + ":" + stageTag
			primaryDigest, err := crane.Digest(primaryImage, registryOptions...)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			finalDigest, err := crane.Digest(finalImage, registryOptions...)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(finalDigest).To(gomega.Equal(primaryDigest))

			cleanupArgs := []string{"cleanup", "--final-repo", finalRepo, "--without-kube", "--save-cleanup-report"}
			for _, dryRun := range []bool{true, false} {
				ginkgo.By("retaining a recently built primary stage and its final counterpart")
				args := append([]string{}, cleanupArgs...)
				if dryRun {
					args = append(args, "--dry-run")
				}
				werfProject.RunCommand(ctx, args, werf.CommonOptions{})
				report := readCleanupReport(repoPath)
				gomega.Expect(report.Command).To(gomega.Equal("cleanup"))
				gomega.Expect(report.DryRun).To(gomega.Equal(dryRun))
				gomega.Expect(report.Repo).To(gomega.Equal(primaryRepo))
				gomega.Expect(report.FinalRepo).To(gomega.Equal(finalRepo))
				gomega.Expect(report.Kept).To(gomega.ContainElements(
					reportItem{Type: "stage", Tag: stageTag, Reason: "built within last 2 hours"},
					reportItem{Type: "finalStage", Tag: stageTag, Reason: "found in repo"},
				))
				gomega.Expect(report.Deleted).To(gomega.BeEmpty())
				gomega.Expect(crane.Digest(finalImage, registryOptions...)).To(gomega.Equal(finalDigest))
			}

			if removePrimaryBeforeCleanup {
				ginkgo.By("removing the primary counterpart before cleanup")
				gomega.Expect(crane.Delete(primaryImage, registryOptions...)).To(gomega.Succeed())
			}

			for _, dryRun := range []bool{true, false} {
				ginkgo.By("deleting an unprotected final stage after its primary counterpart disappears")
				args := append([]string{}, cleanupArgs...)
				args = append(args, "--keep-stages-built-within-last-n-hours=0")
				if dryRun {
					args = append(args, "--dry-run")
				}
				werfProject.RunCommand(ctx, args, werf.CommonOptions{})
				report := readCleanupReport(repoPath)
				gomega.Expect(report.DryRun).To(gomega.Equal(dryRun))
				gomega.Expect(report.Kept).To(gomega.BeEmpty())
				gomega.Expect(report.Deleted).To(gomega.ContainElement(reportItem{Type: "finalStage", Tag: stageTag}))
				if !removePrimaryBeforeCleanup {
					gomega.Expect(report.Deleted).To(gomega.ContainElement(reportItem{Type: "stage", Tag: stageTag}))
				}
				if dryRun {
					gomega.Expect(crane.Digest(finalImage, registryOptions...)).To(gomega.Equal(finalDigest))
					if !removePrimaryBeforeCleanup {
						gomega.Expect(crane.Digest(primaryImage, registryOptions...)).To(gomega.Equal(primaryDigest))
					}
				} else {
					gomega.Expect(crane.ListTags(finalRepo, registryOptions...)).To(gomega.BeEmpty())
					primaryTags, err := crane.ListTags(primaryRepo, registryOptions...)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					gomega.Expect(primaryTags).NotTo(gomega.ContainElement(stageTag))
				}
			}
		},
		ginkgo.Entry("when primary cleanup removes the counterpart in the same run", false),
		ginkgo.Entry("when the final stage is already orphaned", true),
	)
})
