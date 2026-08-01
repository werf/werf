package true_git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/werf/logboek"
)

func syncSubmodules(ctx context.Context, repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Sync submodules in work tree %q", workTreeDir)
	return logboek.Context(ctx).Info().LogProcess(logProcessMsg).DoError(func() error {
		includePathOpts, err := getIncludePathOptions(ctx, repoDir)
		if err != nil {
			return err
		}

		submSyncCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, append(includePathOpts, "submodule", "sync", "--recursive")...)
		if err := submSyncCmd.Run(ctx); err != nil {
			return fmt.Errorf("submodule sync command failed: %w", err)
		}

		return nil
	})
}

func updateSubmodules(ctx context.Context, repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Update submodules in work tree %q", workTreeDir)
	return logboek.Context(ctx).Info().LogProcess(logProcessMsg).DoError(func() error {
		includePathOpts, err := getIncludePathOptions(ctx, repoDir)
		if err != nil {
			return err
		}

		localURLOpts, blockers, err := submoduleLocalURLOverrides(ctx, workTreeDir)
		if err != nil {
			return err
		}
		if len(blockers) > 0 {
			logboek.Context(ctx).Warn().LogF(
				"WARNING: unable to reuse local git objects for submodules: %s.\nAll submodules will be checked out from their remotes, which may request credentials. Run `git submodule update --init --recursive` in your working tree beforehand to let werf reuse the local objects instead.\n",
				strings.Join(blockers, ", "),
			)
		}

		runUpdate := func(reuseOpts []string) error {
			updateArgs := make([]string, 0, len(includePathOpts)+len(reuseOpts)+8)
			updateArgs = append(updateArgs, includePathOpts...)
			if len(reuseOpts) > 0 {
				// The redirects point submodule URLs at the on-disk module store, which git reaches
				// over the file transport — blocked by the default protocol.file.allow=user for
				// submodules. Safe only because reuse is all-or-nothing: every URL used by this
				// invocation is then a path werf computed itself, never one from committed
				// .gitmodules.
				updateArgs = append(updateArgs, "-c", "protocol.file.allow=always")
				updateArgs = append(updateArgs, reuseOpts...)
			}
			updateArgs = append(updateArgs, "submodule", "update", "--checkout", "--force", "--init", "--recursive")

			submUpdateCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, updateArgs...)
			return submUpdateCmd.Run(ctx)
		}

		if err := runUpdate(localURLOpts); err != nil {
			// Reuse rests on where git keeps submodule git dirs and on -c reaching nested updates,
			// neither of which git promises. Anything unanticipated there must cost a slower build,
			// not a broken one, so fall back to the plain remote update this has always been.
			if len(localURLOpts) == 0 {
				return fmt.Errorf("submodule update command failed: %w", err)
			}
			logboek.Context(ctx).Warn().LogF("WARNING: unable to check out submodules from the local object store (%s), retrying from their remotes, which may request credentials.\n", err)
			// The failed attempt leaves the service worktree's own submodule git dirs pointing at the
			// store, and an initialized submodule is never re-cloned, so the retry would inherit the
			// very state that just failed. Discard them; they belong to this worktree alone.
			if discardErr := discardWorktreeSubmoduleGitDirs(ctx, workTreeDir); discardErr != nil {
				return fmt.Errorf("submodule update command failed: %w (unable to discard the failed attempt: %s)", err, discardErr)
			}
			if retryErr := runUpdate(nil); retryErr != nil {
				return fmt.Errorf("submodule update command failed: %w (local object store reuse had failed with: %s)", retryErr, err)
			}
		}

		return nil
	})
}

type submoduleNode struct {
	name string
	// displayPath is the slash-joined path through the tree, used only for messages.
	displayPath string
	// moduleDir is the superproject-side store the objects are reused from.
	moduleDir string
	// worktreeModuleDir is where git keeps this submodule's git dir for the service worktree. It is
	// separate from moduleDir, which is exactly why an already-populated submodule fetches instead
	// of reusing what the superproject already has.
	worktreeModuleDir string
}

