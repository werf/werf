package gitdata

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Git cache access errors", func() {
	ginkgo.BeforeEach(func() {
		ginkgo.GinkgoT().Setenv("WERF_HOME", ginkgo.GinkgoT().TempDir())
		ginkgo.GinkgoT().Setenv("WERF_TMP_DIR", ginkgo.GinkgoT().TempDir())
		gomega.Expect(werf.Init("", "")).To(gomega.Succeed())
		gcFixture()
	})
	ginkgo.DescribeTable("preserves the unreadable entry, cleans the readable ones and still fails",
		func(ctx ginkgo.SpecContext, relative, fileName, healthyRelative string) {
			dir := filepath.Join(werf.GetLocalCacheDir(), relative)
			unreadable := filepath.Join(dir, fileName)
			healthy := filepath.Join(werf.GetLocalCacheDir(), healthyRelative)

			gomega.Expect(os.RemoveAll(unreadable)).To(gomega.Succeed())
			gomega.Expect(os.Symlink(fileName, unreadable)).To(gomega.Succeed())

			gomega.Expect(RunGC(ctx, RunGCOptions{})).NotTo(gomega.Succeed())

			gomega.Expect(dir).To(gomega.BeADirectory())
			target, err := os.Readlink(unreadable)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(target).To(gomega.Equal(fileName))

			gomega.Expect(healthy).NotTo(gomega.BeADirectory())
		},
		ginkgo.Entry("full mirror access marker",
			filepath.Join("git_repos", git_repo.GitReposCacheVersion, "valid"), "last_access_at",
			filepath.Join("git_worktrees", git_repo.GitWorktreesCacheVersion, "local", "valid")),
		ginkgo.Entry("shallow mirror access marker",
			filepath.Join("git_mirrors", git_repo.GitMirrorsCacheVersion, "valid", "shallow"), "last_access_at",
			filepath.Join("git_repos", git_repo.GitReposCacheVersion, "valid")),
		ginkgo.Entry("worktree access marker",
			filepath.Join("git_worktrees", git_repo.GitWorktreesCacheVersion, "local", "valid"), "last_access_at",
			filepath.Join("git_repos", git_repo.GitReposCacheVersion, "valid")),
		ginkgo.Entry("worktree origin",
			filepath.Join("git_worktrees", git_repo.GitWorktreesCacheVersion, "local", "valid"), "git_dir",
			filepath.Join("git_repos", git_repo.GitReposCacheVersion, "valid")),
	)
})
