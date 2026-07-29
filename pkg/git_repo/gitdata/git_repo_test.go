package gitdata

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util/timestamps"
)

func writeMirror(mirrorPath string, ts time.Time) string {
	Expect(os.MkdirAll(mirrorPath, 0o755)).To(Succeed())
	Expect(timestamps.WriteTimestampFile(filepath.Join(mirrorPath, "last_access_at"), ts)).To(Succeed())
	return mirrorPath
}

var _ = Describe("GetGitReposAndRemoveInvalid", func() {
	It("returns nil, nil when root does not exist", func(ctx SpecContext) {
		res, err := GetGitReposAndRemoveInvalid(ctx, filepath.Join(GinkgoT().TempDir(), "missing"))
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeNil())
	})

	It("removes root when it is not a directory", func(ctx SpecContext) {
		root := filepath.Join(GinkgoT().TempDir(), "root")
		Expect(os.WriteFile(root, []byte("x"), 0o644)).To(Succeed())

		res, err := GetGitReposAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeNil())
		Expect(root).NotTo(BeAnExistingFile())
	})

	It("removes non-directory entries directly under root", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		stray := filepath.Join(root, "stray")
		Expect(os.WriteFile(stray, []byte("x"), 0o644)).To(Succeed())

		res, err := GetGitReposAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeEmpty())
		Expect(stray).NotTo(BeAnExistingFile())
	})

	It("yields one entry per flat repo dir", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		now := time.Now().Truncate(time.Second)
		mirrorA := writeMirror(filepath.Join(root, "aaa"), now)
		mirrorB := writeMirror(filepath.Join(root, "bbb"), now)

		res, err := GetGitReposAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(HaveLen(2))

		paths := make(map[string]GitDataEntry, len(res))
		for _, e := range res {
			Expect(e.GetCacheBasePath()).To(Equal(root))
			Expect(e.GetLastAccessAt().Unix()).To(Equal(now.Unix()))
			Expect(e.GetPaths()).To(HaveLen(1))
			paths[e.GetPaths()[0]] = e
		}
		Expect(paths).To(HaveKey(mirrorA))
		Expect(paths).To(HaveKey(mirrorB))
	})

	DescribeTable("repo dir with missing or corrupt last_access_at yields a zero LastAccessAt entry so LRU can prune it",
		func(ctx SpecContext, brokenSetup func(path string)) {
			root := GinkgoT().TempDir()
			now := time.Now().Truncate(time.Second)
			validMirror := writeMirror(filepath.Join(root, "valid"), now)

			brokenMirror := filepath.Join(root, "broken")
			Expect(os.MkdirAll(brokenMirror, 0o755)).To(Succeed())
			brokenSetup(filepath.Join(brokenMirror, "last_access_at"))

			res, err := GetGitReposAndRemoveInvalid(ctx, root)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(HaveLen(2))

			byPath := make(map[string]GitDataEntry, len(res))
			for _, e := range res {
				byPath[e.GetPaths()[0]] = e
			}
			Expect(byPath[validMirror].GetLastAccessAt().Unix()).To(Equal(now.Unix()))
			Expect(byPath[brokenMirror].GetLastAccessAt().IsZero()).To(BeTrue())
		},
		Entry("missing timestamp file", func(_ string) {}),
		Entry("corrupt timestamp file", func(p string) {
			Expect(os.WriteFile(p, []byte("not-a-timestamp"), 0o644)).To(Succeed())
		}),
	)
})
