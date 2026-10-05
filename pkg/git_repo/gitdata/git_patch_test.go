package gitdata

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = ginkgo.Describe("GetGitPatchesAndRemoveInvalid", func() {
	var root, hashDir string

	patchPath := func(id string) string { return filepath.Join(hashDir, id+".patch") }
	metaPath := func(id string) string { return filepath.Join(hashDir, id+".meta.json") }
	sidecarPath := func(id, paramsHash, suffix string) string {
		return filepath.Join(hashDir, id+".patch."+paramsHash+suffix)
	}

	writeEntry := func(id, meta, payload string) (string, string) {
		return writeCacheFile(metaPath(id), meta), writeCacheFile(patchPath(id), payload)
	}

	ginkgo.BeforeEach(func() {
		root = ginkgo.GinkgoT().TempDir()
		hashDir = filepath.Join(root, "0f1ddce0", "ab")
	})

	ginkgo.It("returns nil, nil when root does not exist", func(ctx ginkgo.SpecContext) {
		res, err := GetGitPatchesAndRemoveInvalid(ctx, filepath.Join(root, "missing"), ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.BeNil())
	})

	ginkgo.It("keeps the filtered archive and paths_list sidecars of a valid patch and accounts for them", func(ctx ginkgo.SpecContext) {
		meta, patch := writeEntry("abc", `{"LastAccessTimestamp":1}`, "12345")
		archive := writeCacheFile(sidecarPath("abc", "p1", ".archive"), "123")
		pathsList := writeCacheFile(sidecarPath("abc", "p1", ".paths_list"), "12")

		res, err := GetGitPatchesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.HaveLen(1))
		gomega.Expect(res[0].GetPaths()).To(gomega.ConsistOf(meta, patch, archive, pathsList))
		gomega.Expect(res[0].GetSize()).To(gomega.Equal(uint64(10)))
		gomega.Expect(res[0].GetCacheBasePath()).To(gomega.Equal(root))
		gomega.Expect(res[0].GetLastAccessAt().Unix()).To(gomega.Equal(int64(1)))
		gomega.Expect(archive).To(gomega.BeARegularFile())
		gomega.Expect(pathsList).To(gomega.BeARegularFile())
	})

	ginkgo.It("yields entries that removeGitDataEntries actually deletes, sidecars included", func(ctx ginkgo.SpecContext) {
		meta, patch := writeEntry("abc", `{"LastAccessTimestamp":1}`, "12345")
		archive := writeCacheFile(sidecarPath("abc", "p1", ".archive"), "123")

		res, err := GetGitPatchesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.HaveLen(1))

		freed, err := removeGitDataEntries(ctx, res, removeGitDataEntriesOptions{
			BytesToFree:            1 << 20,
			TargetVolumeUsageBytes: 0,
			GetUsedBytes:           func() (uint64, error) { return 1 << 30, nil },
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(freed).To(gomega.Equal(uint64(8)))
		expectGone(meta)
		expectGone(patch)
		expectGone(archive)
		expectGone(hashDir)

		expectGone(filepath.Dir(hashDir))
		gomega.Expect(root).To(gomega.BeADirectory())
	})

	ginkgo.DescribeTable("removes the whole broken entry including its sidecars",
		func(ctx ginkgo.SpecContext, meta, payload *string) {
			if meta != nil {
				writeCacheFile(metaPath("abc"), *meta)
			}
			if payload != nil {
				writeCacheFile(patchPath("abc"), *payload)
			}
			archive := writeCacheFile(sidecarPath("abc", "p1", ".archive"), "123")
			pathsList := writeCacheFile(sidecarPath("abc", "p1", ".paths_list"), "12")

			res, err := GetGitPatchesAndRemoveInvalid(ctx, root, ScanOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(res).To(gomega.BeEmpty())
			expectGone(metaPath("abc"))
			expectGone(patchPath("abc"))
			expectGone(archive)
			expectGone(pathsList)
		},
		ginkgo.Entry("metadata absent", nil, lo.ToPtr("12345")),
		ginkgo.Entry("metadata corrupt", lo.ToPtr("{not json"), lo.ToPtr("12345")),
		ginkgo.Entry("metadata null", lo.ToPtr("null"), lo.ToPtr("12345")),
		ginkgo.Entry("payload absent", lo.ToPtr(`{"LastAccessTimestamp":1}`), nil),
		ginkgo.Entry("both absent", nil, nil),
	)

	ginkgo.It("keeps an entry whose metadata carries no timestamp", func(ctx ginkgo.SpecContext) {
		meta, patch := writeEntry("abc", `{}`, "12345")

		res, err := GetGitPatchesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.HaveLen(1))
		gomega.Expect(res[0].GetLastAccessAt().Unix()).To(gomega.Equal(int64(0)))
		gomega.Expect(res[0].GetPaths()).To(gomega.ConsistOf(meta, patch))
	})

	ginkgo.It("removes only the sidecars of the broken entry when another id shares its prefix", func(ctx ginkgo.SpecContext) {
		writeCacheFile(metaPath("abc"), "{not json")
		writeCacheFile(patchPath("abc"), "12345")
		brokenArchive := writeCacheFile(sidecarPath("abc", "p1", ".archive"), "123")

		meta, patch := writeEntry("abcd", `{"LastAccessTimestamp":1}`, "12345")
		keptArchive := writeCacheFile(sidecarPath("abcd", "p1", ".archive"), "123")
		keptPathsList := writeCacheFile(sidecarPath("abcd", "p1", ".paths_list"), "12")

		res, err := GetGitPatchesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.HaveLen(1))
		gomega.Expect(res[0].GetPaths()).To(gomega.ConsistOf(meta, patch, keptArchive, keptPathsList))
		gomega.Expect(keptArchive).To(gomega.BeARegularFile())
		gomega.Expect(keptPathsList).To(gomega.BeARegularFile())
		expectGone(brokenArchive)
		expectGone(metaPath("abc"))
		expectGone(patchPath("abc"))
	})

	ginkgo.It("does not let a valid entry claim the sidecars of a longer id sharing its prefix", func(ctx ginkgo.SpecContext) {
		meta, patch := writeEntry("abc", `{"LastAccessTimestamp":1}`, "12345")
		keptArchive := writeCacheFile(sidecarPath("abc", "p1", ".archive"), "123")

		writeCacheFile(metaPath("abcd"), "{not json")
		writeCacheFile(patchPath("abcd"), "12345")
		brokenArchive := writeCacheFile(sidecarPath("abcd", "p1", ".archive"), "123")

		res, err := GetGitPatchesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.HaveLen(1))
		gomega.Expect(res[0].GetPaths()).To(gomega.ConsistOf(meta, patch, keptArchive))
		gomega.Expect(res[0].GetSize()).To(gomega.Equal(uint64(8)))
		gomega.Expect(keptArchive).To(gomega.BeARegularFile())
		expectGone(brokenArchive)
		expectGone(metaPath("abcd"))
		expectGone(patchPath("abcd"))
	})

	ginkgo.It("removes unknown files and directories inside a hash group dir", func(ctx ginkgo.SpecContext) {
		junk := writeCacheFile(filepath.Join(hashDir, "abc.junk"), "x")
		orphanSidecar := writeCacheFile(filepath.Join(hashDir, "abc.p1.paths_list"), "x")
		dir := filepath.Join(hashDir, "abc.patch.p1.archive.d")
		gomega.Expect(os.MkdirAll(dir, 0o755)).To(gomega.Succeed())

		res, err := GetGitPatchesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.BeEmpty())
		expectGone(junk)
		expectGone(orphanSidecar)
		expectGone(dir)
	})

	ginkgo.It("preserves everything when a hash group dir cannot be read", func(ctx ginkgo.SpecContext) {
		meta, patch := writeEntry("abc", `{"LastAccessTimestamp":1}`, "12345")
		gomega.Expect(os.Chmod(hashDir, 0o000)).To(gomega.Succeed())
		ginkgo.DeferCleanup(func() { gomega.Expect(os.Chmod(hashDir, 0o755)).To(gomega.Succeed()) })

		_, err := GetGitPatchesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(os.Chmod(hashDir, 0o755)).To(gomega.Succeed())
		gomega.Expect(meta).To(gomega.BeARegularFile())
		gomega.Expect(patch).To(gomega.BeARegularFile())
	})
})
