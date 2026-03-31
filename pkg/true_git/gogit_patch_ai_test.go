//go:build ai_tests

package true_git

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/path_matcher"
)

func TestAI_GoGitPatch_EmptyDiff(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	commit := commitTestFile(t, repo, repoDir, "file.txt", "content", "initial")

	desc, out := runGoGitPatch(t, repoDir, PatchOptions{
		FromCommit:  commit.String(),
		ToCommit:    commit.String(),
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})

	assert.Empty(t, desc.Paths)
	assert.Empty(t, desc.BinaryPaths)
	assert.Empty(t, desc.PathsToRemove)
	assert.Empty(t, out)
}

func TestAI_GoGitPatch_FileAdd(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	from := commitTestFile(t, repo, repoDir, "base.txt", "base", "base")
	to := commitTestFile(t, repo, repoDir, "new.txt", "new", "add")

	desc, out := runGoGitPatch(t, repoDir, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: path_matcher.NewTruePathMatcher(),
		WithBinary:  true,
	})

	assert.Contains(t, desc.Paths, "new.txt")
	assert.NotContains(t, desc.PathsToRemove, "new.txt")
	assert.NotContains(t, desc.BinaryPaths, "new.txt")
	assert.Contains(t, out, "diff --git a/new.txt b/new.txt")
}

func TestAI_GoGitPatch_FileDelete(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	from := commitTestFile(t, repo, repoDir, "delete.txt", "old", "add")
	to := commitTestRemove(t, repo, repoDir, "delete.txt", "remove")

	desc, out := runGoGitPatch(t, repoDir, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: path_matcher.NewTruePathMatcher(),
		WithBinary:  true,
	})

	assert.Contains(t, desc.Paths, "delete.txt")
	assert.Contains(t, desc.PathsToRemove, "delete.txt")
	assert.NotContains(t, desc.BinaryPaths, "delete.txt")
	assert.Contains(t, out, "diff --git a/delete.txt b/delete.txt")
}

func TestAI_GoGitPatch_FileModify(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	from := commitTestFile(t, repo, repoDir, "modify.txt", "old", "add")
	to := commitTestFile(t, repo, repoDir, "modify.txt", "new", "modify")

	desc, out := runGoGitPatch(t, repoDir, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})

	assert.Contains(t, desc.Paths, "modify.txt")
	assert.Empty(t, desc.PathsToRemove)
	assert.Contains(t, out, "@@")
}

func TestAI_GoGitPatch_BinaryFileDetection(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	from := commitTestFile(t, repo, repoDir, "text.txt", "text", "add")
	to := commitTestFileBytes(t, repo, repoDir, "bin.dat", []byte{0x00, 0xff, 0x10}, "binary")

	desc, _ := runGoGitPatch(t, repoDir, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: path_matcher.NewTruePathMatcher(),
		WithBinary:  true,
	})

	assert.Contains(t, desc.BinaryPaths, "bin.dat")
}

func TestAI_GoGitPatch_PathScopeFiltering(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	from := commitTestFile(t, repo, repoDir, "dir/keep.txt", "old", "add")
	_ = commitTestFile(t, repo, repoDir, "other/skip.txt", "old", "add2")
	to := commitTestFile(t, repo, repoDir, "dir/keep.txt", "new", "modify")

	desc, _ := runGoGitPatch(t, repoDir, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathScope:   "dir",
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})

	assert.Contains(t, desc.Paths, "keep.txt")
	assert.NotContains(t, desc.Paths, "other/skip.txt")
}

func TestAI_GoGitPatch_PathMatcherFiltering(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	from := commitTestFile(t, repo, repoDir, "a.txt", "old", "add")
	_ = commitTestFile(t, repo, repoDir, "b.txt", "old", "add2")
	to := commitTestFile(t, repo, repoDir, "a.txt", "new", "modify")

	matcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		IncludeGlobs: []string{"a.txt"},
	})

	desc, _ := runGoGitPatch(t, repoDir, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: matcher,
	})

	assert.Equal(t, []string{"a.txt"}, desc.Paths)
}

func TestAI_GoGitPatch_FileRenames(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	from := commitTestFile(t, repo, repoDir, "old/name.txt", "old", "add")
	to := commitTestFile(t, repo, repoDir, "old/name.txt", "new", "modify")

	fileRenames := map[string]string{"old/name.txt": "newname.txt"}

	desc, _ := runGoGitPatch(t, repoDir, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathScope:   "old",
		PathMatcher: path_matcher.NewTruePathMatcher(),
		FileRenames: fileRenames,
	})

	assert.Contains(t, desc.Paths, "newname.txt")
}

func TestAI_GoGitPatch_EntireFileContext(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	from := commitTestFile(t, repo, repoDir, "full.txt", "line1\nline2\nline3\n", "add")
	to := commitTestFile(t, repo, repoDir, "full.txt", "line1\nline2\nline3\nline4\n", "modify")

	_, out := runGoGitPatch(t, repoDir, PatchOptions{
		FromCommit:            from.String(),
		ToCommit:              to.String(),
		PathMatcher:           path_matcher.NewTruePathMatcher(),
		WithEntireFileContext: true,
	})

	assert.Contains(t, out, " line1")
	assert.Contains(t, out, " line2")
	assert.Contains(t, out, " line3")
}

