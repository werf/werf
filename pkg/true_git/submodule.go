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

		localURLOpts, remoteOnly, err := submoduleLocalURLOverrides(ctx, workTreeDir)
		if err != nil {
			return err
		}
		if len(remoteOnly) > 0 {
			logboek.Context(ctx).Warn().LogF(
				"WARNING: submodules %s are not present in the local object store and will be fetched from their remotes (credentials may be requested).\nRun `git submodule update --init --recursive` in your working tree beforehand so werf reuses the local objects and touches no network.\n",
				strings.Join(remoteOnly, ", "),
			)
		}

		updateArgs := make([]string, 0, len(includePathOpts)+len(localURLOpts)+8)
		updateArgs = append(updateArgs, includePathOpts...)
		if len(localURLOpts) > 0 {
			// The overrides point submodule URLs at the on-disk module store, which git clones over
			// the file transport — blocked by the default protocol.file.allow=user for submodules.
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

// submoduleLocalURLOverrides walks the superproject's submodules recursively and, for those whose
// pinned commit is already present in the local module store (<parent-module-dir>/modules/<name>,
// rooted at <git-common-dir>), returns `-c submodule.<name>.url=<local-path>` overrides so
// `git submodule update --recursive` clones them from those local objects instead of their remotes
// — no network, no credentials. These -c options propagate to the nested update invocations via
// GIT_CONFIG_PARAMETERS, so nested submodules are covered too. Submodules missing locally keep
// their remote URL and are returned in remoteOnly so the caller can warn they will be fetched.
//
// A submodule.<name>.url override is global to the git process, so when the same submodule name
// occurs more than once at different module dirs it cannot be expressed unambiguously; such names
// are dropped from the overrides and fall back to their remotes instead of risking a wrong source.
func submoduleLocalURLOverrides(ctx context.Context, workTreeDir string) (overrides, remoteOnly []string, err error) {
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

	var localNodes []submoduleNode
	for i, name := range names {
		pinnedCommit, err := lsTreeGitlink(ctx, workTreeDir, "HEAD", paths[i])
		if err != nil {
			return nil, nil, err
		}
		if err := walkSubmodule(ctx, commonDir, name, paths[i], pinnedCommit, &localNodes, &remoteOnly); err != nil {
			return nil, nil, err
		}
	}

	nameCount := make(map[string]int, len(localNodes))
	for _, node := range localNodes {
		nameCount[node.name]++
	}
	for _, node := range localNodes {
		if nameCount[node.name] == 1 {
			overrides = append(overrides, "-c", "submodule."+node.name+".url="+node.moduleDir)
		} else {
			remoteOnly = append(remoteOnly, node.displayPath)
		}
	}

	return overrides, remoteOnly, nil
}

// walkSubmodule resolves one submodule against the local module store and recurses into its nested
// submodules, reading their definitions from the module store at the pinned commit (no checkout
// needed). parentModuleDir is the git dir whose modules/<name> hosts this submodule's store.
func walkSubmodule(ctx context.Context, parentModuleDir, name, displayPath, pinnedCommit string, localNodes *[]submoduleNode, remoteOnly *[]string) error {
	if pinnedCommit == "" {
		return nil
	}

	moduleDir := filepath.Join(parentModuleDir, "modules", name)
	if !moduleStoreHasCommit(ctx, moduleDir, pinnedCommit) {
		*remoteOnly = append(*remoteOnly, displayPath)
		return nil
	}
	*localNodes = append(*localNodes, submoduleNode{name: name, displayPath: displayPath, moduleDir: moduleDir})

	nestedNames, nestedPaths, err := parseGitmodulesPaths(ctx, &GitCmdOptions{RepoDir: moduleDir}, "config", "--blob", pinnedCommit+":.gitmodules")
	if err != nil {
		return err
	}
	for i, nestedName := range nestedNames {
		nestedPin, err := lsTreeGitlink(ctx, moduleDir, pinnedCommit, nestedPaths[i])
		if err != nil {
			return err
		}
		if err := walkSubmodule(ctx, moduleDir, nestedName, displayPath+"/"+nestedPaths[i], nestedPin, localNodes, remoteOnly); err != nil {
			return err
		}
	}

	return nil
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
	lsTreeCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoDir}, "ls-tree", treeish, "--", submodulePath)
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
	catFileCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: moduleDir}, "cat-file", "-e", commit+"^{commit}")
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
