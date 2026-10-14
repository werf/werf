package true_git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/opstats"
)

const disableShallowSubmodulesEnvVar = "WERF_DISABLE_SHALLOW_SUBMODULES"

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
	return updateSubmodulesWithCheckoutOptions(ctx, repoDir, workTreeDir, nil)
}

func updateSubmodulesWithCheckoutOptions(ctx context.Context, repoDir, workTreeDir string, checkoutGitOptions []string) error {
	defer opstats.Observe(ctx, opstats.OperationGitSubmoduleUpdate)()

	logProcessMsg := fmt.Sprintf("Update submodules in work tree %q", workTreeDir)
	return logboek.Context(ctx).Info().LogProcess(logProcessMsg).DoError(func() error {
		includePathOpts, err := getIncludePathOptions(ctx, repoDir)
		if err != nil {
			return err
		}

		localURLOpts, blockers, err := submoduleLocalURLOverrides(ctx, workTreeDir)
		if err != nil {
			// Reuse is an optimization, so a failure to even determine whether it applies must cost a
			// slower build rather than a failed one, exactly like a store that cannot serve the checkout.
			localURLOpts, blockers = nil, []string{fmt.Sprintf("unable to inspect the local object store: %s", err)}
		}
		if len(blockers) > 0 {
			// The blockers describe the user's checkout, not this worktree switch, so warning on every
			// switch of every worktree in a build would repeat the same text many times.
			submoduleReuseWarnOnce.Do(func() {
				logboek.Context(ctx).Warn().LogF(
					"WARNING: unable to reuse local git objects for submodules: %s.\nAll submodules will be checked out from their remotes, which may request credentials. Run `git submodule update --init --recursive` in your working tree beforehand to let werf reuse the local objects instead.\n",
					strings.Join(blockers, ", "),
				)
			})
		}

		gitOpts := slices.Concat(checkoutGitOptions, includePathOpts)

		if len(localURLOpts) > 0 {
			updateArgs := slices.Clone(gitOpts)
			// The redirects point submodule URLs at the on-disk module store, which git reaches
			// over the file transport — blocked by the default protocol.file.allow=user for
			// submodules. Safe only because reuse is all-or-nothing: no URL read from committed
			// .gitmodules is then used while the transport is relaxed (CVE-2022-39253 class).
			// Ambient trusted configuration may still rewrite the store paths werf computes.
			updateArgs = append(updateArgs, "-c", "protocol.file.allow=always")
			updateArgs = append(updateArgs, localURLOpts...)
			updateArgs = append(updateArgs, "submodule", "update", "--checkout", "--force", "--init", "--recursive")

			reuseCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, updateArgs...)
			reuseErr := reuseCmd.Run(ctx)
			if reuseErr == nil {
				return finishSubmoduleUpdate(ctx, repoDir, workTreeDir)
			}
			// Reuse rests on where git keeps submodule git dirs and on -c reaching nested updates,
			// neither of which git promises. Anything unanticipated there must cost a slower build,
			// not a broken one, so fall back to the remote update this has always been.
			if ctx.Err() != nil {
				return fmt.Errorf("submodule update command failed: %w", reuseErr)
			}
			logboek.Context(ctx).Warn().LogF("WARNING: unable to check out submodules from the local object store (%s), retrying from their remotes, which may request credentials.\n", reuseErr)
			// The failed attempt leaves the service worktree's own submodule git dirs pointing at the
			// store, and an initialized submodule is never re-cloned, so the retry would inherit the
			// very state that just failed. Discard them; they belong to this worktree alone.
			if discardErr := discardWorktreeSubmoduleGitDirs(ctx, workTreeDir); discardErr != nil {
				return fmt.Errorf("submodule update command failed: %w (unable to discard the failed attempt: %s)", reuseErr, discardErr)
			}
			if err := updateSubmodulesFromRemotes(ctx, workTreeDir, gitOpts); err != nil {
				return fmt.Errorf("%w (local object store reuse had failed with: %s)", err, reuseErr)
			}
			return finishSubmoduleUpdate(ctx, repoDir, workTreeDir)
		}

		if err := updateSubmodulesFromRemotes(ctx, workTreeDir, gitOpts); err != nil {
			return err
		}
		return finishSubmoduleUpdate(ctx, repoDir, workTreeDir)
	})
}

func finishSubmoduleUpdate(ctx context.Context, repoDir, workTreeDir string) error {
	if err := pinSubmoduleCommits(ctx, workTreeDir); err != nil {
		return err
	}
	return initSubmoduleConfig(ctx, repoDir, workTreeDir)
}

