package gitdata

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("wipeCacheDirs", func() {
	staleTime := time.Now().Add(-cacheVersionStalenessWindow - time.Hour)

	writeFileWithMtime := func(path string, mtime time.Time) {
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, []byte("x"), 0o644)).To(Succeed())
		Expect(os.Chtimes(path, mtime, mtime)).To(Succeed())
	}

	DescribeTable("removes only stale non-kept children",
		func(ctx SpecContext, setup func(root string) string, expectRemoved bool) {
			root := GinkgoT().TempDir()
			child := setup(root)

			Expect(wipeCacheDirs(ctx, root, []string{"5"})).To(Succeed())

			if expectRemoved {
				Expect(child).NotTo(BeAnExistingFile())
			} else {
				Expect(child).To(BeAnExistingFile())
			}
		},
		Entry("keeps a foreign version dir with a fresh nested file",
			func(root string) string {
				dir := filepath.Join(root, "6")
				writeFileWithMtime(filepath.Join(dir, "repo", "last_access_at"), time.Now())
				return dir
			}, false),
		Entry("removes a foreign version dir with only stale files",
			func(root string) string {
				dir := filepath.Join(root, "6")
				writeFileWithMtime(filepath.Join(dir, "repo", "last_access_at"), staleTime)
				return dir
			}, true),
		Entry("removes an empty foreign version dir",
			func(root string) string {
				dir := filepath.Join(root, "6")
				Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
				return dir
			}, true),
		Entry("keeps the current version dir even when stale",
			func(root string) string {
				dir := filepath.Join(root, "5")
				writeFileWithMtime(filepath.Join(dir, "repo", "last_access_at"), staleTime)
				return dir
			}, false),
		Entry("keeps a fresh stray file directly under root",
			func(root string) string {
				path := filepath.Join(root, ".DS_Store")
				writeFileWithMtime(path, time.Now())
				return path
			}, false),
		Entry("removes a stale stray file directly under root",
			func(root string) string {
				path := filepath.Join(root, ".DS_Store")
				writeFileWithMtime(path, staleTime)
				return path
			}, true),
	)
})