// submoduleLocalURLOverrides walks the superproject's submodules recursively and, when EVERY
// submodule's pinned commit is already present in the local module store
// (<parent-module-dir>/modules/<name>, rooted at <git-common-dir>), returns
// `-c submodule.<name>.url=<local-path>` overrides so `git submodule update --recursive` checks them
// all out from those local objects instead of their remotes — no network, no credentials. These -c
// options propagate into the nested update invocations via GIT_CONFIG_PARAMETERS.
//
// The overrides are deliberately all-or-nothing: whatever cannot be served locally is reported in
// blockers and NO override is emitted, leaving the plain remote update werf has always done. Partial
// reuse is unsafe because `submodule.<name>.url` is global to the git process, so an override for one
// submodule also redirects a same-named submodule elsewhere in the tree — and a subtree that is not
// available locally cannot be enumerated at all, so such a collision could not even be detected.
// All-or-nothing additionally keeps `protocol.file.allow=always`, which the local paths require, from
// ever applying to a URL werf did not compute itself.
func submoduleLocalURLOverrides(ctx context.Context, workTreeDir string) (overrides, blockers []string, err error) {
	names, paths, err := parseGitmodulesPaths(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "config", "-f", ".gitmodules")
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list submodules: %w", err)
	}
	if len(names) == 0 {
		return nil, nil, nil
	}

	commonDir, err := gitCommonDir(ctx, workTreeDir)
	if err != nil {
		return nil, nil, err
	}

	worktreeGitDir, err := gitAbsoluteGitDir(ctx, workTreeDir)
	if err != nil {
		return nil, nil, err
	}

	var nodes []submoduleNode
	for i, name := range names {
		pinnedCommit, err := lsTreeGitlink(ctx, workTreeDir, "HEAD", paths[i])
		if err != nil {
			return nil, nil, err
		}
		if err := walkSubmodule(ctx, commonDir, worktreeGitDir, name, paths[i], pinnedCommit, &nodes, &blockers); err != nil {
			return nil, nil, err
		}
	}

	seenName := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		if _, ok := seenName[node.name]; ok {
			blockers = append(blockers, fmt.Sprintf("%s (submodule name %q is used more than once)", node.displayPath, node.name))
		}
		seenName[node.name] = struct{}{}
	}

	// An already-populated submodule is not re-cloned, so submodule.<name>.url does not reach it:
	// git fetches the moved gitlink from the URL recorded in the service worktree's own submodule
	// git dir. Rewriting that URL to the superproject store keeps such a fetch local too.
	fetchURLs := make(map[string]string, len(nodes))
	seenURL := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		if node.worktreeModuleDir == "" {
			continue
		}
		fetchURL, err := gitConfigValue(ctx, node.worktreeModuleDir, "remote.origin.url")
		if err != nil {
			return nil, nil, err
		}
		if fetchURL == "" {
			blockers = append(blockers, fmt.Sprintf("%s (submodule has no remote URL to redirect)", node.displayPath))
			continue
		}
		// Nothing to redirect when the recorded remote is the store itself, and skipping it also
		// keeps the store-prefix blocker below from matching this URL against its own store and
		// disabling reuse outright.
		if fetchURL == node.moduleDir {
			continue
		}
		if _, ok := seenURL[fetchURL]; ok {
			blockers = append(blockers, fmt.Sprintf("%s (remote URL is shared with another submodule)", node.displayPath))
			continue
		}
		seenURL[fetchURL] = struct{}{}
		fetchURLs[node.displayPath] = fetchURL
	}

	// insteadOf matches by prefix, so a remote URL that prefixes one of the store paths would also
	// rewrite the store URL the clone path relies on.
	for _, owner := range nodes {
		fetchURL, ok := fetchURLs[owner.displayPath]
		if !ok {
			continue
		}
		for _, node := range nodes {
			if strings.HasPrefix(node.moduleDir, fetchURL) {
				blockers = append(blockers, fmt.Sprintf("%s (its remote URL prefixes the object store path of %s)", owner.displayPath, node.displayPath))
			}
		}
	}

	if len(blockers) > 0 {
		return nil, blockers, nil
	}

	for _, node := range nodes {
		overrides = append(overrides, "-c", "submodule."+node.name+".url="+node.moduleDir)
		if fetchURL, ok := fetchURLs[node.displayPath]; ok {
			overrides = append(overrides, "-c", "url."+node.moduleDir+".insteadOf="+fetchURL)
		}
	}

	return overrides, nil, nil
}

