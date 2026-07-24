package gitdata

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util/timestamps"
)

func writeMirror(root, hash, kind string, ts time.Time) string {
	mirror := filepath.Join(root, hash, kind)
	Expect(os.MkdirAll(mirror, 0o755)).To(Succeed())
	Expect(timestamps.WriteTimestampFile(filepath.Join(mirror, "last_access_at"), ts)).To(Succeed())
	return mirror
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

	DescribeTable("mirror subdir combinations",
		func(ctx SpecContext, kinds []string, expected int) {
			root := GinkgoT().TempDir()
			hash := "abc"
			now := time.Now().Truncate(time.Second)

			var mirrorPaths []string
			for _, k := range kinds {
				mirrorPaths = append(mirrorPaths, writeMirror(root, hash, k, now))
			}

			res, err := GetGitReposAndRemoveInvalid(ctx, root)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(HaveLen(expected))

			paths := make(map[string]GitDataEntry, len(res))
			for _, e := range res {
				Expect(e.GetCacheBasePath()).To(Equal(root))
				Expect(e.GetLastAccessAt().Unix()).To(Equal(now.Unix()))
				Expect(e.GetPaths()).To(HaveLen(1))
				paths[e.GetPaths()[0]] = e
			}
			for _, mp := range mirrorPaths {
				Expect(paths).To(HaveKey(mp))
			}
		},
		Entry("only full", []string{"full"}, 1),
		Entry("only shallow", []string{"shallow"}, 1),
		Entry("both full and shallow", []string{"full", "shallow"}, 2),
	)

	It("removes repo dir with neither full nor shallow, even if requires_full marker is present", func(ctx SpecContext) {
		root := GinkgoT().TempDir()
		repo := filepath.Join(root, "abc")
		Expect(os.MkdirAll(repo, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, "requires_full"), nil, 0o644)).To(Succeed())

		res, err := GetGitReposAndRemoveInvalid(ctx, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeEmpty())
		Expect(repo).NotTo(BeAnExistingFile())
	})

	DescribeTable("mirror with missing or corrupt last_access_at yields a zero LastAccessAt entry so LRU can prune it",
		func(ctx SpecContext, brokenSetup func(path string)) {
			root := GinkgoT().TempDir()
			hash := "abc"
			now := time.Now().Truncate(time.Second)
			validMirror := writeMirror(root, hash, "full", now)

			brokenMirror := filepath.Join(root, hash, "shallow")
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
