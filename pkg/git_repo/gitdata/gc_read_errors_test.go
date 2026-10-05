package gitdata

import (
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = ginkgo.Describe("Git cache access errors", func() {
	ginkgo.BeforeEach(func() {
		ginkgo.GinkgoT().Setenv("WERF_HOME", ginkgo.GinkgoT().TempDir())
		ginkgo.GinkgoT().Setenv("WERF_TMP_DIR", ginkgo.GinkgoT().TempDir())
		gomega.Expect(werf.Init("", "")).To(gomega.Succeed())
		gcFixture()
	})
	ginkgo.DescribeTable("preserves the unreadable entry, cleans the readable ones and still fails",
		func(ctx ginkgo.SpecContext, relative, fileName, siblingRelative string) {
			dir := filepath.Join(werf.GetLocalCacheDir(), relative)
			unreadable := filepath.Join(dir, fileName)

			gomega.Expect(os.RemoveAll(unreadable)).To(gomega.Succeed())
			gomega.Expect(os.Symlink(fileName, unreadable)).To(gomega.Succeed())

			// A readable entry collected by the same collector must still be
			// reclaimed under volume pressure.
			sibling := writeMirror(filepath.Join(werf.GetLocalCacheDir(), siblingRelative), time.Now().Add(-24*time.Hour))
			writeCacheFile(filepath.Join(sibling, "payload"), "sibling payload")

			gomega.Expect(RunGC(ctx, RunGCOptions{})).NotTo(gomega.Succeed())

			gomega.Expect(dir).To(gomega.BeADirectory())
			target, err := os.Readlink(unreadable)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(target).To(gomega.Equal(fileName))

			expectGone(sibling)
		},
		ginkgo.Entry("full mirror access marker",
			filepath.Join("git_repos", git_repo.GitReposCacheVersion, "valid"), "last_access_at",
			filepath.Join("git_repos", git_repo.GitReposCacheVersion, "sibling")),
		ginkgo.Entry("shallow mirror access marker",
			filepath.Join("git_mirrors", git_repo.GitMirrorsCacheVersion, "valid", "shallow"), "last_access_at",
			filepath.Join("git_mirrors", git_repo.GitMirrorsCacheVersion, "sibling", "shallow")),
		ginkgo.Entry("worktree access marker",
			filepath.Join("git_worktrees", git_repo.GitWorktreesCacheVersion, "local", "valid"), "last_access_at",
			filepath.Join("git_worktrees", git_repo.GitWorktreesCacheVersion, "local", "sibling")),
		ginkgo.Entry("worktree origin",
			filepath.Join("git_worktrees", git_repo.GitWorktreesCacheVersion, "local", "valid"), "git_dir",
			filepath.Join("git_worktrees", git_repo.GitWorktreesCacheVersion, "local", "sibling")),
	)
})
