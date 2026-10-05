package gitdata

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Orphan local worktree cleanup", func() {
	ginkgo.BeforeEach(func() {
		ginkgo.GinkgoT().Setenv("WERF_HOME", ginkgo.GinkgoT().TempDir())
		ginkgo.GinkgoT().Setenv("WERF_TMP_DIR", ginkgo.GinkgoT().TempDir())
		gomega.Expect(werf.Init("", "")).To(gomega.Succeed())
	})

	ginkgo.It("prioritizes a stale missing-origin local checkout over reusable artifacts", func() {
		source := filepath.Join(ginkgo.GinkgoT().TempDir(), "deleted", ".git")
		dir := createOrphanWorktreeFixture("local", "orphan", source, time.Now().Add(-4*time.Hour))
		entries, err := collectWorktrees()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(entries).To(gomega.HaveLen(1))
		archive := &GitArchiveDesc{Size: 1, Metadata: &ArchiveMetadata{LastAccessTimestamp: time.Now().Add(-48 * time.Hour).Unix()}}
		candidates := keepGitDataByLru(append(entries, archive))
		gomega.Expect(candidates[0].GetPaths()).To(gomega.Equal([]string{dir}))
	})

	ginkgo.DescribeTable("preserves whitespace in an existing origin path",
		func(suffix string) {
			source := filepath.Join(ginkgo.GinkgoT().TempDir(), "git-dir"+suffix)
			gomega.Expect(os.MkdirAll(source, 0o755)).To(gomega.Succeed())
			createOrphanWorktreeFixture("local", "existing", source+"\n", time.Now().Add(-24*time.Hour))

			entries, err := collectWorktrees()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(entries).To(gomega.HaveLen(1))
			archive := &GitArchiveDesc{Size: 1, Metadata: &ArchiveMetadata{LastAccessTimestamp: time.Now().Add(-4 * time.Hour).Unix()}}
			candidates := keepGitDataByLru(append(entries, archive))
			gomega.Expect(candidates[0]).To(gomega.BeIdenticalTo(archive))
		},
		ginkgo.Entry("space", " "),
		ginkgo.Entry("tab", "\t"),
		ginkgo.Entry("newline", "\n"),
	)

	ginkgo.It("keeps the existing age guard for orphan worktrees", func() {
		createOrphanWorktreeFixture("local", "fresh", filepath.Join(ginkgo.GinkgoT().TempDir(), "gone"), time.Now())
		entries, err := collectWorktrees()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(entries).To(gomega.HaveLen(1))
		gomega.Expect(getGitDataEntryRank(entries[0])).To(gomega.Equal(gitDataEntryRankWorktree))
		candidates := keepGitDataByLru(entries)
		gomega.Expect(candidates).To(gomega.BeEmpty())
	})

	ginkgo.It("does not label remote or unknown-origin checkouts as local orphans", func() {
		for _, kind := range []string{"remote", "local"} {
			source := ""
			if kind == "remote" {
				source = filepath.Join(ginkgo.GinkgoT().TempDir(), "gone")
			}
			createOrphanWorktreeFixture(kind, kind, source, time.Now().Add(-24*time.Hour))
		}
		entries, err := collectWorktrees()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(entries).To(gomega.HaveLen(2))
		archive := &GitArchiveDesc{Size: 1, Metadata: &ArchiveMetadata{LastAccessTimestamp: time.Now().Add(-4 * time.Hour).Unix()}}
		candidates := keepGitDataByLru(append(entries, archive))
		gomega.Expect(candidates[0]).To(gomega.BeIdenticalTo(archive))
	})

	ginkgo.It("rechecks a previously missing origin immediately before removal", func() {
		source := filepath.Join(ginkgo.GinkgoT().TempDir(), "returning", ".git")
		dir := createOrphanWorktreeFixture("local", "restored", source, time.Now().Add(-24*time.Hour))
		entries, err := collectWorktrees()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(entries).To(gomega.HaveLen(1))
		gomega.Expect(os.MkdirAll(source, 0o755)).To(gomega.Succeed())
		freed, err := removeGitDataEntries(context.Background(), entries, removeGitDataEntriesOptions{BytesToFree: 1, GetUsedBytes: func() (uint64, error) { return 0, nil }})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(freed).To(gomega.BeZero())
		gomega.Expect(dir).To(gomega.BeADirectory())
	})

	ginkgo.It("removes a stale orphan while preserving its version scope", func() {
		source := filepath.Join(ginkgo.GinkgoT().TempDir(), "gone")
		dir := createOrphanWorktreeFixture("local", "obsolete", source, time.Now().Add(-24*time.Hour))
		entries, err := collectWorktrees()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(entries).To(gomega.HaveLen(1))
		freed, err := removeGitDataEntries(context.Background(), keepGitDataByLru(entries), removeGitDataEntriesOptions{BytesToFree: 1, GetUsedBytes: func() (uint64, error) { return 0, nil }})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(freed).To(gomega.Equal(entries[0].GetSize()))
		gomega.Expect(dir).NotTo(gomega.BeADirectory())
		gomega.Expect(entries[0].GetCacheBasePath()).To(gomega.BeADirectory())
	})

	ginkgo.It("preserves an entry when access metadata cannot be read", func() {
		dir := createOrphanWorktreeFixture("local", "unreadable", "", time.Now().Add(-24*time.Hour))
		marker := filepath.Join(dir, "last_access_at")
		gomega.Expect(os.Remove(marker)).To(gomega.Succeed())
		gomega.Expect(os.Symlink("last_access_at", marker)).To(gomega.Succeed())
		_, err := collectWorktrees()
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(dir).To(gomega.BeADirectory())
	})

	ginkgo.It("treats an origin it cannot probe as present instead of orphaned", func() {
		source := filepath.Join(ginkgo.GinkgoT().TempDir(), "loop")
		gomega.Expect(os.Symlink("loop", source)).To(gomega.Succeed())
		dir := createOrphanWorktreeFixture("local", "unprobeable-origin", source, time.Now().Add(-24*time.Hour))
		entries, err := collectWorktrees()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(entries).To(gomega.HaveLen(1))
		gomega.Expect(getGitDataEntryRank(entries[0])).To(gomega.Equal(gitDataEntryRankWorktree))
		gomega.Expect(dir).To(gomega.BeADirectory())
	})
})
