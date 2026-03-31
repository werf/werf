//go:build ai_tests

package true_git

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/path_matcher"
)

type tarEntry struct {
	header  *tar.Header
	content []byte
}

func TestAI_GoGitArchive_RegularFile(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	commit := commitTestFile(t, repo, repoDir, "file.txt", "content", "add")

	var out bytes.Buffer
	err := GoGitArchive(context.Background(), &out, repoDir, ArchiveOptions{
		Commit:      commit.String(),
		PathScope:   ".",
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})
	require.NoError(t, err)

	entries := readTarEntries(t, out.Bytes())
	entry, ok := findEntry(entries, "file.txt")
	require.True(t, ok)
	assert.Equal(t, "content", string(entry.content))
	assert.Equal(t, int64(0o644), entry.header.Mode&0o777)
}

func TestAI_GoGitArchive_ExecutableFile(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	commit := commitExecutableFile(t, repo, repoDir, "exec.sh", "echo ok", "add exec")

	var out bytes.Buffer
	err := GoGitArchive(context.Background(), &out, repoDir, ArchiveOptions{
		Commit:      commit.String(),
		PathScope:   ".",
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})
	require.NoError(t, err)

	entries := readTarEntries(t, out.Bytes())
	entry, ok := findEntry(entries, "exec.sh")
	require.True(t, ok)
	assert.Equal(t, int64(0o755), entry.header.Mode&0o777)
}

func TestAI_GoGitArchive_Symlink(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	commit := commitSymlink(t, repo, repoDir, "link", "target", "add link")

	var out bytes.Buffer
	err := GoGitArchive(context.Background(), &out, repoDir, ArchiveOptions{
		Commit:      commit.String(),
		PathScope:   ".",
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})
	require.NoError(t, err)

	entries := readTarEntries(t, out.Bytes())
	entry, ok := findEntry(entries, "link")
	require.True(t, ok)
	assert.Equal(t, byte(tar.TypeSymlink), entry.header.Typeflag)
	assert.Equal(t, "target", entry.header.Linkname)
}

func TestAI_GoGitArchive_DirectoryEntries(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	commit := commitTestFile(t, repo, repoDir, "dir/file.txt", "data", "add dir")

	var out bytes.Buffer
	err := GoGitArchive(context.Background(), &out, repoDir, ArchiveOptions{
		Commit:      commit.String(),
		PathScope:   ".",
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})
	require.NoError(t, err)

	entries := readTarEntries(t, out.Bytes())
	dirIndex := entryIndex(entries, "dir")
	fileIndex := entryIndex(entries, "dir/file.txt")
	require.GreaterOrEqual(t, dirIndex, 0)
	require.GreaterOrEqual(t, fileIndex, 0)
	assert.Less(t, dirIndex, fileIndex)
	assert.Equal(t, byte(tar.TypeDir), entries[dirIndex].header.Typeflag)
}

func TestAI_GoGitArchive_PathScope(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	_ = commitTestFile(t, repo, repoDir, "dir/keep.txt", "keep", "add keep")
	commit := commitTestFile(t, repo, repoDir, "other/skip.txt", "skip", "add skip")

	var out bytes.Buffer
	err := GoGitArchive(context.Background(), &out, repoDir, ArchiveOptions{
		Commit:      commit.String(),
		PathScope:   "dir",
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})
	require.NoError(t, err)

	entries := readTarEntries(t, out.Bytes())
	_, ok := findEntry(entries, "keep.txt")
	assert.True(t, ok)
	_, ok = findEntry(entries, "other/skip.txt")
	assert.False(t, ok)
}

func TestAI_GoGitArchive_FileRenames(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	commit := commitTestFile(t, repo, repoDir, "old/name.txt", "data", "add")

	var out bytes.Buffer
	err := GoGitArchive(context.Background(), &out, repoDir, ArchiveOptions{
		Commit:      commit.String(),
		PathScope:   "old",
		PathMatcher: path_matcher.NewTruePathMatcher(),
		FileRenames: map[string]string{"old/name.txt": "renamed.txt"},
	})
	require.NoError(t, err)

	entries := readTarEntries(t, out.Bytes())
	_, ok := findEntry(entries, "renamed.txt")
	assert.True(t, ok)
	_, ok = findEntry(entries, "name.txt")
	assert.False(t, ok)
}

func TestAI_GoGitArchive_Ownership(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	commit := commitTestFile(t, repo, repoDir, "file.txt", "data", "add")

	var out bytes.Buffer
	err := GoGitArchive(context.Background(), &out, repoDir, ArchiveOptions{
		Commit:      commit.String(),
		PathScope:   ".",
		PathMatcher: path_matcher.NewTruePathMatcher(),
		Owner:       "1234",
		Group:       "staff",
	})
	require.NoError(t, err)

	entries := readTarEntries(t, out.Bytes())
	entry, ok := findEntry(entries, "file.txt")
	require.True(t, ok)
	assert.Equal(t, 1234, entry.header.Uid)
	assert.Equal(t, "", entry.header.Uname)
	assert.Equal(t, "staff", entry.header.Gname)
}

func TestAI_GoGitArchive_WithSubmodules(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initTestRepo(t, subDir)
	subCommit := commitTestFile(t, subRepo, subDir, "sub.txt", "sub", "submodule")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", subCommit)
	commit := commitTestRepo(t, repo, "add submodule")

	var out bytes.Buffer
	err := GoGitArchiveWithSubmodules(context.Background(), &out, repoDir, repoDir, ArchiveOptions{
		Commit:      commit.String(),
		PathScope:   ".",
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})
	require.NoError(t, err)

	entries := readTarEntries(t, out.Bytes())
	_, ok := findEntry(entries, "sub/sub.txt")
	assert.True(t, ok)
}

func readTarEntries(t *testing.T, data []byte) []tarEntry {
	t.Helper()

	tr := tar.NewReader(bytes.NewReader(data))
	var entries []tarEntry
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		var content []byte
		if hdr.Size > 0 {
			content, err = io.ReadAll(tr)
			require.NoError(t, err)
		}

		headerCopy := *hdr
		entries = append(entries, tarEntry{header: &headerCopy, content: content})
	}

	return entries
}

func findEntry(entries []tarEntry, name string) (*tarEntry, bool) {
	for i := range entries {
		if entries[i].header.Name == name {
			return &entries[i], true
		}
	}

	return nil, false
}

func entryIndex(entries []tarEntry, name string) int {
	for i, entry := range entries {
		if entry.header.Name == name {
			return i
		}
	}

	return -1
}

func commitExecutableFile(t *testing.T, repo *git.Repository, repoDir, path, content, message string) plumbing.Hash {
	filePath := filepath.Join(repoDir, path)
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o755))
	require.NoError(t, os.Chmod(filePath, 0o755))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add(path)
	require.NoError(t, err)

	idx, err := repo.Storer.Index()
	require.NoError(t, err)

	path = filepath.ToSlash(path)
	updated := false
	for _, entry := range idx.Entries {
		if entry.Name != path {
			continue
		}
		entry.Mode = filemode.Executable
		updated = true
		break
	}
	require.True(t, updated)
	require.NoError(t, repo.Storer.SetIndex(idx))

	return commitTestRepo(t, repo, message)
}

func commitSymlink(t *testing.T, repo *git.Repository, repoDir, path, target, message string) plumbing.Hash {
	filePath := filepath.Join(repoDir, path)
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.Symlink(target, filePath))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add(path)
	require.NoError(t, err)

	return commitTestRepo(t, repo, message)
}