// pinSubmoduleCommits records the commit every populated submodule is checked out at under
// refs/werf/commits/ in that submodule's git dir. A detached HEAD is the only thing holding such a
// commit otherwise, so once the gitlink moves on the commit becomes unreachable: git's gc may prune
// it, and `submodule update` sees it as missing and fetches it again when the gitlink moves back,
// which in a shallow clone also happens for every switch to an older commit. The refs live and die
// with the service worktree's submodule git dirs.
func pinSubmoduleCommits(ctx context.Context, workTreeDir string) error {
	pinCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "submodule", "foreach", "--quiet", "--recursive", "git update-ref refs/werf/commits/$sha1 $sha1")
	if err := pinCmd.Run(ctx); err != nil {
		return fmt.Errorf("pin submodule commits: %w", err)
	}
	return nil
}

type submoduleUpdateParams struct {
	workTreeDir string
	rootGitDir  string
	gitOpts     []string
	shallow     bool
}

// updateSubmodulesFromRemotes checks the submodules of workTreeDir out from their remotes one at a
// time and level by level instead of with a single recursive `git submodule update`, because that
// is the only way to clone shallow selectively. --depth applies to every clone AND every fetch of
// one update run, and a fetch with --depth into a full clone turns it shallow, losing history
// behind the commits it holds. So a submodule is cloned or fetched with --depth=1 only when its
// git dir in this worktree is absent or already shallow; one whose .gitmodules entry says
// shallow=false, or whose name is unsafe to resolve into a git dir path, is left to git's defaults.
// A failed shallow attempt is retried once in full for that submodule alone — a server without
// shallow support or one refusing to serve an unadvertised commit still works, just slower.
func updateSubmodulesFromRemotes(ctx context.Context, workTreeDir string, gitOpts []string) error {
	rootGitDir, commonDir, err := resolveWorkTreeGitDirs(ctx, workTreeDir)
	if err != nil {
		return err
	}

	params := submoduleUpdateParams{
		workTreeDir: workTreeDir,
		rootGitDir:  rootGitDir,
		gitOpts:     gitOpts,
		// The full retry discards the failed shallow git dir, which in a main work tree would be the
		// repository-wide submodule store, so only a linked work tree gets the shallow attempt at all.
		shallow: !util.GetBoolEnvironmentDefaultFalse(disableShallowSubmodulesEnvVar) && util.IsSubpathOfBasePath(filepath.Join(commonDir, "worktrees"), rootGitDir),
	}

	return updateSubmoduleLevel(ctx, params, workTreeDir, rootGitDir)
}

func updateSubmoduleLevel(ctx context.Context, params submoduleUpdateParams, levelDir, levelGitDir string) error {
	names, paths, err := parseGitmodulesPaths(ctx, &GitCmdOptions{RepoDir: levelDir}, "config", "-f", ".gitmodules")
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}

	// Without paths `submodule init` honors submodule.active exactly like `update --init` does, and
	// the per-path update then skips whatever stayed uninitialized.
	initCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: levelDir}, append(slices.Clone(params.gitOpts), "submodule", "init")...)
	if err := initCmd.Run(ctx); err != nil {
		return fmt.Errorf("submodule init command failed: %w", err)
	}

	shallowDisabled, err := gitmodulesShallowDisabled(ctx, levelDir)
	if err != nil {
		return err
	}

	for i, name := range names {
		path := paths[i]
		pinnedCommit, err := lsTreeGitlink(ctx, levelDir, "HEAD", path)
		if err != nil {
			return err
		}
		if pinnedCommit == "" {
			continue
		}

		displayPath, err := filepath.Rel(params.workTreeDir, filepath.Join(levelDir, path))
		if err != nil {
			return fmt.Errorf("resolve submodule path %s relative to %s: %w", filepath.Join(levelDir, path), params.workTreeDir, err)
		}
		displayPath = filepath.ToSlash(displayPath)

		moduleDir := filepath.Join(levelGitDir, "modules", name)
		shallow := false
		if params.shallow && !shallowDisabled[name] && !submoduleNameUnsafe(name) {
			if shallow, err = submoduleShallowEligible(moduleDir); err != nil {
				return err
			}
		}

		logboek.Context(ctx).Info().LogF("Update submodule %s to %s (shallow: %t)\n", displayPath, pinnedCommit, shallow)
		if err := updateSubmodulePath(ctx, params, levelDir, path, shallow); err != nil {
			if !shallow || ctx.Err() != nil {
				return fmt.Errorf("update submodule %s: %w", displayPath, err)
			}
			logboek.Context(ctx).Warn().LogF("WARNING: unable to check out submodule %s shallow, checking it out in full: %s\n", displayPath, err)
			if err := discardShallowSubmodule(ctx, params, levelDir, path, moduleDir); err != nil {
				return err
			}
			if err := updateSubmodulePath(ctx, params, levelDir, path, false); err != nil {
				return fmt.Errorf("update submodule %s: %w", displayPath, err)
			}
		}

		submoduleDir := filepath.Join(levelDir, path)
		if _, err := os.Stat(filepath.Join(submoduleDir, ".git")); errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return fmt.Errorf("unable to access %s: %w", filepath.Join(submoduleDir, ".git"), err)
		}

		// The nested git dir is read back from disk rather than assumed, so a name that resolves
		// somewhere other than modules/<name> only costs its subtree the shallow attempt.
		submoduleGitDir, _, err := resolveWorkTreeGitDirs(ctx, submoduleDir)
		if err != nil {
			return err
		}
		nestedParams := params
		nestedParams.shallow = params.shallow && util.IsSubpathOfBasePath(filepath.Join(params.rootGitDir, "modules"), submoduleGitDir)
		if err := updateSubmoduleLevel(ctx, nestedParams, submoduleDir, submoduleGitDir); err != nil {
			return err
		}
	}

	return nil
}

