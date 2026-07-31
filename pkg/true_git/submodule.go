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

		updateArgs := make([]string, 0, len(includePathOpts)+len(localURLOpts)+8)
		updateArgs = append(updateArgs, includePathOpts...)
		if len(localURLOpts) > 0 {
			// The overrides point submodule URLs at the on-disk module store, which git clones over
			// the file transport — blocked by the default protocol.file.allow=user for submodules.
			// Safe only because overrides are all-or-nothing: every URL used by this invocation is
			// then a path werf computed itself, never one taken from committed .gitmodules.
			updateArgs = append(updateArgs, "-c", "protocol.file.allow=always")
			updateArgs = append(updateArgs, localURLOpts...)
		}
		updateArgs = append(updateArgs, "submodule", "update", "--checkout", "--force", "--init", "--recursive")

		submUpdateCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, updateArgs...)
		if err := submUpdateCmd.Run(ctx); err != nil {
			return fmt.Errorf("submodule update command failed: %w", err)
		}

		return nil
	})
}

type submoduleNode struct {
	name        string
	displayPath string
	moduleDir   string
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

	var nodes []submoduleNode
	for i, name := range names {
		pinnedCommit, err := lsTreeGitlink(ctx, workTreeDir, "HEAD", paths[i])
		if err != nil {
			return nil, nil, err
		}
		if err := walkSubmodule(ctx, commonDir, name, paths[i], pinnedCommit, &nodes, &blockers); err != nil {
			return nil, nil, err
		}
	}

	seen := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		if _, ok := seen[node.name]; ok {
			blockers = append(blockers, fmt.Sprintf("%s (submodule name %q is used more than once)", node.displayPath, node.name))
		}
		seen[node.name] = struct{}{}
	}

	if len(blockers) > 0 {
		return nil, blockers, nil
	}

	for _, node := range nodes {
		overrides = append(overrides, "-c", "submodule."+node.name+".url="+node.moduleDir)
	}

	return overrides, nil, nil
}

// walkSubmodule resolves one submodule against the local module store and recurses into its nested
// submodules, reading their definitions from the module store at the pinned commit (no checkout
// needed). parentModuleDir is the git dir whose modules/<name> hosts this submodule's store.
func walkSubmodule(ctx context.Context, parentModuleDir, name, displayPath, pinnedCommit string, nodes *[]submoduleNode, blockers *[]string) error {
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

	*nodes = append(*nodes, submoduleNode{name: name, displayPath: displayPath, moduleDir: moduleDir})

	nestedNames, nestedPaths, err := parseGitmodulesPaths(ctx, &GitCmdOptions{RepoDir: moduleDir}, "config", "--blob", pinnedCommit+":.gitmodules")
	if err != nil {
		return err
	}
	for i, nestedName := range nestedNames {
		nestedPin, err := lsTreeGitlink(ctx, moduleDir, pinnedCommit, nestedPaths[i])
		if err != nil {
			return err
		}
		if err := walkSubmodule(ctx, moduleDir, nestedName, displayPath+"/"+nestedPaths[i], nestedPin, nodes, blockers); err != nil {
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