func TestAI_GoGitPatch_DiffParserCompatibility(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)
	from := commitTestFile(t, repo, repoDir, "compat.txt", "old", "add")
	to := commitTestFile(t, repo, repoDir, "compat.txt", "new", "modify")

	_, out := runGoGitPatch(t, repoDir, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: path_matcher.NewTruePathMatcher(),
	})

	parser := makeDiffParser(&bytes.Buffer{}, "", path_matcher.NewTruePathMatcher(), nil)
	require.NoError(t, parser.HandleStdout([]byte(out)))
	assert.Contains(t, parser.Paths, "compat.txt")
}

func TestAI_GoGitPatch_SubmoduleFileChange(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initTestRepo(t, subDir)

	_ = commitTestFile(t, subRepo, subDir, "modify.txt", "old", "add modify")
	fromSub := commitTestFile(t, subRepo, subDir, "remove.txt", "old", "add remove")

	_ = commitTestFile(t, subRepo, subDir, "modify.txt", "new", "modify")
	_ = commitTestRemove(t, subRepo, subDir, "remove.txt", "remove")
	toSub := commitTestFile(t, subRepo, subDir, "new.txt", "new", "add new")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", fromSub)
	from := commitTestRepo(t, repo, "add submodule")

	setSubmoduleEntry(t, repo, "sub", toSub)
	to := commitTestRepo(t, repo, "update submodule")

	matcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: []string{"sub/**"}})
	desc, out := runGoGitPatchWithSubmodules(t, repoDir, true, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: matcher,
	})

	assert.Contains(t, desc.Paths, "sub/modify.txt")
	assert.Contains(t, desc.Paths, "sub/new.txt")
	assert.Contains(t, desc.Paths, "sub/remove.txt")
	assert.Contains(t, desc.PathsToRemove, "sub/remove.txt")
	assert.Contains(t, out, "diff --git a/sub/modify.txt b/sub/modify.txt")
	assert.NotContains(t, out, "diff --git a/sub b/sub")
}

func TestAI_GoGitPatch_NestedSubmodule(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initTestRepo(t, subDir)

	nestedDir := filepath.Join(subDir, "nested")
	nestedRepo := initTestRepo(t, nestedDir)

	fromNested := commitTestFile(t, nestedRepo, nestedDir, "nested.txt", "old", "add nested")

	writeGitmodules(t, subRepo, subDir, []submoduleSpec{{Name: "nested", Path: "nested", URL: "../nested"}})
	addSubmoduleEntry(t, subRepo, "nested", fromNested)
	fromSub := commitTestRepo(t, subRepo, "add nested submodule")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", fromSub)
	from := commitTestRepo(t, repo, "add submodule")

	toNested := commitTestFile(t, nestedRepo, nestedDir, "nested.txt", "new", "modify nested")
	setSubmoduleEntry(t, subRepo, "nested", toNested)
	toSub := commitTestRepo(t, subRepo, "update nested")

	setSubmoduleEntry(t, repo, "sub", toSub)
	to := commitTestRepo(t, repo, "update submodule")

	matcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: []string{"sub/**"}})
	desc, out := runGoGitPatchWithSubmodules(t, repoDir, true, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: matcher,
	})

	assert.Contains(t, desc.Paths, "sub/nested/nested.txt")
	assert.Contains(t, out, "diff --git a/sub/nested/nested.txt b/sub/nested/nested.txt")
}

func TestAI_GoGitPatch_SubmoduleAdded(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)

	from := commitTestRepo(t, repo, "base")

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initTestRepo(t, subDir)
	subCommit := commitTestFile(t, subRepo, subDir, "sub.txt", "content", "submodule")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", subCommit)
	to := commitTestRepo(t, repo, "add submodule")

	matcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: []string{"sub/**"}})
	desc, out := runGoGitPatchWithSubmodules(t, repoDir, true, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: matcher,
	})

	assert.Contains(t, desc.Paths, "sub/sub.txt")
	assert.Contains(t, out, "diff --git a/sub/sub.txt b/sub/sub.txt")
}

func TestAI_GoGitPatch_SubmoduleRemoved(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initTestRepo(t, subDir)
	subCommit := commitTestFile(t, subRepo, subDir, "sub.txt", "content", "submodule")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", subCommit)
	from := commitTestRepo(t, repo, "add submodule")

	removeSubmoduleEntry(t, repo, "sub")
	removeFileFromRepo(t, repo, repoDir, ".gitmodules")
	to := commitTestRepo(t, repo, "remove submodule")

	matcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: []string{"sub/**"}})
	desc, out := runGoGitPatchWithSubmodules(t, repoDir, true, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: matcher,
	})

	assert.Contains(t, desc.Paths, "sub/sub.txt")
	assert.Contains(t, desc.PathsToRemove, "sub/sub.txt")
	assert.Contains(t, out, "diff --git a/sub/sub.txt b/sub/sub.txt")
}

