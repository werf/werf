package gitdata

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/werf"
)

func createOrphanWorktreeFixture(kind, id, source string, accessed time.Time) string {
	dir := filepath.Join(werf.GetLocalCacheDir(), "worktrees", kind, id)
	gomega.Expect(os.MkdirAll(filepath.Join(dir, "worktree"), 0o755)).To(gomega.Succeed())
	gomega.Expect(os.WriteFile(filepath.Join(dir, "worktree", "file"), []byte("checkout"), 0o644)).To(gomega.Succeed())
	gomega.Expect(os.WriteFile(filepath.Join(dir, "last_access_at"), []byte(strconv.FormatInt(accessed.Unix(), 10)), 0o644)).To(gomega.Succeed())
	if source != "" {
		gomega.Expect(os.WriteFile(filepath.Join(dir, "git_dir"), []byte(source), 0o644)).To(gomega.Succeed())
	}
	return dir
}

func writeCacheFile(path, data string) string {
	ginkgo.GinkgoHelper()
	gomega.Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(gomega.Succeed())
	gomega.Expect(os.WriteFile(path, []byte(data), 0o644)).To(gomega.Succeed())
	return path
}

func expectGone(path string) {
	ginkgo.GinkgoHelper()
	_, err := os.Stat(path)
	gomega.Expect(os.IsNotExist(err)).To(gomega.BeTrue(), "expected %q to be gone, stat error: %v", path, err)
}

type fsSnapshotEntry struct {
	info    fs.FileInfo
	content []byte
}

func snapshotTree(root string) map[string]fsSnapshotEntry {
	res := map[string]fsSnapshotEntry{}

	gomega.Expect(filepath.WalkDir(root, func(path string, dirEntry fs.DirEntry, err error) error {
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		info, err := os.Lstat(path)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		entry := fsSnapshotEntry{info: info}
		if info.Mode().IsRegular() {
			entry.content, err = os.ReadFile(path)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}

		rel, err := filepath.Rel(root, path)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		res[rel] = entry

		return nil
	})).To(gomega.Succeed())

	return res
}

func expectTreeUnchanged(root string, before map[string]fsSnapshotEntry) {
	after := snapshotTree(root)

	for rel, beforeEntry := range before {
		afterEntry, found := after[rel]
		gomega.Expect(found).To(gomega.BeTrue(), "%q disappeared", rel)
		gomega.Expect(os.SameFile(beforeEntry.info, afterEntry.info)).To(gomega.BeTrue(), "%q was replaced", rel)
		gomega.Expect(afterEntry.info.Mode()).To(gomega.Equal(beforeEntry.info.Mode()), "mode of %q changed", rel)
		gomega.Expect(afterEntry.info.ModTime()).To(gomega.Equal(beforeEntry.info.ModTime()), "mtime of %q changed", rel)
		gomega.Expect(afterEntry.content).To(gomega.Equal(beforeEntry.content), "content of %q changed", rel)
	}

	for rel := range after {
		gomega.Expect(before).To(gomega.HaveKey(rel), "%q appeared", rel)
	}
}

func gcFixture() {
	localCache := werf.GetLocalCacheDir()
	lastAccessAt := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
	staleTime := time.Now().Add(-cacheVersionStalenessWindow - time.Hour)

	writeFile := func(path string, data []byte) {
		gomega.Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(gomega.Succeed())
		gomega.Expect(os.WriteFile(path, data, 0o644)).To(gomega.Succeed())
	}

	staleVersionDir := func(root string) {
		path := filepath.Join(localCache, root, "stale-version", "repo", "data")
		writeFile(path, []byte("stale"))
		gomega.Expect(os.Chtimes(path, staleTime, staleTime)).To(gomega.Succeed())
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
	gomega.Expect(os.MkdirAll(filepath.Join(mirrorsRoot, "empty-repo"), 0o755)).To(gomega.Succeed())
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
	writeFile(filepath.Join(patchesRoot, "repo", "ab", "orphan.patch"), []byte("orphan"))
	writeFile(filepath.Join(patchesRoot, "repo", "ab", "orphan.patch.params.archive"), []byte("filtered"))
	writeFile(filepath.Join(patchesRoot, "repo", "ab", "abcd.patch.params.paths_list"), []byte("paths"))
	writeFile(filepath.Join(patchesRoot, "repo", "ab", "abcd.patch.params.archive"), []byte("filtered"))
	writeFile(filepath.Join(archivesRoot, "repo", "ab", "orphan.tar"), []byte("orphan"))
}
