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
	isolateGitConfigAI(t)

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

	require.NoError(t, syncSubmodules(ctx, filepath.Join(superRepo, ".git"), workTreeDir))
	require.NoError(t, updateSubmodules(ctx, filepath.Join(superRepo, ".git"), workTreeDir))

	content, err := os.ReadFile(filepath.Join(workTreeDir, "sub", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "hello", string(content))
}

func TestAI_UpdateSubmodulesReusesLocalObjectsForNestedWithoutRemote(t *testing.T) {
	ctx := context.Background()
	isolateGitConfigAI(t)

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

	require.NoError(t, syncSubmodules(ctx, filepath.Join(superRepo, ".git"), workTreeDir))
	require.NoError(t, updateSubmodules(ctx, filepath.Join(superRepo, ".git"), workTreeDir))

	content, err := os.ReadFile(filepath.Join(workTreeDir, "mid", "leaf", "leaf.txt"))
	require.NoError(t, err)
	require.Equal(t, "LEAF", string(content))
}

// An already-populated submodule is not re-cloned, so it is fetched from the URL recorded in the
// service worktree's own submodule git dir. Moving the gitlink must therefore still resolve from
// the superproject store rather than from the remote.
func TestAI_UpdateSubmodulesReusesLocalObjectsAfterGitlinkMovedInWarmWorkTree(t *testing.T) {
	ctx := context.Background()
	isolateGitConfigAI(t)

	subRemote := t.TempDir()
	initGitRepoAI(t, subRemote)
	require.NoError(t, os.WriteFile(filepath.Join(subRemote, "file.txt"), []byte("first"), 0o644))
	runGitAI(t, subRemote, "add", ".")
	runGitAI(t, subRemote, "commit", "-m", "first")

	superRepo := t.TempDir()
	initGitRepoAI(t, superRepo)
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", subRemote, "sub")
	runGitAI(t, superRepo, "commit", "-m", "add submodule")
	firstSHA := strings.TrimSpace(runGitAI(t, superRepo, "rev-parse", "HEAD"))

	// Warm the service worktree on the first gitlink, exactly as an earlier werf run would — sync
	// included, since it rewrites the submodule remote from .gitmodules on every run.
	gitDir := filepath.Join(superRepo, ".git")
	workTreeDir := filepath.Join(t.TempDir(), "worktree")
	runGitAI(t, superRepo, "worktree", "add", "--detach", workTreeDir, firstSHA)
	require.NoError(t, syncSubmodules(ctx, gitDir, workTreeDir))
	require.NoError(t, updateSubmodules(ctx, gitDir, workTreeDir))

	// Move the gitlink, and let the superproject store learn the new commit the way CI would.
	require.NoError(t, os.WriteFile(filepath.Join(subRemote, "file.txt"), []byte("second"), 0o644))
	runGitAI(t, subRemote, "commit", "-am", "second")
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "update", "--remote", "sub")
	runGitAI(t, superRepo, "commit", "-am", "bump submodule")
	secondSHA := strings.TrimSpace(runGitAI(t, superRepo, "rev-parse", "HEAD"))
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "update", "--init")

	// Only now drop the remote: the moved commit exists locally, so no fetch may be needed.
	require.NoError(t, os.RemoveAll(subRemote))

	runGitAI(t, workTreeDir, "checkout", "--force", "--detach", secondSHA)
	require.NoError(t, syncSubmodules(ctx, gitDir, workTreeDir))
	require.NoError(t, updateSubmodules(ctx, gitDir, workTreeDir))

	content, err := os.ReadFile(filepath.Join(workTreeDir, "sub", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "second", string(content))
}