func updateSubmodulePath(ctx context.Context, params submoduleUpdateParams, levelDir, path string, shallow bool) error {
	args := slices.Clone(params.gitOpts)
	args = append(args, "submodule", "update", "--checkout", "--force")
	if shallow {
		args = append(args, "--depth=1")
	}
	args = append(args, "--", ":(literal)"+path)
	updateCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: levelDir}, args...)
	return updateCmd.Run(ctx)
}

// submoduleShallowEligible reports whether --depth=1 can reach moduleDir without taking history
// away: an absent git dir is about to be cloned, and a shallow one has none to lose.
func submoduleShallowEligible(moduleDir string) (bool, error) {
	if _, err := os.Stat(moduleDir); errors.Is(err, fs.ErrNotExist) {
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("unable to access %s: %w", moduleDir, err)
	}
	return util.FileExists(filepath.Join(moduleDir, "shallow"))
}

// discardShallowSubmodule removes what a failed shallow attempt left of one submodule so the full
// retry clones it afresh: an existing shallow git dir is never re-cloned, and a plain fetch into it
// would keep it shallow. A git dir without the shallow marker is left alone — it was never shallow
// and holds history the retry must not destroy.
func discardShallowSubmodule(ctx context.Context, params submoduleUpdateParams, levelDir, path, moduleDir string) error {
	if !util.IsSubpathOfBasePath(filepath.Join(params.rootGitDir, "modules"), moduleDir) {
		return fmt.Errorf("refusing to discard submodule git dir %s: outside of work tree git dir %s", moduleDir, params.rootGitDir)
	}

	isShallow, err := util.FileExists(filepath.Join(moduleDir, "shallow"))
	if err != nil {
		return fmt.Errorf("unable to access %s: %w", filepath.Join(moduleDir, "shallow"), err)
	}
	if !isShallow {
		return nil
	}

	if err := os.RemoveAll(moduleDir); err != nil {
		return fmt.Errorf("unable to discard submodule git dir %s: %w", moduleDir, err)
	}
	// The path is a gitlink of HEAD, checked by the caller, so no component leading to it is a
	// symlink RemoveAll could follow out of the work tree.
	if err := os.RemoveAll(filepath.Join(levelDir, path)); err != nil {
		return fmt.Errorf("unable to discard submodule work tree %s: %w", filepath.Join(levelDir, path), err)
	}
	return nil
}

// gitmodulesShallowDisabled returns the submodule names whose .gitmodules entry explicitly sets
// shallow to false: that is the one way a repository can rule the shallow attempt out for a
// submodule, since git itself ignores the setting once --depth is given.
func gitmodulesShallowDisabled(ctx context.Context, levelDir string) (map[string]bool, error) {
	configCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: levelDir}, "config", "-f", ".gitmodules", "--bool", "-z", "--get-regexp", "^submodule\\..*\\.shallow$")
	if err := configCmd.Run(ctx); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("git config command failed: %w", err)
	}

	disabled := make(map[string]bool)
	for _, entry := range strings.Split(configCmd.OutBuf.String(), "\x00") {
		key, value, found := strings.Cut(entry, "\n")
		if !found || value != "false" {
			continue
		}
		disabled[strings.TrimSuffix(strings.TrimPrefix(key, "submodule."), ".shallow")] = true
	}
	return disabled, nil
}

