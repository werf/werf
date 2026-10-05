package gitdata

import (
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util/timestamps"
	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/werf"
)

type fsSnapshotEntry struct {
	info    fs.FileInfo
	content []byte
}

func snapshotTree(root string) map[string]fsSnapshotEntry {
	res := map[string]fsSnapshotEntry{}

	Expect(filepath.WalkDir(root, func(path string, dirEntry fs.DirEntry, err error) error {
		Expect(err).NotTo(HaveOccurred())

		info, err := os.Lstat(path)
		Expect(err).NotTo(HaveOccurred())

		entry := fsSnapshotEntry{info: info}
		if info.Mode().IsRegular() {
			entry.content, err = os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
		}

		rel, err := filepath.Rel(root, path)
		Expect(err).NotTo(HaveOccurred())
		res[rel] = entry

		return nil
	})).To(Succeed())

	return res
}

func expectTreeUnchanged(root string, before map[string]fsSnapshotEntry) {
	after := snapshotTree(root)

	for rel, beforeEntry := range before {
		afterEntry, found := after[rel]
		Expect(found).To(BeTrue(), "%q disappeared", rel)
		Expect(os.SameFile(beforeEntry.info, afterEntry.info)).To(BeTrue(), "%q was replaced", rel)
		Expect(afterEntry.info.Mode()).To(Equal(beforeEntry.info.Mode()), "mode of %q changed", rel)
		Expect(afterEntry.info.ModTime()).To(Equal(beforeEntry.info.ModTime()), "mtime of %q changed", rel)
		Expect(afterEntry.content).To(Equal(beforeEntry.content), "content of %q changed", rel)
	}

	for rel := range after {
		Expect(before).To(HaveKey(rel), "%q appeared", rel)
	}
}

// gcFixture fills the local cache with one valid LRU entry per cache root plus
// the invalid shapes every collector reclaims, and a stale foreign version dir
// per root for the wipe path.
func gcFixture() {
	localCache := werf.GetLocalCacheDir()
	lastAccessAt := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
	staleTime := time.Now().Add(-cacheVersionStalenessWindow - time.Hour)

	writeFile := func(path string, data []byte) {
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, data, 0o644)).To(Succeed())
	}

	staleVersionDir := func(root string) {
		path := filepath.Join(localCache, root, "stale-version", "repo", "data")
		writeFile(path, []byte("stale"))
		Expect(os.Chtimes(path, staleTime, staleTime)).To(Succeed())
	}

	for _, root := range []string{"git_repos", "git_mirrors", "git_worktrees", "git_archives", "git_patches"} {
		staleVersionDir(root)
	}

	reposRoot := filepath.Join(localCache, "git_repos", git_repo.GitReposCacheVersion)
	writeMirror(filepath.Join(reposRoot, "valid"), lastAccessAt)
	writeFile(filepath.Join(reposRoot, "valid", "objects", "pack", "pack-1.pack"), []byte("repo payload"))
	writeFile(filepath.Join(reposRoot, "stray-file"), []byte("not a dir"))

	mirrorsRoot := filepath.Join(localCache, "git_mirrors", git_repo.GitMirrorsCacheVersion)
	writeMirror(filepath.Join(mirrorsRoot, "valid", "shallow"), lastAccessAt)
	writeFile(filepath.Join(mirrorsRoot, "valid", "shallow", "objects", "pack", "pack-1.pack"), []byte("shallow payload"))
	writeFile(filepath.Join(mirrorsRoot, "valid", "shallow.abc.tmp", "data"), []byte("in-flight clone"))
	writeFile(filepath.Join(mirrorsRoot, "pinned", "requires_full"), nil)
	Expect(os.MkdirAll(filepath.Join(mirrorsRoot, "empty-repo"), 0o755)).To(Succeed())
	writeMirror(filepath.Join(mirrorsRoot, "corrupt-ts", "shallow"), lastAccessAt)
	writeFile(filepath.Join(mirrorsRoot, "corrupt-ts", "shallow", "last_access_at"), []byte("not-a-timestamp"))

	worktreesRoot := filepath.Join(localCache, "git_worktrees", git_repo.GitWorktreesCacheVersion)
	for _, kind := range []string{"local", "remote"} {
		writeMirror(filepath.Join(worktreesRoot, kind, "valid"), lastAccessAt)
		writeFile(filepath.Join(worktreesRoot, kind, "valid", "worktree", "file.txt"), []byte("worktree payload"))
		writeFile(filepath.Join(worktreesRoot, kind, "stray-file"), []byte("not a dir"))
	}

	archivesRoot := filepath.Join(localCache, "git_archives", GitArchivesCacheVersion)
	writeFile(filepath.Join(archivesRoot, "repo", "ab", "abcd.tar"), []byte("archive payload"))
	writeFile(filepath.Join(archivesRoot, "repo", "ab", "abcd.meta.json"), []byte(`{"LastAccessTimestamp":1}`))
	writeFile(filepath.Join(archivesRoot, "repo", "ab", "dangling.meta.json"), []byte(`{"LastAccessTimestamp":1}`))
	writeFile(filepath.Join(archivesRoot, "repo", "ab", "corrupt.meta.json"), []byte("}not json{"))
	writeFile(filepath.Join(archivesRoot, "repo", "ab", "unrecognized.txt"), []byte("junk"))
	writeFile(filepath.Join(archivesRoot, "stray-file"), []byte("not a dir"))

	patchesRoot := filepath.Join(localCache, "git_patches", GitPatchesCacheVersion)
	writeFile(filepath.Join(patchesRoot, "repo", "ab", "abcd.patch"), []byte("patch payload"))
	writeFile(filepath.Join(patchesRoot, "repo", "ab", "abcd.patch.ff.paths_list"), []byte("a.txt\n"))
	writeFile(filepath.Join(patchesRoot, "repo", "ab", "abcd.meta.json"), []byte(`{"LastAccessTimestamp":1}`))
	writeFile(filepath.Join(patchesRoot, "repo", "ab", "dangling.meta.json"), []byte(`{"LastAccessTimestamp":1}`))
	writeFile(filepath.Join(patchesRoot, "repo", "ab", "corrupt.meta.json"), []byte("}not json{"))
	writeFile(filepath.Join(patchesRoot, "repo", "ab", "unrecognized.txt"), []byte("junk"))
	writeFile(filepath.Join(patchesRoot, "stray-file"), []byte("not a dir"))
}

