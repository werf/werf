package true_git

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/opstats"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Git preparation operations", func() {
	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		gomega.Expect(Init(ctx, Options{})).To(gomega.Succeed())
		isolateGitConfig()
	})

	ginkgo.DescribeTable("finishes lock timing before the protected callback",
		func(ctx ginkgo.SpecContext, withLock func(context.Context, string, func() error) error) {
			collector := opstats.NewCollector()
			observedCtx := opstats.NewContext(ctx, collector)
			callbackErr := errors.New("callback failed")
			var atCallback []opstats.OperationSummary
			err := withLock(observedCtx, ginkgo.GinkgoT().TempDir(), func() error {
				atCallback = collector.Summary()
				gomega.Expect(atCallback).To(gomega.ConsistOf(gomega.SatisfyAll(
					gomega.HaveField("Operation", opstats.OperationGitLockWait),
					gomega.HaveField("Count", 1),
				)))
				return callbackErr
			})
			gomega.Expect(err).To(gomega.MatchError(callbackErr))
			gomega.Expect(collector.Summary()).To(gomega.Equal(atCallback))
		},
		ginkgo.Entry("worktree cache", withWorkTreeCacheLock),
		ginkgo.Entry("repository directory", withRepoDirLock),
	)

	ginkgo.It("finishes worktree timing before consuming the prepared files", func(ctx ginkgo.SpecContext) {
		repoDir := filepath.Join(SuiteData.TestDirPath, "repo")
		gitInitRepo(ctx, repoDir)
		commit := gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")
		collector := opstats.NewCollector()
		observedCtx := opstats.NewContext(ctx, collector)
		callbackErr := errors.New("consumer failed")
		var atCallback []opstats.OperationSummary
		err := WithWorkTree(observedCtx, filepath.Join(repoDir, ".git"), filepath.Join(SuiteData.TestDirPath, "cache"), commit, WithWorkTreeOptions{}, func(worktree string) error {
			gomega.Expect(filepath.Join(worktree, ".git")).To(gomega.BeARegularFile())
			atCallback = collector.Summary()
			gomega.Expect(atCallback).To(gomega.ContainElement(gomega.SatisfyAll(
				gomega.HaveField("Operation", opstats.OperationGitWorktree),
				gomega.HaveField("Count", 1),
			)))
			return callbackErr
		})
		gomega.Expect(err).To(gomega.MatchError(callbackErr))
		gomega.Expect(collector.Summary()).To(gomega.Equal(atCallback))
	})

	ginkgo.It("records failed worktree preparation", func(ctx ginkgo.SpecContext) {
		cache := filepath.Join(SuiteData.TestDirPath, "not-a-directory")
		gomega.Expect(os.WriteFile(cache, nil, 0o600)).To(gomega.Succeed())
		collector := opstats.NewCollector()
		_, err := prepareWorkTree(opstats.NewContext(ctx, collector), SuiteData.TestDirPath, cache, "HEAD", false)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(collector.Summary()).To(gomega.ConsistOf(gomega.SatisfyAll(
			gomega.HaveField("Operation", opstats.OperationGitWorktree),
			gomega.HaveField("Count", 1),
		)))
	})

	ginkgo.It("records failed submodule updates", func(ctx ginkgo.SpecContext) {
		repoDir := filepath.Join(SuiteData.TestDirPath, "repo")
		gitInitRepo(ctx, repoDir)
		commit := gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")
		gomega.Expect(os.WriteFile(filepath.Join(repoDir, ".gitmodules"), []byte("[submodule \"sub\"]\npath = sub\nurl = forbidden://sub\n"), 0o600)).To(gomega.Succeed())
		gitSucceed(ctx, repoDir, "update-index", "--add", "--cacheinfo", "160000,"+commit+",sub")
		gitSucceed(ctx, repoDir, "add", ".gitmodules")
		gitSucceed(ctx, repoDir, "commit", "-m", "add submodule")
		collector := opstats.NewCollector()
		err := updateSubmodules(opstats.NewContext(ctx, collector), filepath.Join(repoDir, ".git"), repoDir)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("transport 'forbidden' not allowed")))
		gomega.Expect(collector.Summary()).To(gomega.ConsistOf(gomega.SatisfyAll(
			gomega.HaveField("Operation", opstats.OperationGitSubmoduleUpdate),
			gomega.HaveField("Count", 1),
		)))
	})
})