func initSubmoduleConfig(ctx context.Context, repoDir, workTreeDir string) error {
	includePathOpts, err := getIncludePathOptions(ctx, repoDir)
	if err != nil {
		return err
	}

	// Temporary URL overrides can populate worktrees without persisting the registration go-git needs.
	for _, args := range [][]string{
		{"submodule", "init"},
		{"submodule", "foreach", "--recursive", "git", "submodule", "init"},
		{"submodule", "sync", "--recursive"},
	} {
		cmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, append(includePathOpts, args...)...)
		if err := cmd.Run(ctx); err != nil {
			return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
	}

	return nil
}

var submoduleReuseWarnOnce sync.Once

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
	// fetchURL is the remote recorded in worktreeModuleDir, empty when there is nothing to redirect.
	fetchURL string
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
// ever applying to a URL read from committed .gitmodules.
func submoduleLocalURLOverrides(ctx context.Context, workTreeDir string) ([]string, []string, error) {
	names, paths, err := parseGitmodulesPaths(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "config", "-f", ".gitmodules")
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list submodules: %w", err)
	}
	if len(names) == 0 {
		return nil, nil, nil
	}

	worktreeGitDir, commonDir, err := resolveWorkTreeGitDirs(ctx, workTreeDir)
	if err != nil {
		return nil, nil, err
	}

	var nodes []submoduleNode
	var blockers []string
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
	seenURL := make(map[string]struct{}, len(nodes))
	for i := range nodes {
		node := &nodes[i]
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
		node.fetchURL = fetchURL
	}

	// insteadOf matches by prefix, so a remote URL that prefixes one of the store paths would also
	// rewrite the store URL the clone path relies on.
	for _, owner := range nodes {
		if owner.fetchURL == "" {
			continue
		}
		for _, node := range nodes {
			if strings.HasPrefix(node.moduleDir, owner.fetchURL) {
				blockers = append(blockers, fmt.Sprintf("%s (its remote URL prefixes the object store path of %s)", owner.displayPath, node.displayPath))
			}
		}
	}

	if len(blockers) > 0 {
		return nil, blockers, nil
	}

	var overrides []string
	for _, node := range nodes {
		overrides = append(overrides, "-c", "submodule."+node.name+".url="+node.moduleDir)
		if node.fetchURL != "" {
			overrides = append(overrides, "-c", "url."+node.moduleDir+".insteadOf="+node.fetchURL)
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
	worktreeGitDir, commonDir, err := resolveWorkTreeGitDirs(ctx, workTreeDir)
	if err != nil {
		return err
	}
	if !util.IsSubpathOfBasePath(filepath.Join(commonDir, "worktrees"), worktreeGitDir) {
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

// resolveWorkTreeGitDirs returns the git dir of workTreeDir and the repository's common dir, read
// from disk rather than from `git rev-parse`: an inherited GIT_DIR or GIT_COMMON_DIR redirects
// rev-parse at an unrelated repository, and both the store paths werf injects and the discard's
// os.RemoveAll are derived from these. The .git pointer of the very work tree being operated on is
// immune to that.
func resolveWorkTreeGitDirs(ctx context.Context, workTreeDir string) (string, string, error) {
	dotGitPath := filepath.Join(workTreeDir, ".git")
	info, err := os.Stat(dotGitPath)
	if err != nil {
		return "", "", fmt.Errorf("unable to access %s: %w", dotGitPath, err)
	}
	if info.IsDir() {
		return dotGitPath, dotGitPath, nil
	}

	gitDir, err := resolveDotGitFile(ctx, dotGitPath)
	if err != nil {
		return "", "", fmt.Errorf("unable to resolve dot-git file %q: %w", dotGitPath, err)
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(workTreeDir, gitDir)
	}
	gitDir = filepath.Clean(gitDir)

	commonDirPath := filepath.Join(gitDir, "commondir")
	data, err := os.ReadFile(commonDirPath)
	if errors.Is(err, fs.ErrNotExist) {
		return gitDir, gitDir, nil
	}
	if err != nil {
		return "", "", fmt.Errorf("error reading %q: %w", commonDirPath, err)
	}

	commonDir := strings.TrimSpace(string(data))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(gitDir, commonDir)
	}

	return gitDir, filepath.Clean(commonDir), nil
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
	// The store path lands in the key of `-c url.<path>.insteadOf`, which git rejects outright when
	// the path itself carries a `=` or a newline.
	if strings.ContainsAny(moduleDir, "=\n") {
		*blockers = append(*blockers, fmt.Sprintf("%s (object store path %q cannot be expressed as a git config key)", displayPath, moduleDir))
		return nil
	}
	if _, err := os.Stat(moduleDir); err != nil {
		*blockers = append(*blockers, fmt.Sprintf("%s (not initialized locally)", displayPath))
		return nil
	}

	// A partial clone can hold the commit while its blobs still live on the promisor remote, so
	// cloning from it would fail or need the network — exactly what reuse is meant to avoid. This runs
	// before any object is looked up, because GIT_NO_LAZY_FETCH is honored only from git 2.45 and werf
	// supports older git: on those the lookup itself could reach the promisor remote.
	isPartial, err := moduleStoreIsPartial(ctx, moduleDir)
	if err != nil {
		return err
	}
	if isPartial {
		*blockers = append(*blockers, fmt.Sprintf("%s (local clone is partial)", displayPath))
		return nil
	}

	// Peeling to the tree covers both the commit and the tree werf enumerates nested submodules from,
	// so a store holding the commit alone is not mistaken for a serviceable one.
	hasPinnedTree, err := gitObjectExists(ctx, moduleDir, pinnedCommit+"^{tree}")
	if err != nil {
		return err
	}
	if !hasPinnedTree {
		*blockers = append(*blockers, fmt.Sprintf("%s (not initialized locally)", displayPath))
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

	// `git config --blob` of a missing blob writes an error to stderr, which every leaf submodule
	// would otherwise produce on the happy path.
	hasGitmodules, err := gitObjectExists(ctx, moduleDir, pinnedCommit+":.gitmodules")
	if err != nil {
		return err
	}
	if !hasGitmodules {
		return nil
	}

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

// submoduleNameUnsafe rejects a repo-controlled name that cannot be carried safely by a
// `-c submodule.<name>.url` option: `=` ends the config key early, so the rest of the name becomes
// part of the URL value git uses, and a `.`/`..`/empty path component makes the store path the name
// is joined into resolve somewhere other than the name reads, with `..` climbing out of modules/
// entirely. A plain `/` is allowed: git itself defaults a submodule's name to its path, and stores it
// under the same nested modules/<name> layout the name is joined into here. A component with a
// trailing dot or space is rejected too: NTFS resolves `lib`, `lib.` and `lib ` to the same
// directory, so two such names would silently share one module store on Windows. The fixed
// `submodule.` prefix keeps such a name out of unrelated config keys. .gitmodules is
// repo-controlled and `git config -f` does not reject these itself.
func submoduleNameUnsafe(name string) bool {
	if strings.ContainsAny(name, "=\n\\") {
		return true
	}
	for _, component := range strings.Split(name, "/") {
		if component == "" || component == "." || component == ".." ||
			strings.HasSuffix(component, ".") || strings.HasSuffix(component, " ") {
			return true
		}
	}
	return false
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

// parseGitmodulesPaths runs a `git config ... -z --get-regexp ^submodule\..*\.path$` over a
// .gitmodules source (a working-tree file via -f, or a tree blob via --blob) and returns the
// submodule names and paths. A missing source or no matches (git exit code 1) yields empty slices,
// not an error.
func parseGitmodulesPaths(ctx context.Context, opts *GitCmdOptions, configArgs ...string) ([]string, []string, error) {
	args := make([]string, 0, len(configArgs)+3)
	args = append(args, configArgs...)
	args = append(args, "-z", "--get-regexp", "^submodule\\..*\\.path$")
	configCmd := NewGitCmd(ctx, opts, args...)
	if err := configCmd.Run(ctx); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("git config command failed: %w", err)
	}

	var names, paths []string
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
	// An absolute or ".." path from committed .gitmodules, which git validates in no way, makes the
	// pathspec below fail the whole command. git checks out no submodule at such a path either.
	if !filepath.IsLocal(submodulePath) {
		return "", nil
	}

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

// gitObjectExists reports whether rev resolves to an object present in repoDir, quietly: --verify
// --quiet exits 1 without writing to stderr. GIT_NO_LAZY_FETCH keeps the lookup from reaching a
// promisor remote on git 2.45+; on older git the partial-store blocker is what covers that.
func gitObjectExists(ctx context.Context, repoDir, rev string) (bool, error) {
	revParseCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoDir, Env: []string{"GIT_NO_LAZY_FETCH=1"}}, "rev-parse", "--verify", "--quiet", rev)
	if err := revParseCmd.Run(ctx); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("git rev-parse command failed: %w", err)
	}
	return true, nil
}