func TestAI_GoGitPatch_SubmoduleBinaryFile(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initTestRepo(t, subDir)

	fromSub := commitTestFile(t, subRepo, subDir, "text.txt", "text", "add text")
	toSub := commitTestFileBytes(t, subRepo, subDir, "bin.dat", []byte{0x00, 0xff, 0x10}, "add binary")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", fromSub)
	from := commitTestRepo(t, repo, "add submodule")

	setSubmoduleEntry(t, repo, "sub", toSub)
	to := commitTestRepo(t, repo, "update submodule")

	matcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: []string{"sub/**"}})
	desc, _ := runGoGitPatchWithSubmodules(t, repoDir, true, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: matcher,
		WithBinary:  true,
	})

	assert.Contains(t, desc.BinaryPaths, "sub/bin.dat")
}

func TestAI_GoGitPatch_WithoutSubmodules(t *testing.T) {
	repoDir := t.TempDir()
	repo := initTestRepo(t, repoDir)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initTestRepo(t, subDir)

	fromSub := commitTestFile(t, subRepo, subDir, "sub.txt", "old", "submodule")
	toSub := commitTestFile(t, subRepo, subDir, "sub.txt", "new", "submodule update")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", fromSub)
	from := commitTestRepo(t, repo, "add submodule")

	setSubmoduleEntry(t, repo, "sub", toSub)
	to := commitTestRepo(t, repo, "update submodule")

	matcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: []string{"sub"}})
	desc, out := runGoGitPatchWithSubmodules(t, repoDir, false, PatchOptions{
		FromCommit:  from.String(),
		ToCommit:    to.String(),
		PathMatcher: matcher,
	})

	assert.Empty(t, desc.Paths)
	assert.Empty(t, desc.PathsToRemove)
	assert.Empty(t, desc.BinaryPaths)
	assert.Empty(t, out)
}

func runGoGitPatch(t *testing.T, repoDir string, opts PatchOptions) (*PatchDescriptor, string) {
	if opts.PathMatcher == nil {
		opts.PathMatcher = path_matcher.NewTruePathMatcher()
	}

	var out bytes.Buffer
	desc, err := GoGitPatch(context.Background(), &out, repoDir, "", false, opts)
	require.NoError(t, err)

	return desc, out.String()
}

func runGoGitPatchWithSubmodules(t *testing.T, repoDir string, withSubmodules bool, opts PatchOptions) (*PatchDescriptor, string) {
	if opts.PathMatcher == nil {
		opts.PathMatcher = path_matcher.NewTruePathMatcher()
	}

	workTreeDir := ""
	if withSubmodules {
		workTreeDir = repoDir
	}

	var out bytes.Buffer
	desc, err := GoGitPatch(context.Background(), &out, repoDir, workTreeDir, withSubmodules, opts)
	require.NoError(t, err)

	return desc, out.String()
}

func initTestRepo(t *testing.T, dir string) *git.Repository {
	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)

	return repo
}

func commitTestFile(t *testing.T, repo *git.Repository, repoDir, path, content, message string) plumbing.Hash {
	return commitTestFileBytes(t, repo, repoDir, path, []byte(content), message)
}

func commitTestFileBytes(t *testing.T, repo *git.Repository, repoDir, path string, content []byte, message string) plumbing.Hash {
	filePath := filepath.Join(repoDir, path)
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, content, 0o644))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add(path)
	require.NoError(t, err)

	return commitTestRepo(t, repo, message)
}

func commitTestRemove(t *testing.T, repo *git.Repository, repoDir, path, message string) plumbing.Hash {
	filePath := filepath.Join(repoDir, path)
	require.NoError(t, os.Remove(filePath))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Remove(path)
	require.NoError(t, err)

	return commitTestRepo(t, repo, message)
}

func commitTestRepo(t *testing.T, repo *git.Repository, message string) plumbing.Hash {
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	hash, err := worktree.Commit(message, &git.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "AI",
			Email: "ai@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	return hash
}

func setSubmoduleEntry(t *testing.T, repo *git.Repository, path string, hash plumbing.Hash) {
	idx, err := repo.Storer.Index()
	require.NoError(t, err)

	path = filepath.ToSlash(path)
	for _, entry := range idx.Entries {
		if entry.Name != path {
			continue
		}
		entry.Hash = hash
		entry.Mode = filemode.Submodule
		require.NoError(t, repo.Storer.SetIndex(idx))
		return
	}

	addSubmoduleEntry(t, repo, path, hash)
}

func removeSubmoduleEntry(t *testing.T, repo *git.Repository, path string) {
	idx, err := repo.Storer.Index()
	require.NoError(t, err)

	path = filepath.ToSlash(path)
	entries := idx.Entries[:0]
	for _, entry := range idx.Entries {
		if entry.Name == path {
			continue
		}
		entries = append(entries, entry)
	}
	idx.Entries = entries
	require.NoError(t, repo.Storer.SetIndex(idx))
}

func removeFileFromRepo(t *testing.T, repo *git.Repository, repoDir, path string) {
	filePath := filepath.Join(repoDir, path)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		require.NoError(t, err)
	}

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		require.NoError(t, err)
	}
}