// discardWorktreeSubmoduleGitDirs drops everything the service worktree holds for its submodules,
// so the next update clones them afresh. Submodule working directories are tracked gitlinks, which
// git clean leaves alone, and a leftover .git file there would outlive its discarded git dir.
//
// Everything removed here must belong to the service worktree alone. A linked worktree has its own
// git dir, so its modules/ holds only its submodule git dirs; in a main worktree that same path is
// the repository-wide submodule store, and removing it would destroy the user's own submodule data.
// Refuse in that case rather than trust the caller.
func discardWorktreeSubmoduleGitDirs(ctx context.Context, workTreeDir string) error {
	worktreeGitDir, err := gitAbsoluteGitDir(ctx, workTreeDir)
	if err != nil {
		return err
	}
	commonDir, err := gitCommonDir(ctx, workTreeDir)
	if err != nil {
		return err
	}
	// The two are compared through EvalSymlinks because they are produced differently: git resolves
	// the one, werf joins the other onto a caller-supplied path that may still hold a symlink.
	if resolvePathForCompare(worktreeGitDir) == resolvePathForCompare(commonDir) {
		return fmt.Errorf("refusing to discard submodule git dirs of %s: not a linked work tree", workTreeDir)
	}

	if err := os.RemoveAll(filepath.Join(worktreeGitDir, "modules")); err != nil {
		return fmt.Errorf("unable to discard submodule git dirs of work tree %s: %w", workTreeDir, err)
	}

	_, paths, err := parseGitmodulesPaths(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "config", "-f", ".gitmodules")
	if err != nil {
		return err
	}
	for _, path := range paths {
		// .gitmodules is repo-controlled and git validates none of it, so two things must hold before
		// a path reaches RemoveAll. It has to be work-tree local, which rules out an absolute or
		// ".." path git would anyway refuse as a pathspec; and HEAD has to record it as a gitlink,
		// which rules out a symlinked component RemoveAll would follow out of the work tree, since
		// every component leading to a gitlink is itself a tree.
		if !filepath.IsLocal(path) {
			continue
		}
		gitlink, err := lsTreeGitlink(ctx, workTreeDir, "HEAD", path)
		if err != nil {
			return err
		}
		if gitlink == "" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(workTreeDir, path)); err != nil {
			return fmt.Errorf("unable to discard submodule work tree %s: %w", path, err)
		}
	}

	return nil
}

func resolvePathForCompare(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

func gitAbsoluteGitDir(ctx context.Context, workTreeDir string) (string, error) {
	revParseCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "rev-parse", "--absolute-git-dir")
	if err := revParseCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git rev-parse --absolute-git-dir command failed: %w", err)
	}
	return strings.TrimSpace(revParseCmd.OutBuf.String()), nil
}

func gitConfigValue(ctx context.Context, repoDir, key string) (string, error) {
	configCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoDir}, "config", "--get", key)
	if err := configCmd.Run(ctx); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", fmt.Errorf("git config command failed: %w", err)
	}
	return strings.TrimSpace(configCmd.OutBuf.String()), nil
}

// walkSubmodule resolves one submodule against the local module store and recurses into its nested
// submodules, reading their definitions from the module store at the pinned commit (no checkout
// needed). parentModuleDir is the git dir whose modules/<name> hosts this submodule's store.
func walkSubmodule(ctx context.Context, parentModuleDir, parentWorktreeModuleDir, name, displayPath, pinnedCommit string, nodes *[]submoduleNode, blockers *[]string) error {
	// A .gitmodules entry without a gitlink in the tree is not checked out by git either.
	if pinnedCommit == "" {
		return nil
	}

	if submoduleNameUnsafe(name) {
		*blockers = append(*blockers, fmt.Sprintf("%s (unsupported submodule name %q)", displayPath, name))
		return nil
	}

	moduleDir := filepath.Join(parentModuleDir, "modules", name)
	if !moduleStoreHasCommit(ctx, moduleDir, pinnedCommit) {
		*blockers = append(*blockers, fmt.Sprintf("%s (not initialized locally)", displayPath))
		return nil
	}

	// A partial clone can hold the commit while its blobs still live on the promisor remote, so
	// cloning from it would fail or need the network — exactly what reuse is meant to avoid.
	isPartial, err := moduleStoreIsPartial(ctx, moduleDir)
	if err != nil {
		return err
	}
	if isPartial {
		*blockers = append(*blockers, fmt.Sprintf("%s (local clone is partial)", displayPath))
		return nil
	}

	// Absent parent means the whole subtree is unpopulated in the service worktree; keep the empty
	// marker instead of joining onto it, which would yield a relative path probed against the CWD.
	var worktreeModuleDir string
	if parentWorktreeModuleDir != "" {
		candidate := filepath.Join(parentWorktreeModuleDir, "modules", name)
		if _, err := os.Stat(candidate); err == nil {
			worktreeModuleDir = candidate
		}
	}

	*nodes = append(*nodes, submoduleNode{
		name:              name,
		displayPath:       displayPath,
		moduleDir:         moduleDir,
		worktreeModuleDir: worktreeModuleDir,
	})

	nestedNames, nestedPaths, err := parseGitmodulesPaths(ctx, &GitCmdOptions{RepoDir: moduleDir}, "config", "--blob", pinnedCommit+":.gitmodules")
	if err != nil {
		return err
	}
	for i, nestedName := range nestedNames {
		nestedPin, err := lsTreeGitlink(ctx, moduleDir, pinnedCommit, nestedPaths[i])
		if err != nil {
			return err
		}
		if err := walkSubmodule(ctx, moduleDir, worktreeModuleDir, nestedName, displayPath+"/"+nestedPaths[i], nestedPin, nodes, blockers); err != nil {
			return err
		}
	}

	return nil
}

