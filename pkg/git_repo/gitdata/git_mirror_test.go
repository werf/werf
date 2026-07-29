package gitdata

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetGitMirrorsAndRemoveInvalid", func() {
	It("returns nil, nil when root does not exist", func(ctx SpecContext) {
		res, err := GetGitMirrorsAndRemoveInvalid(ctx, filepath.Join(GinkgoT().TempDir(), "missing"))
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeNil())
	})

	It("keeps a repo dir holding only the requires_full marker without yielding an entry", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		repo := filepath.Join(root, "abc")
		Expect(os.MkdirAll(repo, 0o755)).To(Succeed())
		marker := filepath.Join(repo, "requires_full")
		Expect(os.WriteFile(marker, nil, 0o644)).To(Succeed())

		res, err := GetGitMirrorsAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeEmpty())
		Expect(marker).To(BeARegularFile())
	})

	It("yields one entry for a shallow-only repo dir", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		now := time.Now().Truncate(time.Second)
		shallow := writeMirror(filepath.Join(root, "abc", "shallow"), now)

		res, err := GetGitMirrorsAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(HaveLen(1))
		Expect(res[0].GetPaths()).To(ConsistOf(shallow))
		Expect(res[0].GetCacheBasePath()).To(Equal(root))
		Expect(res[0].GetLastAccessAt().Unix()).To(Equal(now.Unix()))
	})

	It("yields exactly one entry for marker plus shallow and preserves the marker", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		now := time.Now().Truncate(time.Second)
		shallow := writeMirror(filepath.Join(root, "abc", "shallow"), now)
		marker := filepath.Join(root, "abc", "requires_full")
		Expect(os.WriteFile(marker, nil, 0o644)).To(Succeed())

		res, err := GetGitMirrorsAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(HaveLen(1))
		Expect(res[0].GetPaths()).To(ConsistOf(shallow))
		Expect(marker).To(BeARegularFile())
	})

	It("removes an empty repo dir", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		repo := filepath.Join(root, "abc")
		Expect(os.MkdirAll(repo, 0o755)).To(Succeed())

		res, err := GetGitMirrorsAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeEmpty())
		Expect(repo).NotTo(BeAnExistingFile())
	})

	It("removes non-directory entries directly under root", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		stray := filepath.Join(root, "stray")
		Expect(os.WriteFile(stray, []byte("x"), 0o644)).To(Succeed())

		res, err := GetGitMirrorsAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeEmpty())
		Expect(stray).NotTo(BeAnExistingFile())
	})

	It("removes a malformed shallow that is not a directory, then the repo dir with nothing valid left", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		repo := filepath.Join(root, "abc")
		Expect(os.MkdirAll(repo, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, "shallow"), []byte("x"), 0o644)).To(Succeed())

		res, err := GetGitMirrorsAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeEmpty())
		Expect(repo).NotTo(BeAnExistingFile())
	})

	It("removes a malformed requires_full that is a directory but keeps the valid shallow sibling", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		now := time.Now().Truncate(time.Second)
		shallow := writeMirror(filepath.Join(root, "abc", "shallow"), now)
		malformedMarker := filepath.Join(root, "abc", "requires_full")
		Expect(os.MkdirAll(malformedMarker, 0o755)).To(Succeed())

		res, err := GetGitMirrorsAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(HaveLen(1))
		Expect(res[0].GetPaths()).To(ConsistOf(shallow))
		Expect(malformedMarker).NotTo(BeAnExistingFile())
	})

	It("removes unknown children like a leftover marker tmp file", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		now := time.Now().Truncate(time.Second)
		writeMirror(filepath.Join(root, "abc", "shallow"), now)
		leftover := filepath.Join(root, "abc", "requires_full.tmp")
		Expect(os.WriteFile(leftover, nil, 0o644)).To(Succeed())

		res, err := GetGitMirrorsAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(HaveLen(1))
		Expect(leftover).NotTo(BeAnExistingFile())
	})
})
