//go:build ai_tests

package true_git

import (
	"archive/tar"
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/git_repo/repo_handle"
	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/pkg/true_git/ls_tree"
)

func TestAI_FullIntegration(t *testing.T) {
	t.Setenv("WERF_GIT_USE_WORKTREE", "1")

	ctx := context.Background()
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)

	_ = commitTestFile(t, repo, repoDir, "regular.txt", "regular-v1", "add regular")
	_ = commitTestFile(t, repo, repoDir, "remove.txt", "remove-me", "add remove")
	archiveBin := []byte{0x00, 0xaa, 0xbb}
	_ = commitTestFileBytes(t, repo, repoDir, "archive.bin", archiveBin, "add archive bin")
	archiveCommit := commitSymlink(t, repo, repoDir, "link", "regular.txt", "add link")

	var archiveOut bytes.Buffer
	err := GoGitArchive(ctx, &archiveOut, repoDir, ArchiveOptions{
		Commit:      archiveCommit.String(),
		PathScope:   ".",
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})
	require.NoError(t, err)
	archiveEntries := readTarEntries(t, archiveOut.Bytes())
	regularEntry, ok := findEntry(archiveEntries, "regular.txt")
	require.True(t, ok)
	assert.Equal(t, []byte("regular-v1"), regularEntry.content)
	archiveBinEntry, ok := findEntry(archiveEntries, "archive.bin")
	require.True(t, ok)
	assert.Equal(t, archiveBin, archiveBinEntry.content)
	linkEntry, ok := findEntry(archiveEntries, "link")
	require.True(t, ok)
	assert.Equal(t, byte(tar.TypeSymlink), linkEntry.header.Typeflag)
	assert.Equal(t, "regular.txt", linkEntry.header.Linkname)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initTestRepo(t, subDir)

	nestedDir := filepath.Join(subDir, "nested")
	nestedRepo := initTestRepo(t, nestedDir)

	nestedCommit1 := commitTestFile(t, nestedRepo, nestedDir, "nested.txt", "nested-v1", "add nested")
	_ = commitTestFile(t, subRepo, subDir, "sub.txt", "sub-v1", "add sub")
	writeGitmodules(t, subRepo, subDir, []submoduleSpec{{Name: "nested", Path: "nested", URL: "../nested"}})
	addSubmoduleEntry(t, subRepo, "nested", nestedCommit1)
	subCommit1 := commitTestRepo(t, subRepo, "add nested submodule")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", subCommit1)
	fromCommit := commitTestRepo(t, repo, "add submodule")

	_ = commitTestFile(t, repo, repoDir, "regular.txt", "regular-v2", "update regular")
	_ = commitTestRemove(t, repo, repoDir, "remove.txt", "remove file")
	rootBin := []byte{0x00, 0xff, 0x10}
	_ = commitTestFileBytes(t, repo, repoDir, "bin.dat", rootBin, "add binary")

	_ = commitTestFile(t, subRepo, subDir, "sub.txt", "sub-v2", "update sub")
	subBin := []byte{0x00, 0x01}
	_ = commitTestFileBytes(t, subRepo, subDir, "sub.bin", subBin, "add sub bin")

	nestedCommit2 := commitTestFile(t, nestedRepo, nestedDir, "nested.txt", "nested-v2", "update nested")
	setSubmoduleEntry(t, subRepo, "nested", nestedCommit2)
	subCommit2 := commitTestRepo(t, subRepo, "update nested submodule")

	setSubmoduleEntry(t, repo, "sub", subCommit2)
	toCommit := commitTestRepo(t, repo, "update submodule")

	var patchOut bytes.Buffer
	patchDesc, err := GoGitPatch(ctx, &patchOut, repoDir, repoDir, true, PatchOptions{
		FromCommit:  fromCommit.String(),
		ToCommit:    toCommit.String(),
		PathMatcher: path_matcher.NewTruePathMatcher(),
		WithBinary:  true,
	})
	require.NoError(t, err)
	assert.Contains(t, patchDesc.Paths, "regular.txt")
	assert.Contains(t, patchDesc.Paths, "bin.dat")
	assert.Contains(t, patchDesc.Paths, "remove.txt")
	assert.Contains(t, patchDesc.PathsToRemove, "remove.txt")
	assert.Contains(t, patchDesc.Paths, "sub/sub.txt")
	assert.Contains(t, patchDesc.Paths, "sub/sub.bin")
	assert.Contains(t, patchDesc.Paths, "sub/nested/nested.txt")
	assert.Contains(t, patchDesc.BinaryPaths, "bin.dat")
	assert.Contains(t, patchDesc.BinaryPaths, "sub/sub.bin")
	assert.Contains(t, patchOut.String(), "diff --git a/regular.txt b/regular.txt")
	assert.Contains(t, patchOut.String(), "diff --git a/sub/sub.txt b/sub/sub.txt")

	var subArchiveOut bytes.Buffer
	archiveMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		IncludeGlobs: []string{"sub/sub.txt", "sub/sub.bin"},
	})
	err = GoGitArchiveWithSubmodules(ctx, &subArchiveOut, repoDir, repoDir, ArchiveOptions{
		Commit:      toCommit.String(),
		PathScope:   ".",
		PathMatcher: archiveMatcher,
	})
	require.NoError(t, err)
	subArchiveEntries := readTarEntries(t, subArchiveOut.Bytes())
	entry, ok := findEntry(subArchiveEntries, "sub/sub.txt")
	require.True(t, ok)
	assert.Equal(t, []byte("sub-v2"), entry.content)
	entry, ok = findEntry(subArchiveEntries, "sub/sub.bin")
	require.True(t, ok)
	assert.Equal(t, subBin, entry.content)

	repoHandle, err := repo_handle.NewHandle(repo, repo_handle.NewHandleOptions{CommitHash: toCommit, WorkTreeDir: repoDir})
	require.NoError(t, err)
	includeMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		IncludeGlobs: []string{"sub/**"},
	})
	result, err := ls_tree.LsTree(ctx, repoHandle, toCommit.String(), ls_tree.LsTreeOptions{
		PathScope:   ".",
		PathMatcher: includeMatcher,
		AllFiles:    true,
	})
	require.NoError(t, err)
	entries := map[string]ls_tree.LsTreeEntry{}
	require.NoError(t, result.Walk(func(entry *ls_tree.LsTreeEntry) error {
		if entry.FullFilepath == "" {
			return nil
		}
		entries[filepath.ToSlash(entry.FullFilepath)] = *entry
		return nil
	}))
	_, ok = entries["sub/sub.txt"]
	assert.True(t, ok)
	_, ok = entries["sub/sub.bin"]
	assert.True(t, ok)
	_, ok = entries["sub/nested/nested.txt"]
	assert.True(t, ok)

	validation, err := ValidateSubmoduleState(ctx, repo, toCommit, repoDir)
	require.NoError(t, err)
	assert.True(t, validation.Valid)
	assert.Empty(t, validation.Errors)

	showRefResult, err := ShowRef(ctx, repoDir)
	require.NoError(t, err)
	require.NotNil(t, showRefResult)
	var hasHead bool
	var hasBranch bool
	for _, ref := range showRefResult.Refs {
		if ref.IsHEAD {
			hasHead = true
		}
		if ref.IsBranch && !ref.IsRemote {
			hasBranch = true
		}
	}
	assert.True(t, hasHead)
	assert.True(t, hasBranch)

	ancestor, err := IsAncestor(ctx, fromCommit.String(), toCommit.String(), repoDir)
	require.NoError(t, err)
	assert.True(t, ancestor)
	ancestor, err = IsAncestor(ctx, toCommit.String(), fromCommit.String(), repoDir)
	require.NoError(t, err)
	assert.False(t, ancestor)
}