// The fetch redirect must reach nested levels too: their git dirs are per-worktree just like the
// top-level one, so a moved nested gitlink is fetched, not re-cloned.
func TestAI_UpdateSubmodulesReusesLocalObjectsAfterNestedGitlinkMovedInWarmWorkTree(t *testing.T) {
	ctx := context.Background()
	isolateGitConfigAI(t)

	leafRemote := t.TempDir()
	initGitRepoAI(t, leafRemote)
	require.NoError(t, os.WriteFile(filepath.Join(leafRemote, "leaf.txt"), []byte("first"), 0o644))
	runGitAI(t, leafRemote, "add", ".")
	runGitAI(t, leafRemote, "commit", "-m", "first")

	midRemote := t.TempDir()
	initGitRepoAI(t, midRemote)
	runGitAI(t, midRemote, "-c", "protocol.file.allow=always", "submodule", "add", leafRemote, "leaf")
	runGitAI(t, midRemote, "commit", "-m", "add leaf")

	superRepo := t.TempDir()
	initGitRepoAI(t, superRepo)
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", midRemote, "mid")
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "update", "--init", "--recursive")
	runGitAI(t, superRepo, "commit", "-m", "add mid")
	firstSHA := strings.TrimSpace(runGitAI(t, superRepo, "rev-parse", "HEAD"))

	gitDir := filepath.Join(superRepo, ".git")
	workTreeDir := filepath.Join(t.TempDir(), "worktree")
	runGitAI(t, superRepo, "worktree", "add", "--detach", workTreeDir, firstSHA)
	require.NoError(t, syncSubmodules(ctx, gitDir, workTreeDir))
	require.NoError(t, updateSubmodules(ctx, gitDir, workTreeDir))

	// Move the NESTED gitlink and let the superproject stores learn the new commits.
	require.NoError(t, os.WriteFile(filepath.Join(leafRemote, "leaf.txt"), []byte("second"), 0o644))
	runGitAI(t, leafRemote, "commit", "-am", "second")
	runGitAI(t, midRemote, "-c", "protocol.file.allow=always", "submodule", "update", "--remote", "leaf")
	runGitAI(t, midRemote, "commit", "-am", "bump leaf")
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "update", "--remote", "mid")
	runGitAI(t, superRepo, "commit", "-am", "bump mid")
	secondSHA := strings.TrimSpace(runGitAI(t, superRepo, "rev-parse", "HEAD"))
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "update", "--init", "--recursive")

	require.NoError(t, os.RemoveAll(leafRemote))
	require.NoError(t, os.RemoveAll(midRemote))

	runGitAI(t, workTreeDir, "checkout", "--force", "--detach", secondSHA)
	require.NoError(t, syncSubmodules(ctx, gitDir, workTreeDir))
	require.NoError(t, updateSubmodules(ctx, gitDir, workTreeDir))

	content, err := os.ReadFile(filepath.Join(workTreeDir, "mid", "leaf", "leaf.txt"))
	require.NoError(t, err)
	require.Equal(t, "second", string(content))
}

// GitLab CI clones submodules shallow for GIT_SUBMODULE_DEPTH. A shallow store is not a partial
// one — it has every object of the commits it holds — so it must stay eligible for reuse.
func TestAI_UpdateSubmodulesReusesLocalObjectsFromShallowSubmoduleStore(t *testing.T) {
	ctx := context.Background()
	isolateGitConfigAI(t)

	subRemote := t.TempDir()
	initGitRepoAI(t, subRemote)
	for _, marker := range []string{"first", "second"} {
		require.NoError(t, os.WriteFile(filepath.Join(subRemote, "file.txt"), []byte(marker), 0o644))
		runGitAI(t, subRemote, "add", ".")
		runGitAI(t, subRemote, "commit", "-m", marker)
	}

	superRepo := t.TempDir()
	initGitRepoAI(t, superRepo)
	// A depth-limited clone is only honored over file://, not over a plain local path.
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", "file://"+subRemote, "sub")
	runGitAI(t, superRepo, "commit", "-m", "add submodule")
	runGitAI(t, superRepo, "submodule", "deinit", "-f", "sub")
	require.NoError(t, os.RemoveAll(filepath.Join(superRepo, ".git", "modules", "sub")))
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "update", "--init", "--depth=1")
	require.FileExists(t, filepath.Join(superRepo, ".git", "modules", "sub", "shallow"))
	headSHA := strings.TrimSpace(runGitAI(t, superRepo, "rev-parse", "HEAD"))

	require.NoError(t, os.RemoveAll(subRemote))

	gitDir := filepath.Join(superRepo, ".git")
	workTreeDir := filepath.Join(t.TempDir(), "worktree")
	runGitAI(t, superRepo, "worktree", "add", "--detach", workTreeDir, headSHA)
	require.NoError(t, syncSubmodules(ctx, gitDir, workTreeDir))
	require.NoError(t, updateSubmodules(ctx, gitDir, workTreeDir))

	content, err := os.ReadFile(filepath.Join(workTreeDir, "sub", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "second", string(content))
}

