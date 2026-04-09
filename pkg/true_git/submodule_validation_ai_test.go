//go:build ai_tests

package true_git

import (
	"context"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_ValidateSubmodule_NoSubmodules(t *testing.T) {
	ctx := context.Background()
	repoDir := t.TempDir()

	repo := initRepo(t, repoDir)
	commitHash := commitFile(t, repo, repoDir, "README.md", "hello", "initial")

	result, err := ValidateSubmoduleState(ctx, repo, commitHash, repoDir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestAI_ValidateSubmodule_ValidSubmodule(t *testing.T) {
	ctx := context.Background()
	repoDir := t.TempDir()
	repo := initRepo(t, repoDir)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initRepo(t, subDir)
	subCommit := commitFile(t, subRepo, subDir, "sub.txt", "sub", "submodule")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", subCommit)
	parentCommit := commitRepo(t, repo, "parent")

	result, err := ValidateSubmoduleState(ctx, repo, parentCommit, repoDir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestAI_ValidateSubmodule_Uninitialized(t *testing.T) {
	ctx := context.Background()
	repoDir := t.TempDir()
	repo := initRepo(t, repoDir)

	subRepoDir := t.TempDir()
	subRepo := initRepo(t, subRepoDir)
	subCommit := commitFile(t, subRepo, subRepoDir, "sub.txt", "sub", "submodule")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: subRepoDir}})
	addSubmoduleEntry(t, repo, "sub", subCommit)
	parentCommit := commitRepo(t, repo, "parent")

	result, err := ValidateSubmoduleState(ctx, repo, parentCommit, repoDir)
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.Len(t, result.Errors, 1)
	errDetails := result.Errors[0]
	assert.Equal(t, "sub", errDetails.SubmodulePath)
	assert.False(t, errDetails.Initialized)
	assert.Contains(t, errDetails.Message, "git submodule update --init")
}

func TestAI_ValidateSubmodule_WrongCommit(t *testing.T) {
	ctx := context.Background()
	repoDir := t.TempDir()
	repo := initRepo(t, repoDir)

	subDir := filepath.Join(repoDir, "sub")
	subRepo := initRepo(t, subDir)
	expectedCommit := commitFile(t, subRepo, subDir, "sub.txt", "sub", "submodule")
	actualCommit := commitFile(t, subRepo, subDir, "sub.txt", "sub-updated", "submodule update")

	writeGitmodules(t, repo, repoDir, []submoduleSpec{{Name: "sub", Path: "sub", URL: "../sub"}})
	addSubmoduleEntry(t, repo, "sub", expectedCommit)
	parentCommit := commitRepo(t, repo, "parent")

	result, err := ValidateSubmoduleState(ctx, repo, parentCommit, repoDir)
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.Len(t, result.Errors, 1)
	errDetails := result.Errors[0]
	assert.True(t, errDetails.Initialized)
	assert.Equal(t, expectedCommit.String(), errDetails.ExpectedCommit)
	assert.Equal(t, actualCommit.String(), errDetails.ActualCommit)
	assert.Contains(t, errDetails.Message, "expected commit")
}

func TestAI_ValidateSubmodule_NestedSubmodules(t *testing.T) {
	ctx := context.Background()
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

	result, err := ValidateSubmoduleState(ctx, repo, parentCommit, repoDir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
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