// submoduleNameUnsafe rejects names that cannot be expressed safely as a `-c submodule.<name>.url`
// option or that would resolve outside the module store: `=` terminates the config key early and
// would set an unrelated option, and separators or `.`/`..` components would move the resolved path.
// .gitmodules is repo-controlled and `git config -f` does not reject these itself.
func submoduleNameUnsafe(name string) bool {
	if strings.ContainsAny(name, "=\n/\\") {
		return true
	}
	return name == "." || name == ".."
}

func moduleStoreIsPartial(ctx context.Context, moduleDir string) (bool, error) {
	configCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: moduleDir}, "config", "--get-regexp", "^(extensions\\.partialclone|remote\\..*\\.(promisor|partialclonefilter))$")
	if err := configCmd.Run(ctx); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("git config command failed: %w", err)
	}
	return configCmd.OutBuf.Len() > 0, nil
}

// parseGitmodulesPaths runs a `git config ... -z --get-regexp \.path$` over a .gitmodules source
// (a working-tree file via -f, or a tree blob via --blob) and returns the submodule names and
// paths. A missing source or no matches (git exit code 1) yields empty slices, not an error.
func parseGitmodulesPaths(ctx context.Context, opts *GitCmdOptions, configArgs ...string) (names, paths []string, err error) {
	args := make([]string, 0, len(configArgs)+3)
	args = append(args, configArgs...)
	args = append(args, "-z", "--get-regexp", "\\.path$")
	configCmd := NewGitCmd(ctx, opts, args...)
	if err := configCmd.Run(ctx); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("git config command failed: %w", err)
	}

	for _, entry := range strings.Split(configCmd.OutBuf.String(), "\x00") {
		if entry == "" {
			continue
		}
		key, value, found := strings.Cut(entry, "\n")
		if !found {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(key, "submodule."), ".path")
		if name == "" || value == "" {
			continue
		}
		names = append(names, name)
		paths = append(paths, value)
	}

	return names, paths, nil
}

func lsTreeGitlink(ctx context.Context, repoDir, treeish, submodulePath string) (string, error) {
	lsTreeCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoDir}, "ls-tree", treeish, "--", ":(literal)"+submodulePath)
	if err := lsTreeCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git ls-tree command failed: %w", err)
	}

	fields := strings.Fields(lsTreeCmd.OutBuf.String())
	if len(fields) < 3 || fields[1] != "commit" {
		return "", nil
	}
	return fields[2], nil
}

func moduleStoreHasCommit(ctx context.Context, moduleDir, commit string) bool {
	if _, err := os.Stat(moduleDir); err != nil {
		return false
	}
	// GIT_NO_LAZY_FETCH keeps the probe itself from reaching out to a promisor remote (git 2.41+).
	catFileCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: moduleDir, Env: []string{"GIT_NO_LAZY_FETCH=1"}}, "cat-file", "-e", commit+"^{commit}")
	return catFileCmd.Run(ctx) == nil
}

func gitCommonDir(ctx context.Context, workTreeDir string) (string, error) {
	revParseCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "rev-parse", "--git-common-dir")
	if err := revParseCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git rev-parse --git-common-dir command failed: %w", err)
	}
	commonDir := strings.TrimSpace(revParseCmd.OutBuf.String())
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(workTreeDir, commonDir)
	}
	return commonDir, nil
}
