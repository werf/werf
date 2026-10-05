package gitdata

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = ginkgo.Describe("GetGitArchivesAndRemoveInvalid", func() {
	var root, hashDir string

	tarPath := func(id string) string { return filepath.Join(hashDir, id+".tar") }
	metaPath := func(id string) string { return filepath.Join(hashDir, id+".meta.json") }

	ginkgo.BeforeEach(func() {
		root = ginkgo.GinkgoT().TempDir()
		hashDir = filepath.Join(root, "39e4985a", "29")
	})

	ginkgo.It("returns nil, nil when root does not exist", func(ctx ginkgo.SpecContext) {
		res, err := GetGitArchivesAndRemoveInvalid(ctx, filepath.Join(root, "missing"), ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.BeNil())
	})

	ginkgo.It("yields a valid archive entry scoped to the cache version root", func(ctx ginkgo.SpecContext) {
		meta := writeCacheFile(metaPath("abc"), `{"LastAccessTimestamp":7}`)
		tar := writeCacheFile(tarPath("abc"), "12345")

		res, err := GetGitArchivesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.HaveLen(1))
		gomega.Expect(res[0].GetPaths()).To(gomega.ConsistOf(meta, tar))
		gomega.Expect(res[0].GetSize()).To(gomega.Equal(uint64(5)))
		gomega.Expect(res[0].GetCacheBasePath()).To(gomega.Equal(root))
		gomega.Expect(res[0].GetLastAccessAt().Unix()).To(gomega.Equal(int64(7)))
	})

	ginkgo.DescribeTable("removes both files of a broken entry",
		func(ctx ginkgo.SpecContext, meta, payload *string) {
			if meta != nil {
				writeCacheFile(metaPath("abc"), *meta)
			}
			if payload != nil {
				writeCacheFile(tarPath("abc"), *payload)
			}
			gomega.Expect(os.MkdirAll(hashDir, 0o755)).To(gomega.Succeed())

			res, err := GetGitArchivesAndRemoveInvalid(ctx, root, ScanOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(res).To(gomega.BeEmpty())
			expectGone(metaPath("abc"))
			expectGone(tarPath("abc"))
		},
		ginkgo.Entry("metadata absent", nil, lo.ToPtr("12345")),
		ginkgo.Entry("metadata corrupt", lo.ToPtr("{not json"), lo.ToPtr("12345")),
		ginkgo.Entry("metadata null", lo.ToPtr("null"), lo.ToPtr("12345")),
		ginkgo.Entry("payload absent", lo.ToPtr(`{"LastAccessTimestamp":7}`), nil),
	)

	ginkgo.It("keeps the valid neighbor of a broken entry sharing its id prefix", func(ctx ginkgo.SpecContext) {
		writeCacheFile(metaPath("abc"), "null")
		writeCacheFile(tarPath("abc"), "12345")
		meta := writeCacheFile(metaPath("abcd"), `{"LastAccessTimestamp":7}`)
		tar := writeCacheFile(tarPath("abcd"), "12345")

		res, err := GetGitArchivesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.HaveLen(1))
		gomega.Expect(res[0].GetPaths()).To(gomega.ConsistOf(meta, tar))
		gomega.Expect(tar).To(gomega.BeARegularFile())
		expectGone(metaPath("abc"))
		expectGone(tarPath("abc"))
	})

	ginkgo.It("removes unknown files and directories inside a hash prefix dir", func(ctx ginkgo.SpecContext) {
		junk := writeCacheFile(filepath.Join(hashDir, "abc.junk"), "x")
		dir := filepath.Join(hashDir, "abc.tar.d")
		gomega.Expect(os.MkdirAll(dir, 0o755)).To(gomega.Succeed())

		res, err := GetGitArchivesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res).To(gomega.BeEmpty())
		expectGone(junk)
		expectGone(dir)
	})

	ginkgo.It("preserves everything when a hash prefix dir cannot be read", func(ctx ginkgo.SpecContext) {
		meta := writeCacheFile(metaPath("abc"), `{"LastAccessTimestamp":7}`)
		tar := writeCacheFile(tarPath("abc"), "12345")
		gomega.Expect(os.Chmod(hashDir, 0o000)).To(gomega.Succeed())
		ginkgo.DeferCleanup(func() { gomega.Expect(os.Chmod(hashDir, 0o755)).To(gomega.Succeed()) })

		_, err := GetGitArchivesAndRemoveInvalid(ctx, root, ScanOptions{})
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(os.Chmod(hashDir, 0o755)).To(gomega.Succeed())
		gomega.Expect(meta).To(gomega.BeARegularFile())
		gomega.Expect(tar).To(gomega.BeARegularFile())
	})
})