// GIT_SUBMODULE_FORCE_HTTPS makes GitLab CI install its own broad insteadOf prefix rewrite. Ours is
// keyed on the full submodule URL, and git resolves insteadOf by longest match, so reuse must win.
func TestAI_UpdateSubmodulesReusesLocalObjectsDespiteExistingInsteadOfRule(t *testing.T) {
	ctx := context.Background()
	isolateGitConfigAI(t)

	subRemote := t.TempDir()
	initGitRepoAI(t, subRemote)
	require.NoError(t, os.WriteFile(filepath.Join(subRemote, "file.txt"), []byte("hello"), 0o644))
	runGitAI(t, subRemote, "add", ".")
	runGitAI(t, subRemote, "commit", "-m", "content")

	superRepo := t.TempDir()
	initGitRepoAI(t, superRepo)
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", subRemote, "sub")
	runGitAI(t, superRepo, "commit", "-m", "add submodule")
	headSHA := strings.TrimSpace(runGitAI(t, superRepo, "rev-parse", "HEAD"))

	// Redirect the remote's parent prefix at a path that does not exist, standing in for the CI's
	// token rewrite: reuse must not depend on where that rule points.
	runGitAI(t, superRepo, "config", "url."+filepath.Join(t.TempDir(), "tokenized")+"/.insteadOf", filepath.Dir(subRemote)+"/")
	require.NoError(t, os.RemoveAll(subRemote))

	gitDir := filepath.Join(superRepo, ".git")
	workTreeDir := filepath.Join(t.TempDir(), "worktree")
	runGitAI(t, superRepo, "worktree", "add", "--detach", workTreeDir, headSHA)
	require.NoError(t, syncSubmodules(ctx, gitDir, workTreeDir))
	require.NoError(t, updateSubmodules(ctx, gitDir, workTreeDir))

	content, err := os.ReadFile(filepath.Join(workTreeDir, "sub", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "hello", string(content))
}

// A submodule name reused at another nesting level cannot be expressed as a process-global
// submodule.<name>.url, so no override may be emitted: an override for the top-level copy would
// otherwise hijack the nested clone and fail the whole update.
func TestAI_UpdateSubmodulesSkipsOverridesOnDuplicateSubmoduleName(t *testing.T) {
	ctx := context.Background()
	isolateGitConfigAI(t)

	newLib := func(marker string) string {
		dir := t.TempDir()
		initGitRepoAI(t, dir)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "who.txt"), []byte(marker), 0o644))
		runGitAI(t, dir, "add", ".")
		runGitAI(t, dir, "commit", "-m", "lib "+marker)
		return dir
	}
	libA := newLib("A")
	libB := newLib("B")

	// mid carries a DIFFERENT repo under the same submodule name "lib".
	midRemote := t.TempDir()
	initGitRepoAI(t, midRemote)
	runGitAI(t, midRemote, "-c", "protocol.file.allow=always", "submodule", "add", libB, "lib")
	runGitAI(t, midRemote, "commit", "-m", "add lib B")

	superRepo := t.TempDir()
	initGitRepoAI(t, superRepo)
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", libA, "lib")
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", midRemote, "mid")
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "update", "--init", "--recursive")
	runGitAI(t, superRepo, "commit", "-m", "add lib A and mid")
	headSHA := strings.TrimSpace(runGitAI(t, superRepo, "rev-parse", "HEAD"))

	overrides, blockers, err := submoduleLocalURLOverrides(ctx, superRepo)
	require.NoError(t, err)
	require.Empty(t, overrides, "no override may be emitted for a duplicated submodule name")
	require.NotEmpty(t, blockers)

	// Remotes stay reachable here, so the plain remote update must still produce correct content for
	// BOTH same-named submodules. These test "remotes" are local paths, hence the explicit transport
	// allowance that a real https remote would not need.
	t.Setenv("GIT_ALLOW_PROTOCOL", "file")
	workTreeDir := filepath.Join(t.TempDir(), "worktree")
	runGitAI(t, superRepo, "worktree", "add", "--detach", workTreeDir, headSHA)
	require.NoError(t, syncSubmodules(ctx, filepath.Join(superRepo, ".git"), workTreeDir))
	require.NoError(t, updateSubmodules(ctx, filepath.Join(superRepo, ".git"), workTreeDir))

	topWho, err := os.ReadFile(filepath.Join(workTreeDir, "lib", "who.txt"))
	require.NoError(t, err)
	require.Equal(t, "A", string(topWho))
	nestedWho, err := os.ReadFile(filepath.Join(workTreeDir, "mid", "lib", "who.txt"))
	require.NoError(t, err)
	require.Equal(t, "B", string(nestedWho))
}

