package gitdata

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("removeGitDataEntries", func() {
	var root string

	newEntry := func(name string, size uint64) *GitWorktreeDesc {
		path := filepath.Join(root, name)
		Expect(os.MkdirAll(path, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(path, "data"), []byte("x"), 0o644)).To(Succeed())
		return &GitWorktreeDesc{Path: path, Size: size, CacheBasePath: root}
	}

	constUsage := func(usedBytes uint64) func() (uint64, error) {
		return func() (uint64, error) { return usedBytes, nil }
	}

	BeforeEach(func() {
		root = GinkgoT().TempDir()
	})

	It("stops once the estimated budget is covered and keeps the rest", func(ctx SpecContext) {
		first := newEntry("first", 60)
		second := newEntry("second", 60)
		third := newEntry("third", 60)

		// Usage never drops: hard-linked or snapshotted data.
		freed, err := removeGitDataEntries(ctx, []GitDataEntry{first, second, third}, removeGitDataEntriesOptions{
			BytesToFree:            100,
			TargetVolumeUsageBytes: 10,
			GetUsedBytes:           constUsage(1000),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(freed).To(Equal(uint64(120)))
		Expect(first.Path).NotTo(BeAnExistingFile())
		Expect(second.Path).NotTo(BeAnExistingFile())
		Expect(third.Path).To(BeADirectory())
	})

	It("stops early when the volume usage actually reached the target", func(ctx SpecContext) {
		first := newEntry("first", 10)
		second := newEntry("second", 10)

		freed, err := removeGitDataEntries(ctx, []GitDataEntry{first, second}, removeGitDataEntriesOptions{
			BytesToFree:            1000,
			TargetVolumeUsageBytes: 500,
			GetUsedBytes:           constUsage(500),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(freed).To(Equal(uint64(10)))
		Expect(first.Path).NotTo(BeAnExistingFile())
		Expect(second.Path).To(BeADirectory())
	})

	It("continues by the estimate when the volume usage cannot be read", func(ctx SpecContext) {
		first := newEntry("first", 10)
		second := newEntry("second", 10)

		freed, err := removeGitDataEntries(ctx, []GitDataEntry{first, second}, removeGitDataEntriesOptions{
			BytesToFree:            20,
			TargetVolumeUsageBytes: 500,
			GetUsedBytes:           func() (uint64, error) { return 0, errors.New("statfs failed") },
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(freed).To(Equal(uint64(20)))
		Expect(second.Path).NotTo(BeAnExistingFile())
	})

	It("terminates on zero-size entries without freeing anything measurable", func(ctx SpecContext) {
		first := newEntry("first", 0)
		second := newEntry("second", 0)

		freed, err := removeGitDataEntries(ctx, []GitDataEntry{first, second}, removeGitDataEntriesOptions{
			BytesToFree:            100,
			TargetVolumeUsageBytes: 10,
			GetUsedBytes:           constUsage(1000),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(freed).To(BeZero())
		Expect(first.Path).NotTo(BeAnExistingFile())
		Expect(second.Path).NotTo(BeAnExistingFile())
	})

	It("removes nothing on dry run", func(ctx SpecContext) {
		entry := newEntry("first", 10)

		freed, err := removeGitDataEntries(ctx, []GitDataEntry{entry}, removeGitDataEntriesOptions{
			BytesToFree:            1000,
			TargetVolumeUsageBytes: 10,
			DryRun:                 true,
			GetUsedBytes:           func() (uint64, error) { Fail("volume usage must not be read on dry run"); return 0, nil },
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(freed).To(Equal(uint64(10)))
		Expect(entry.Path).To(BeADirectory())
	})
})

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
