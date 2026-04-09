//go:build ai_tests

package true_git

import (
	"context"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"
)

func TestAI_IsAncestor_TrueCase(t *testing.T) {
	ctx := context.Background()
	repoDir, cleanup := createTestRepo(t)
	defer cleanup()

	commitA := createCommit(t, repoDir, "file.txt", "content A")
	commitB := createCommit(t, repoDir, "file.txt", "content B")

	ok, err := IsAncestor(ctx, commitA.String(), commitB.String(), repoDir)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestAI_IsAncestor_FalseCase(t *testing.T) {
	ctx := context.Background()
	repoDir, cleanup := createTestRepo(t)
	defer cleanup()

	commitA := createCommit(t, repoDir, "file.txt", "content A")
	_ = createCommit(t, repoDir, "file.txt", "content B")
	commitC := createCommit(t, repoDir, "file.txt", "content C")

	resetToCommit(t, repoDir, commitA)
	commitD := createCommit(t, repoDir, "file.txt", "content D")

	ok, err := IsAncestor(ctx, commitC.String(), commitD.String(), repoDir)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestAI_IsAncestor_SameCommit(t *testing.T) {
	ctx := context.Background()
	repoDir, cleanup := createTestRepo(t)
	defer cleanup()

	commit := createCommit(t, repoDir, "file.txt", "content")

	ok, err := IsAncestor(ctx, commit.String(), commit.String(), repoDir)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestAI_IsAncestor_InvalidAncestor(t *testing.T) {
	ctx := context.Background()
	repoDir, cleanup := createTestRepo(t)
	defer cleanup()

	commit := createCommit(t, repoDir, "file.txt", "content")
	invalidHash := "0000000000000000000000000000000000000000"

	ok, err := IsAncestor(ctx, invalidHash, commit.String(), repoDir)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestAI_IsAncestor_InvalidDescendant(t *testing.T) {
	ctx := context.Background()
	repoDir, cleanup := createTestRepo(t)
	defer cleanup()

	commit := createCommit(t, repoDir, "file.txt", "content")
	invalidHash := "0000000000000000000000000000000000000000"

	ok, err := IsAncestor(ctx, commit.String(), invalidHash, repoDir)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestAI_IsAncestor_LongChain(t *testing.T) {
	ctx := context.Background()
	repoDir, cleanup := createTestRepo(t)
	defer cleanup()

	commitA := createCommit(t, repoDir, "file.txt", "content A")
	commitB := createCommit(t, repoDir, "file.txt", "content B")
	commitC := createCommit(t, repoDir, "file.txt", "content C")
	commitD := createCommit(t, repoDir, "file.txt", "content D")
	commitE := createCommit(t, repoDir, "file.txt", "content E")

	ok, err := IsAncestor(ctx, commitA.String(), commitE.String(), repoDir)
	require.NoError(t, err)
	require.True(t, ok)

	for _, commit := range []plumbing.Hash{commitB, commitC, commitD, commitE} {
		ok, err := IsAncestor(ctx, commitA.String(), commit.String(), repoDir)
		require.NoError(t, err)
		require.True(t, ok)
	}
}

func createTestRepo(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()

	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	cfg, err := repo.Config()
	require.NoError(t, err)
	cfg.User.Name = "Test User"
	cfg.User.Email = "test@example.com"
	err = repo.SetConfig(cfg)
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func createCommit(t *testing.T, repoDir, filename, content string) plumbing.Hash {
	repo, err := git.PlainOpen(repoDir)
	require.NoError(t, err)

	w, err := repo.Worktree()
	require.NoError(t, err)

	f, err := w.Filesystem.Create(filename)
	require.NoError(t, err)
	_, err = f.Write([]byte(content))
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	_, err = w.Add(filename)
	require.NoError(t, err)

	hash, err := w.Commit(content, &git.CommitOptions{})
	require.NoError(t, err)

	return hash
}

func resetToCommit(t *testing.T, repoDir string, hash plumbing.Hash) {
	repo, err := git.PlainOpen(repoDir)
	require.NoError(t, err)

	w, err := repo.Worktree()
	require.NoError(t, err)

	err = w.Reset(&git.ResetOptions{
		Commit: hash,
		Mode:   git.HardReset,
	})
	require.NoError(t, err)
}
