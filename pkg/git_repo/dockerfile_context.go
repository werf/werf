package git_repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/true_git/ls_tree"
)

// GetDockerfileContextChecksum snapshots checkout inputs per context for this repository instance.
// An empty checksum retains commit-based caching when those inputs cannot be determined safely.
func (repo *Local) GetDockerfileContextChecksum(ctx context.Context, opts ChecksumOptions) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("calculate Dockerfile context checksum: %w", err)
	}
	opts.AllFiles = true
	key := opts.ID()
	mutex, _ := repo.dockerfileContextMutex.LoadOrStore(key, &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()
	if checksum, ok := repo.dockerfileContextChecksums.Load(key); ok {
		return checksum.(string), nil
	}
	checksum, err := repo.calculateDockerfileContextChecksum(ctx, opts)
	if err != nil {
		return "", err
	}
	repo.dockerfileContextChecksums.Store(key, checksum)
	return checksum, nil
}

func (repo *Local) calculateDockerfileContextChecksum(ctx context.Context, opts ChecksumOptions) (string, error) {
	fallback := func(reason string) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("check Dockerfile context inputs: %w", err)
		}
		logboek.Context(ctx).Debug().LogF("Use commit-based Dockerfile context cache: %s\n", reason)
		return "", nil
	}
	if os.Getenv("GIT_ATTR_SOURCE") != "" {
		return fallback("Git attribute source is overridden")
	}

	var attributePaths []string
	for _, variable := range []string{"GIT_ATTR_SYSTEM", "GIT_ATTR_GLOBAL"} {
		cmd := true_git.NewGitCmd(ctx, &true_git.GitCmdOptions{RepoDir: repo.WorkTreeDir}, "var", variable)
		if err := cmd.Run(ctx); err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 || cmd.OutBuf.String() != "" || cmd.ErrBuf.String() != "" {
				return fallback(fmt.Sprintf("discover Git attribute sources: %s", err))
			}
		}
		if cmd.ErrBuf.String() != "" {
			return fallback("Git reported warnings while discovering attribute sources")
		}
		attributePaths = append(attributePaths, strings.TrimSuffix(cmd.OutBuf.String(), "\n"))
	}

	configCmd := true_git.NewGitCmd(ctx, &true_git.GitCmdOptions{RepoDir: repo.WorkTreeDir},
		"config", "--null", "--get-regexp", `^(core\.(eol|autocrlf|checkroundtripencoding|safecrlf|symlinks|ignorecase|attributesfile)$|filter\.|i18n\.|attr\.|includeif\.|extensions\.worktreeconfig$)`)
	if err := configCmd.Run(ctx); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 || configCmd.OutBuf.String() != "" || configCmd.ErrBuf.String() != "" {
			return fallback(fmt.Sprintf("check Git configuration: %s", err))
		}
	}
	if configCmd.ErrBuf.String() != "" {
		return fallback("Git reported warnings while checking configuration")
	}
	configuration := configCmd.OutBuf.String()
	hasFilters := false
	if configuration != "" {
		if !strings.HasSuffix(configuration, "\x00") {
			return fallback("invalid Git configuration response")
		}
		for _, record := range strings.Split(strings.TrimSuffix(configuration, "\x00"), "\x00") {
			name, value, hasValue := strings.Cut(record, "\n")
			if strings.HasPrefix(name, "includeif.") || name == "extensions.worktreeconfig" || name == "attr.tree" {
				return fallback("checkout configuration may differ between worktrees")
			}
			if name == "core.symlinks" && hasValue && !slices.Contains([]string{"true", "yes", "on", "1"}, strings.ToLower(value)) {
				return fallback("symlink checkout is disabled or uncertain")
			}
			if strings.HasPrefix(name, "filter.") && (strings.HasSuffix(name, ".smudge") || strings.HasSuffix(name, ".process") || strings.HasSuffix(name, ".required")) {
				hasFilters = true
			}
		}
	}
	versionCmd := true_git.NewGitCmd(ctx, &true_git.GitCmdOptions{RepoDir: repo.WorkTreeDir}, "version")
	if err := versionCmd.Run(ctx); err != nil {
		return fallback(fmt.Sprintf("read Git version: %s", err))
	}
	if versionCmd.ErrBuf.String() != "" {
		return fallback("Git reported warnings while reading its version")
	}
	inputs := []string{runtime.GOOS, versionCmd.OutBuf.String(), configuration}
	infoCmd := true_git.NewGitCmd(ctx, &true_git.GitCmdOptions{RepoDir: repo.WorkTreeDir}, "rev-parse", "--path-format=absolute", "--git-path", "info/attributes")
	if err := infoCmd.Run(ctx); err != nil {
		return fallback(fmt.Sprintf("locate repository attributes: %s", err))
	}
	if infoCmd.ErrBuf.String() != "" {
		return fallback("Git reported warnings while locating repository attributes")
	}
	attributePaths = append(attributePaths, strings.TrimSuffix(infoCmd.OutBuf.String(), "\n"))
	for _, name := range attributePaths {
		inputs = append(inputs, name)
		if name == "" {
			inputs = append(inputs, "absent")
			continue
		}
		if !filepath.IsAbs(name) {
			return fallback("relative attribute paths may differ between worktrees")
		}
		data, err := os.ReadFile(name)
		if errors.Is(err, os.ErrNotExist) {
			inputs = append(inputs, "absent")
			continue
		} else if err != nil {
			return fallback(fmt.Sprintf("read external Git attributes: %s", err))
		}
		inputs = append(inputs, util.Sha256Hash(string(data)))
	}

	repository, err := repo.PlainOpen()
	if err != nil {
		return "", fmt.Errorf("open repository for context checksum: %w", err)
	}
	commitHash, err := newHash(opts.Commit)
	if err != nil {
		return "", fmt.Errorf("parse context commit: %w", err)
	}
	commit, err := repository.CommitObject(commitHash)
	if err != nil {
		return "", fmt.Errorf("read context commit: %w", err)
	}
	hasSubmodules, err := HasSubmodulesInCommit(commit)
	if err != nil {
		return "", fmt.Errorf("check context submodules: %w", err)
	}
	if hasSubmodules {
		return fallback("submodule checkout settings require separate validation")
	}
	result, err := repo.lsTreeResult(ctx, opts.Commit, opts.LsTreeOptions)
	if err != nil {
		return "", fmt.Errorf("list context files: %w", err)
	}
	var paths []string
	selected := make(map[string]bool)
	attributeFiles := make(map[string]bool)
	if err := result.Walk(func(entry *ls_tree.LsTreeEntry) error {
		name := filepath.ToSlash(entry.FullFilepath)
		inputs = append(inputs, name, entry.Hash.String(), entry.Mode.String())
		paths = append(paths, name)
		selected[name] = true
		for dir := path.Dir(name); ; dir = path.Dir(dir) {
			attribute := path.Join(dir, ".gitattributes")
			if attributeFiles[attribute] {
				break
			}
			attributeFiles[attribute] = true
			if dir == "." {
				break
			}
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("collect context paths: %w", err)
	}
	if len(paths) == 0 {
		return "", nil
	}
	tree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("read context attribute tree: %w", err)
	}
	var attributeNames []string
	for name := range attributeFiles {
		attributeNames = append(attributeNames, name)
	}
	slices.Sort(attributeNames)
	for _, name := range attributeNames {
		entry, err := tree.FindEntry(name)
		if errors.Is(err, object.ErrEntryNotFound) || errors.Is(err, object.ErrDirectoryNotFound) {
			continue
		} else if err != nil {
			return "", fmt.Errorf("read context attribute entry %q: %w", name, err)
		}
		inputs = append(inputs, name, entry.Hash.String(), entry.Mode.String())
	}

	if hasFilters {
		attrCmd := true_git.NewGitCmd(ctx, &true_git.GitCmdOptions{RepoDir: repo.WorkTreeDir}, "check-attr", "--source="+opts.Commit, "--all", "-z", "--stdin")
		attrCmd.Stdin = strings.NewReader(strings.Join(paths, "\x00") + "\x00")
		if err := attrCmd.Run(ctx); err != nil {
			return fallback(fmt.Sprintf("check Git attributes: %s", err))
		}
		if attrCmd.ErrBuf.String() != "" {
			return fallback("Git reported warnings while checking attributes")
		}
		fields := strings.Split(attrCmd.OutBuf.String(), "\x00")
		if fields[len(fields)-1] != "" || (len(fields)-1)%3 != 0 {
			return fallback("invalid Git attribute response")
		}
		for i := 0; i < len(fields)-1; i += 3 {
			if !selected[fields[i]] || fields[i+1] == "" {
				return fallback("unexpected Git attribute response")
			}
			if fields[i+1] == "filter" {
				return fallback(fmt.Sprintf("checkout filter applies to %q", fields[i]))
			}
		}
	}
	inputs = append(inputs, result.Checksum(ctx))
	for i := range inputs {
		inputs[i] = util.Sha256Hash(inputs[i])
	}
	return util.Sha256Hash(inputs...), nil
}
