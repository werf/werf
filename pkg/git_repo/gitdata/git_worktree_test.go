package gitdata

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util/timestamps"
)

var _ = Describe("GetGitWorktreesAndRemoveInvalid", func() {
	staleTime := time.Now().Add(-4 * time.Hour)

	writeWorktree := func(root, name string, ts time.Time, submoduleFiles ...string) string {
		cacheDir := filepath.Join(root, "local", name)
		Expect(os.MkdirAll(filepath.Join(cacheDir, "worktree"), 0o755)).To(Succeed())
		Expect(timestamps.WriteTimestampFile(filepath.Join(cacheDir, "last_access_at"), ts)).To(Succeed())
		for _, path := range submoduleFiles {
			full := filepath.Join(cacheDir, path)
			Expect(os.MkdirAll(filepath.Dir(full), 0o755)).To(Succeed())
			Expect(os.WriteFile(full, []byte("[submodule \"x\"]\n"), 0o644)).To(Succeed())
		}
		return cacheDir
	}

	findEntry := func(entries []GitDataEntry, path string) *GitWorktreeDesc {
		for _, entry := range entries {
			if desc, ok := entry.(*GitWorktreeDesc); ok && desc.Path == path {
				return desc
			}
		}
		Fail("no entry for " + path)
		return nil
	}

	It("marks only the worktree carrying .gitmodules as having submodules", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		plain := writeWorktree(root, "plain", staleTime)
		withSubmodules := writeWorktree(root, "submodules", staleTime, filepath.Join("worktree", ".gitmodules"))
		// .gitmodules of a nested submodule checkout, not of the worktree itself.
		nested := writeWorktree(root, "nested", staleTime, filepath.Join("worktree", "sub", ".gitmodules"))

		res, err := GetGitWorktreesAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(HaveLen(3))
		Expect(findEntry(res, plain).HasSubmodules).To(BeFalse())
		Expect(findEntry(res, withSubmodules).HasSubmodules).To(BeTrue())
		Expect(findEntry(res, nested).HasSubmodules).To(BeFalse())
		Expect(getGitDataEntryRank(findEntry(res, withSubmodules))).To(Equal(gitDataEntryRankWorktreeWithSubmodules))
		Expect(getGitDataEntryRank(findEntry(res, plain))).To(Equal(gitDataEntryRankWorktree))
	})

	It("does not classify worktrees inside the preservation window, which are never removed anyway", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		fresh := writeWorktree(root, "fresh", time.Now(), filepath.Join("worktree", ".gitmodules"))

		res, err := GetGitWorktreesAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(findEntry(res, fresh).HasSubmodules).To(BeFalse())
		Expect(keepGitDataByLru(res)).To(BeEmpty())
	})
})