var _ = Describe("RunGC dry run", func() {
	var localCache string

	BeforeEach(func() {
		GinkgoT().Setenv("WERF_HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("WERF_TMP_DIR", GinkgoT().TempDir())
		Expect(werf.Init("", "")).To(Succeed())

		localCache = werf.GetLocalCacheDir()
		gcFixture()
	})

	It("touches nothing under maximum volume pressure", func(ctx SpecContext) {
		before := snapshotTree(localCache)

		Expect(RunGC(ctx, RunGCOptions{DryRun: true})).To(Succeed())

		expectTreeUnchanged(localCache, before)
	})

	It("touches nothing when the volume usage is below the allowed level", func(ctx SpecContext) {
		before := snapshotTree(localCache)

		Expect(RunGC(ctx, RunGCOptions{
			AllowedLocalCacheVolumeUsageBytes: math.MaxUint64,
			DryRun:                            true,
		})).To(Succeed())

		expectTreeUnchanged(localCache, before)
	})

	It("removes the same fixture for real when dry run is off", func(ctx SpecContext) {
		staleVersionDir := filepath.Join(localCache, "git_repos", "stale-version")
		strayFile := filepath.Join(localCache, "git_repos", git_repo.GitReposCacheVersion, "stray-file")
		danglingMeta := filepath.Join(localCache, "git_patches", GitPatchesCacheVersion, "repo", "ab", "dangling.meta.json")
		unrecognized := filepath.Join(localCache, "git_archives", GitArchivesCacheVersion, "repo", "ab", "unrecognized.txt")
		validRepo := filepath.Join(localCache, "git_repos", git_repo.GitReposCacheVersion, "valid")

		Expect(RunGC(ctx, RunGCOptions{})).To(Succeed())

		Expect(staleVersionDir).NotTo(BeADirectory())
		Expect(strayFile).NotTo(BeAnExistingFile())
		Expect(danglingMeta).NotTo(BeAnExistingFile())
		Expect(unrecognized).NotTo(BeAnExistingFile())
		Expect(validRepo).NotTo(BeADirectory())
	})

	It("does not refresh last access timestamps", func(ctx SpecContext) {
		lastAccessAtPath := filepath.Join(localCache, "git_repos", git_repo.GitReposCacheVersion, "valid", "last_access_at")

		before, err := timestamps.ReadTimestampFile(lastAccessAtPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(RunGC(ctx, RunGCOptions{DryRun: true})).To(Succeed())

		after, err := timestamps.ReadTimestampFile(lastAccessAtPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
	})
})
