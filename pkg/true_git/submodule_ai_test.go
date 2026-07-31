package true_git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAI_GetIncludePathOptions(t *testing.T) {
	ctx := context.Background()

	repoDir := t.TempDir()
	initGitRepoAI(t, repoDir)
	gitDir := filepath.Join(repoDir, ".git")

	opts, err := getIncludePathOptions(ctx, gitDir)
	require.NoError(t, err)
	require.Empty(t, opts)

	runGitAI(t, repoDir, "config", "--add", "include.path", "/abs/ext.conf")
	runGitAI(t, repoDir, "config", "--add", "include.path", "rel/ext.conf")
	runGitAI(t, repoDir, "config", "--add", "include.path", "~/tilde-ext.conf")

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

func TestAI_UpdateSubmodulesForwardsIncludePath(t *testing.T) {
	ctx := context.Background()

	repoDir := t.TempDir()
	initGitRepoAI(t, repoDir)
	headSHA := strings.TrimSpace(runGitAI(t, repoDir, "rev-parse", "HEAD"))

	gitmodules := "[submodule \"sub\"]\n\tpath = sub\n\turl = https://werf-test-nonexistent.invalid/sub.git\n"
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, ".gitmodules"), []byte(gitmodules), 0o644))
	runGitAI(t, repoDir, "update-index", "--add", "--cacheinfo", "160000,"+headSHA+",sub")
	runGitAI(t, repoDir, "add", ".gitmodules")
	runGitAI(t, repoDir, "commit", "-m", "add submodule")

	stubPath := filepath.Join(t.TempDir(), "ext.conf")
	stub := "[url \"werf-test-marker://rewritten/\"]\n\tinsteadOf = https://\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o644))
	runGitAI(t, repoDir, "config", "include.path", stubPath)

	err := updateSubmodules(ctx, filepath.Join(repoDir, ".git"), repoDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "werf-test-marker")
}

func TestAI_UpdateSubmodulesReusesLocalObjectsWithoutRemote(t *testing.T) {
	ctx := context.Background()

	subRemote := t.TempDir()
	initGitRepoAI(t, subRemote)
	require.NoError(t, os.WriteFile(filepath.Join(subRemote, "file.txt"), []byte("hello"), 0o644))
	runGitAI(t, subRemote, "add", ".")
	runGitAI(t, subRemote, "commit", "-m", "sub content")

	superRepo := t.TempDir()
	initGitRepoAI(t, superRepo)
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", subRemote, "sub")
	runGitAI(t, superRepo, "commit", "-m", "add submodule")
	headSHA := strings.TrimSpace(runGitAI(t, superRepo, "rev-parse", "HEAD"))

	// `submodule add` already populated superRepo/.git/modules/sub; drop the remote to prove the
	// checkout into a fresh worktree reuses those local objects and touches no network.
	require.NoError(t, os.RemoveAll(subRemote))

	workTreeDir := filepath.Join(t.TempDir(), "worktree")
	runGitAI(t, superRepo, "worktree", "add", "--detach", workTreeDir, headSHA)

	err := updateSubmodules(ctx, filepath.Join(superRepo, ".git"), workTreeDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(workTreeDir, "sub", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "hello", string(content))
}

func TestAI_UpdateSubmodulesReusesLocalObjectsForNestedWithoutRemote(t *testing.T) {
	ctx := context.Background()

	leafRemote := t.TempDir()
	initGitRepoAI(t, leafRemote)
	require.NoError(t, os.WriteFile(filepath.Join(leafRemote, "leaf.txt"), []byte("LEAF"), 0o644))
	runGitAI(t, leafRemote, "add", ".")
	runGitAI(t, leafRemote, "commit", "-m", "leaf content")

	midRemote := t.TempDir()
	initGitRepoAI(t, midRemote)
	runGitAI(t, midRemote, "-c", "protocol.file.allow=always", "submodule", "add", leafRemote, "leaf")
	runGitAI(t, midRemote, "commit", "-m", "add leaf")

	superRepo := t.TempDir()
	initGitRepoAI(t, superRepo)
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", midRemote, "mid")
	// `submodule add` populates only the top module store; recurse once to fill mid/modules/leaf.
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "update", "--init", "--recursive")
	runGitAI(t, superRepo, "commit", "-m", "add mid")
	headSHA := strings.TrimSpace(runGitAI(t, superRepo, "rev-parse", "HEAD"))

	// Drop every remote to prove both levels are checked out from the local object store alone.
	require.NoError(t, os.RemoveAll(leafRemote))
	require.NoError(t, os.RemoveAll(midRemote))

	workTreeDir := filepath.Join(t.TempDir(), "worktree")
	runGitAI(t, superRepo, "worktree", "add", "--detach", workTreeDir, headSHA)

	err := updateSubmodules(ctx, filepath.Join(superRepo, ".git"), workTreeDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(workTreeDir, "mid", "leaf", "leaf.txt"))
	require.NoError(t, err)
	require.Equal(t, "LEAF", string(content))
}

func TestAI_SwitchWorkTreeNonSubmodulesUnaffectedByIncludePath(t *testing.T) {
	ctx := context.Background()

	repoDir := t.TempDir()
	initGitRepoAI(t, repoDir)
	headSHA := strings.TrimSpace(runGitAI(t, repoDir, "rev-parse", "HEAD"))

	stubPath := filepath.Join(t.TempDir(), "ext.conf")
	stub := "[url \"werf-test-marker://rewritten/\"]\n\tinsteadOf = https://\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o644))
	runGitAI(t, repoDir, "config", "include.path", stubPath)

	workTreeDir := filepath.Join(t.TempDir(), "worktree")
	err := switchWorkTree(ctx, filepath.Join(repoDir, ".git"), workTreeDir, headSHA, false)
	require.NoError(t, err)
}
