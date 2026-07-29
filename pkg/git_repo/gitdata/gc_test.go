package gitdata

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util/timestamps"
)

var _ = Describe("getForeignCacheVersionEntries", func() {
	It("returns nil, nil when cache root does not exist", func() {
		res, err := getForeignCacheVersionEntries(filepath.Join(GinkgoT().TempDir(), "missing"), []string{"6"})
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeNil())
	})

	It("skips kept versions and non-directories", func() {
		root := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(root, "6"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "3"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "5"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "stray"), []byte("x"), 0o644)).To(Succeed())

		res, err := getForeignCacheVersionEntries(root, []string{"6", "3"})
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(HaveLen(1))
		Expect(res[0].GetPaths()).To(Equal([]string{filepath.Join(root, "5")}))
		Expect(res[0].GetCacheBasePath()).To(Equal(root))
	})

	It("uses the newest last_access_at found in old flat layout", func() {
		root := GinkgoT().TempDir()
		older := time.Now().Add(-48 * time.Hour).Truncate(time.Second)
		newer := time.Now().Add(-1 * time.Hour).Truncate(time.Second)

		for i, ts := range []time.Time{older, newer} {
			repoDir := filepath.Join(root, "5", fmt.Sprintf("repo%d", i))
			Expect(os.MkdirAll(repoDir, 0o755)).To(Succeed())
			Expect(timestamps.WriteTimestampFile(filepath.Join(repoDir, "last_access_at"), ts)).To(Succeed())
		}

		res, err := getForeignCacheVersionEntries(root, []string{"6"})
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(HaveLen(1))
		Expect(res[0].GetLastAccessAt().Unix()).To(Equal(newer.Unix()))
	})

	It("uses the newest last_access_at found in mirror layout", func() {
		root := GinkgoT().TempDir()
		ts := time.Now().Add(-2 * time.Hour).Truncate(time.Second)

		mirrorDir := filepath.Join(root, "7", "repohash", "full")
		Expect(os.MkdirAll(mirrorDir, 0o755)).To(Succeed())
		Expect(timestamps.WriteTimestampFile(filepath.Join(mirrorDir, "last_access_at"), ts)).To(Succeed())

		res, err := getForeignCacheVersionEntries(root, []string{"6"})
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(HaveLen(1))
		Expect(res[0].GetLastAccessAt().Unix()).To(Equal(ts.Unix()))
	})

	It("falls back to dir mtime when no last_access_at is found", func() {
		root := GinkgoT().TempDir()
		versionDir := filepath.Join(root, "5")
		Expect(os.MkdirAll(versionDir, 0o755)).To(Succeed())

		res, err := getForeignCacheVersionEntries(root, []string{"6"})
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(HaveLen(1))
		Expect(res[0].GetLastAccessAt()).To(BeTemporally("~", time.Now(), time.Minute))
	})
})

var _ = Describe("RemovePathWithEmptyParentDirsInsideScope", func() {
	It("removes target and cleans up empty parents up to scope", func() {
		scope := GinkgoT().TempDir()
		target := filepath.Join(scope, "a", "b", "c")
		Expect(os.MkdirAll(target, 0o755)).To(Succeed())

		Expect(RemovePathWithEmptyParentDirsInsideScope(scope, target)).To(Succeed())
		Expect(filepath.Join(scope, "a")).NotTo(BeADirectory())
		Expect(scope).To(BeADirectory())
	})

	It("keeps non-empty parents", func() {
		scope := GinkgoT().TempDir()
		target := filepath.Join(scope, "a", "b")
		Expect(os.MkdirAll(target, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(scope, "a", "keep"), []byte("x"), 0o644)).To(Succeed())

		Expect(RemovePathWithEmptyParentDirsInsideScope(scope, target)).To(Succeed())
		Expect(target).NotTo(BeADirectory())
		Expect(filepath.Join(scope, "a", "keep")).To(BeAnExistingFile())
	})

	It("does nothing for paths outside the scope", func() {
		scope := GinkgoT().TempDir()
		outside := GinkgoT().TempDir()
		target := filepath.Join(outside, "dir")
		Expect(os.MkdirAll(target, 0o755)).To(Succeed())

		Expect(RemovePathWithEmptyParentDirsInsideScope(scope, target)).To(Succeed())
		Expect(target).To(BeADirectory())
	})

	It("succeeds when target does not exist", func() {
		scope := GinkgoT().TempDir()
		Expect(RemovePathWithEmptyParentDirsInsideScope(scope, filepath.Join(scope, "missing"))).To(Succeed())
	})
})