// protocol.file.allow=always is only safe while every URL of the invocation is one werf computed,
// so a submodule that cannot be served locally must suppress the overrides entirely — otherwise the
// relaxed transport would also un-block a file:// URL taken from committed .gitmodules
// (CVE-2022-39253 class).
func TestAI_UpdateSubmodulesKeepsFileTransportBlockedForNonLocalSubmodule(t *testing.T) {
	ctx := context.Background()
	isolateGitConfigAI(t)

	okRemote := t.TempDir()
	initGitRepoAI(t, okRemote)
	require.NoError(t, os.WriteFile(filepath.Join(okRemote, "ok.txt"), []byte("OK"), 0o644))
	runGitAI(t, okRemote, "add", ".")
	runGitAI(t, okRemote, "commit", "-m", "ok content")

	evilRemote := t.TempDir()
	initGitRepoAI(t, evilRemote)
	require.NoError(t, os.WriteFile(filepath.Join(evilRemote, "evil.txt"), []byte("EVIL"), 0o644))
	runGitAI(t, evilRemote, "add", ".")
	runGitAI(t, evilRemote, "commit", "-m", "evil content")

	superRepo := t.TempDir()
	initGitRepoAI(t, superRepo)
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", okRemote, "ok")
	runGitAI(t, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", "file://"+evilRemote, "evil")
	runGitAI(t, superRepo, "commit", "-m", "add submodules")
	headSHA := strings.TrimSpace(runGitAI(t, superRepo, "rev-parse", "HEAD"))

	// Only "ok" is available locally: deinit "evil" so its module store no longer holds the pin.
	runGitAI(t, superRepo, "submodule", "deinit", "-f", "evil")
	require.NoError(t, os.RemoveAll(filepath.Join(superRepo, ".git", "modules", "evil")))

	overrides, blockers, err := submoduleLocalURLOverrides(ctx, superRepo)
	require.NoError(t, err)
	require.Empty(t, overrides, "a non-local submodule must suppress all overrides")
	require.NotEmpty(t, blockers)

	workTreeDir := filepath.Join(t.TempDir(), "worktree")
	runGitAI(t, superRepo, "worktree", "add", "--detach", workTreeDir, headSHA)

	require.NoError(t, syncSubmodules(ctx, filepath.Join(superRepo, ".git"), workTreeDir))
	err = updateSubmodules(ctx, filepath.Join(superRepo, ".git"), workTreeDir)
	require.Error(t, err, "git must still refuse the file:// submodule")
	require.Contains(t, err.Error(), "transport 'file' not allowed")
	require.NoFileExists(t, filepath.Join(workTreeDir, "evil", "evil.txt"))
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
