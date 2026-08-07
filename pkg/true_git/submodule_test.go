package true_git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetIncludePathOptions(t *testing.T) {
	ctx := context.Background()

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	gitDir := filepath.Join(repoDir, ".git")

	opts, err := getIncludePathOptions(ctx, gitDir)
	require.NoError(t, err)
	require.Empty(t, opts)

	runGit(t, repoDir, "config", "--add", "include.path", "/abs/ext.conf")
	runGit(t, repoDir, "config", "--add", "include.path", "rel/ext.conf")
	runGit(t, repoDir, "config", "--add", "include.path", "~/tilde-ext.conf")

	opts, err = getIncludePathOptions(ctx, gitDir)
	require.NoError(t, err)
	require.Len(t, opts, 6)
	require.Equal(t, "-c", opts[0])
	require.Equal(t, "include.path=/abs/ext.conf", opts[1])
	require.Equal(t, "-c", opts[2])
	require.True(t, strings.HasPrefix(opts[3], "include.path="))
	relResolved := strings.TrimPrefix(opts[3], "include.path=")
	require.True(t, filepath.IsAbs(relResolved))
	require.True(t, strings.HasSuffix(relResolved, filepath.Join(".git", "rel", "ext.conf")))
	require.Equal(t, "-c", opts[4])
	require.Equal(t, "include.path=~/tilde-ext.conf", opts[5])
}

func TestUpdateSubmodulesForwardsIncludePath(t *testing.T) {
	ctx := context.Background()

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	headSHA := strings.TrimSpace(runGit(t, repoDir, "rev-parse", "HEAD"))

	gitmodules := "[submodule \"sub\"]\n\tpath = sub\n\turl = https://werf-test-nonexistent.invalid/sub.git\n"
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, ".gitmodules"), []byte(gitmodules), 0o644))
	runGit(t, repoDir, "update-index", "--add", "--cacheinfo", "160000,"+headSHA+",sub")
	runGit(t, repoDir, "add", ".gitmodules")
	runGit(t, repoDir, "commit", "-m", "add submodule")

	stubPath := filepath.Join(t.TempDir(), "ext.conf")
	stub := "[url \"werf-test-marker://rewritten/\"]\n\tinsteadOf = https://\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o644))
	runGit(t, repoDir, "config", "include.path", stubPath)

	err := updateSubmodules(ctx, filepath.Join(repoDir, ".git"), repoDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "werf-test-marker")
}

func TestSwitchWorkTreeNonSubmodulesUnaffectedByIncludePath(t *testing.T) {
	ctx := context.Background()

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	headSHA := strings.TrimSpace(runGit(t, repoDir, "rev-parse", "HEAD"))

	stubPath := filepath.Join(t.TempDir(), "ext.conf")
	stub := "[url \"werf-test-marker://rewritten/\"]\n\tinsteadOf = https://\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o644))
	runGit(t, repoDir, "config", "include.path", stubPath)

	workTreeDir := filepath.Join(t.TempDir(), "worktree")
	err := switchWorkTree(ctx, filepath.Join(repoDir, ".git"), workTreeDir, headSHA, false)
	require.NoError(t, err)
}
