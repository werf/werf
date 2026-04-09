//go:build ai_tests

package repo_handle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_NewHandle_NoSubmodules(t *testing.T) {
	repoDir := t.TempDir()
	repo := initRepo(t, repoDir)
	commitHash := commitFile(t, repo, repoDir, "README.md", "hello", "initial")

	handle := newHandleWithOptions(t, repo, repoDir, commitHash)
	assert.Empty(t, handle.Submodules())
}

func TestAI_NewHandle_SingleSubmodule(t *testing.T) {
	repoDir := t.TempDir()
	repo := initRepo(t, repoDir)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initRepo(t, subDir)
	subCommit := commitFile(t, subRepo, subDir, "sub.txt", "sub", "submodule")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", subCommit)
	parentCommit := commitRepo(t, repo, "parent")

	handle := newHandleWithOptions(t, repo, repoDir, parentCommit)
	require.Len(t, handle.Submodules(), 1)

	submoduleHandle := handle.Submodules()[0]
	assert.Equal(t, "sub", submoduleHandle.Config().Path)
	assert.Equal(t, subCommit, submoduleHandle.Status().Expected)
}

func TestAI_NewHandle_NestedSubmodules(t *testing.T) {
	repoDir := t.TempDir()
	repo := initRepo(t, repoDir)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initRepo(t, subDir)

	nestedDir := filepath.Join(subDir, "nested")
	nestedRepo := initRepo(t, nestedDir)
	nestedCommit := commitFile(t, nestedRepo, nestedDir, "nested.txt", "nested", "nested")

	writeGitmodules(t, subRepo, subDir, []submoduleSpec{{Name: "nested", Path: "nested", URL: "../nested"}})
	addSubmoduleEntry(t, subRepo, "nested", nestedCommit)
	subCommit := commitRepo(t, subRepo, "submodule")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", subCommit)
	parentCommit := commitRepo(t, repo, "parent")

	handle := newHandleWithOptions(t, repo, repoDir, parentCommit)
	require.Len(t, handle.Submodules(), 1)

	submoduleHandle := handle.Submodules()[0]
	nestedSubmodules := submoduleHandle.Submodules()
	require.Len(t, nestedSubmodules, 1)
	assert.Equal(t, nestedCommit, nestedSubmodules[0].Status().Expected)
}

func TestAI_NewHandle_BareRepo_NoOptions(t *testing.T) {
	storage := memory.NewStorage()
	repo, err := git.Init(storage, nil)
	require.NoError(t, err)

	commitHash := plumbing.ComputeHash(plumbing.CommitObject, []byte("dummy"))
	ref := plumbing.NewHashReference(plumbing.HEAD, commitHash)
	err = repo.Storer.SetReference(ref)
	require.NoError(t, err)

	handle, err := NewHandle(repo)
	require.NoError(t, err)
	assert.Empty(t, handle.Submodules())
}

type submoduleSpec struct {
	Name string
	Path string
	URL  string
}

func initRepo(t *testing.T, dir string) *git.Repository {
	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)
	return repo
}

func newHandleWithOptions(t *testing.T, repo *git.Repository, workTreeDir string, commitHash plumbing.Hash) Handle {
	handle, err := NewHandle(repo, NewHandleOptions{CommitHash: commitHash, WorkTreeDir: workTreeDir})
	require.NoError(t, err)
	return handle
}

func writeGitmodules(t *testing.T, repo *git.Repository, repoDir string, modules []submoduleSpec) {
	var builder strings.Builder
	for _, module := range modules {
		builder.WriteString(fmt.Sprintf("[submodule \"%s\"]\n\tpath = %s\n\turl = %s\n", module.Name, module.Path, module.URL))
	}

	gitmodulesPath := filepath.Join(repoDir, ".gitmodules")
	require.NoError(t, os.WriteFile(gitmodulesPath, []byte(builder.String()), 0o644))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add(".gitmodules")
	require.NoError(t, err)
}

func addSubmoduleEntry(t *testing.T, repo *git.Repository, path string, hash plumbing.Hash) {
	idx, err := repo.Storer.Index()
	require.NoError(t, err)
	idx.Entries = append(idx.Entries, &index.Entry{
		Name: filepath.ToSlash(path),
		Hash: hash,
		Mode: filemode.Submodule,
	})
	require.NoError(t, repo.Storer.SetIndex(idx))
}

func commitFile(t *testing.T, repo *git.Repository, repoDir, path, content, message string) plumbing.Hash {
	filePath := filepath.Join(repoDir, path)
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add(path)
	require.NoError(t, err)

	return commitRepo(t, repo, message)
}

func commitRepo(t *testing.T, repo *git.Repository, message string) plumbing.Hash {
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	hash, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "AI",
			Email: "ai@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
	return hash
}
