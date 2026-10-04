package gitdata

import (
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util/timestamps"
)

var _ = ginkgo.Describe("GetGitWorktreesAndRemoveInvalid", func() {
	staleTime := time.Now().Add(-4 * time.Hour)

	writeWorktree := func(root, name string, ts time.Time, submoduleFiles ...string) string {
		cacheDir := filepath.Join(root, "local", name)
		gomega.Expect(os.MkdirAll(filepath.Join(cacheDir, "worktree"), 0o755)).To(gomega.Succeed())
		gomega.Expect(timestamps.WriteTimestampFile(filepath.Join(cacheDir, "last_access_at"), ts)).To(gomega.Succeed())
		for _, path := range submoduleFiles {
			full := filepath.Join(cacheDir, path)
			gomega.Expect(os.MkdirAll(filepath.Dir(full), 0o755)).To(gomega.Succeed())
			gomega.Expect(os.WriteFile(full, []byte("[submodule \"x\"]\n"), 0o644)).To(gomega.Succeed())
		}
		return cacheDir
	}

	findEntry := func(entries []GitDataEntry, path string) *GitWorktreeDesc {
		for _, entry := range entries {
			if desc, ok := entry.(*GitWorktreeDesc); ok && desc.Path == path {
				return desc
			}
		}
		ginkgo.Fail("no entry for " + path)
		return nil
	}

	ginkgo.It("marks only the worktree carrying .gitmodules as having submodules", func(ctx ginkgo.SpecContext) {
		root := ginkgo.GinkgoT().TempDir()
		plain := writeWorktree(root, "plain", staleTime)
		withSubmodules := writeWorktree(root, "submodules", staleTime, filepath.Join("worktree", ".gitmodules"))
		// .gitmodules of a nested submodule checkout, not of the worktree itself.
		nested := writeWorktree(root, "nested", staleTime, filepath.Join("worktree", "sub", ".gitmodules"))

		res, err := GetGitWorktreesAndRemoveInvalid(ctx, root)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.HaveLen(3))
		gomega.Expect(findEntry(res, plain).HasSubmodules).To(gomega.BeFalse())
		gomega.Expect(findEntry(res, withSubmodules).HasSubmodules).To(gomega.BeTrue())
		gomega.Expect(findEntry(res, nested).HasSubmodules).To(gomega.BeFalse())
		gomega.Expect(getGitDataEntryRank(findEntry(res, withSubmodules))).To(gomega.Equal(gitDataEntryRankWorktreeWithSubmodules))
		gomega.Expect(getGitDataEntryRank(findEntry(res, plain))).To(gomega.Equal(gitDataEntryRankWorktree))
	})

	ginkgo.It("does not classify worktrees inside the preservation window, which are never removed anyway", func(ctx ginkgo.SpecContext) {
		root := ginkgo.GinkgoT().TempDir()
		fresh := writeWorktree(root, "fresh", time.Now(), filepath.Join("worktree", ".gitmodules"))

		res, err := GetGitWorktreesAndRemoveInvalid(ctx, root)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(findEntry(res, fresh).HasSubmodules).To(gomega.BeFalse())
		gomega.Expect(keepGitDataByLru(res)).To(gomega.BeEmpty())
	})
})
