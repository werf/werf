package true_git

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/util/timestamps"
	"github.com/werf/werf/pkg/werf"
)

var ErrInvalidDotGit = errors.New("invalid file format: expected gitdir record")

type WithWorkTreeOptions struct {
	HasSubmodules bool
}

func WithWorkTree(ctx context.Context, gitDir, workTreeCacheDir, commit string, opts WithWorkTreeOptions, f func(workTreeDir string) error) error {
	return withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		var err error

		gitDir, err = filepath.Abs(gitDir)
		if err != nil {
			return fmt.Errorf("bad git dir %s: %w", gitDir, err)
		}

		workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
		if err != nil {
			return fmt.Errorf("bad work tree cache dir %s: %w", workTreeCacheDir, err)
		}

		workTreeDir, err := prepareWorkTree(ctx, gitDir, workTreeCacheDir, commit, opts.HasSubmodules)
		if err != nil {
			return fmt.Errorf("cannot prepare worktree: %w", err)
		}

		return f(workTreeDir)
	})
}

func withWorkTreeCacheLock(ctx context.Context, workTreeCacheDir string, f func() error) error {
	lockName := fmt.Sprintf("git_work_tree_cache %s", workTreeCacheDir)
	return werf.WithHostLock(ctx, lockName, lockgate.AcquireOptions{Timeout: 600 * time.Second}, f)
}

func prepareWorkTree(ctx context.Context, repoDir, workTreeCacheDir, commit string, withSubmodules bool) (string, error) {
	if err := os.MkdirAll(workTreeCacheDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("unable to create dir %s: %w", workTreeCacheDir, err)
	}

	lastAccessAtPath := filepath.Join(workTreeCacheDir, "last_access_at")
	if err := timestamps.WriteTimestampFile(lastAccessAtPath, time.Now()); err != nil {
		return "", fmt.Errorf("error writing timestamp file %q: %w", lastAccessAtPath, err)
	}

	gitDirPath := filepath.Join(workTreeCacheDir, "git_dir")
	if _, err := os.Stat(gitDirPath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(gitDirPath, []byte(repoDir+"\n"), 0o644); err != nil {
			return "", fmt.Errorf("error writing %s: %w", gitDirPath, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("unable to access %s: %w", gitDirPath, err)
	}

	workTreeDir := filepath.Join(workTreeCacheDir, "worktree")
	currentCommit := ""
	currentCommitPath := filepath.Join(workTreeCacheDir, "current_commit")

	_, err := os.Stat(workTreeDir)
	switch {
	case os.IsNotExist(err):

	case err != nil:
		return "", fmt.Errorf("error accessing %q: %w", workTreeDir, err)
	default:
		resolvedWorkTreeDir, err := filepath.EvalSymlinks(workTreeDir)
		if err != nil {
			return "", fmt.Errorf("unable to eval symlinks of %q: %w", workTreeDir, err)
		}

		isWorkTreeRegistered := false
		if workTreeList, err := GetWorkTreeList(ctx, repoDir); err != nil {
			return "", fmt.Errorf("unable to get worktree list for repo %s: %w", repoDir, err)
		} else {
			for _, workTreeDesc := range workTreeList {
				if filepath.ToSlash(workTreeDesc.Path) == filepath.ToSlash(resolvedWorkTreeDir) {
					isWorkTreeRegistered = true
				}
			}
		}
		if !isWorkTreeRegistered {
			logboek.Context(ctx).Default().LogFDetails("Detected unregistered work tree dir %q of repo %s\n", workTreeDir, repoDir)
		}

		isWorkTreeConsistent, err := verifyWorkTreeConsistency(ctx, repoDir, workTreeDir)
		if err != nil {
			return "", fmt.Errorf("unable to verify work tree %q consistency: %w", workTreeDir, err)
		}
		if !isWorkTreeConsistent {
			logboek.Context(ctx).Default().LogFDetails("Detected inconsistent work tree dir %q of repo %s\n", workTreeDir, repoDir)
		}

		if !isWorkTreeRegistered || !isWorkTreeConsistent {
			logboek.Context(ctx).Default().LogF("Removing invalidated work tree dir %q of repo %s\n", workTreeDir, repoDir)

			if err := os.RemoveAll(currentCommitPath); err != nil {
				return "", fmt.Errorf("unable to remove %s: %w", currentCommitPath, err)
			}

			if err := os.RemoveAll(workTreeDir); err != nil {
				return "", fmt.Errorf("unable to remove invalidated work tree dir %s: %w", workTreeDir, err)
			}
		} else {
			currentCommitPathExists := true
			if _, err := os.Stat(currentCommitPath); os.IsNotExist(err) {
				currentCommitPathExists = false
			} else if err != nil {
				return "", fmt.Errorf("unable to access %s: %w", currentCommitPath, err)
			}

			if currentCommitPathExists {
				if data, err := ioutil.ReadFile(currentCommitPath); err == nil {
					currentCommit = strings.TrimSpace(string(data))

					if currentCommit == commit {
						return workTreeDir, nil
					}
				} else {
					return "", fmt.Errorf("error reading %s: %w", currentCommitPath, err)
				}

				if err := os.RemoveAll(currentCommitPath); err != nil {
					return "", fmt.Errorf("unable to remove %s: %w", currentCommitPath, err)
				}
			}
		}
	}

	// Switch worktree state to the desired commit.
	// If worktree already exists — it will be used as a cache.
	logProcessMsg := fmt.Sprintf("Switch work tree %s to commit %s", workTreeDir, commit)
	if err := logboek.Context(ctx).Info().LogProcess(logProcessMsg).DoError(func() error {
		logboek.Context(ctx).Info().LogFDetails("Work tree dir: %s\n", workTreeDir)
		logboek.Context(ctx).Info().LogFDetails("Commit: %s\n", commit)
		if currentCommit != "" {
			logboek.Context(ctx).Info().LogFDetails("Current commit: %s\n", currentCommit)
		}

		return switchWorkTree(ctx, repoDir, workTreeDir, commit, withSubmodules)
	}); err != nil {
		return "", fmt.Errorf("unable to switch work tree %s to commit %s: %w", workTreeDir, commit, err)
	}

	if err := ioutil.WriteFile(currentCommitPath, []byte(commit+"\n"), 0o644); err != nil {
		return "", fmt.Errorf("error writing %s: %w", currentCommitPath, err)
	}

	return workTreeDir, nil
}

func verifyWorkTreeConsistency(ctx context.Context, repoDir, workTreeDir string) (bool, error) {
	resolvedGitDir, err := resolveDotGitFile(ctx, filepath.Join(workTreeDir, ".git"))
	if err != nil {
		return false, fmt.Errorf("unable to resolve dot-git file %q: %w", filepath.Join(workTreeDir, ".git"), err)
	}

	if !util.IsSubpathOfBasePath(repoDir, resolvedGitDir) {
		return false, nil
	}

	_, err = os.Stat(resolvedGitDir)
	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("error accessing resolved dot git dir %q: %w", resolvedGitDir, err)
	}

	return true, nil
}

func resolveDotGitFile(ctx context.Context, dotGitPath string) (string, error) {
	data, err := os.ReadFile(dotGitPath)
	if err != nil {
		return "", fmt.Errorf("error reading %q: %w", dotGitPath, err)
	}

	lines := util.SplitLines(string(data))
	if len(lines) == 0 {
		goto InvalidDotGit
	}

	if !strings.HasPrefix(lines[0], "gitdir: ") {
		goto InvalidDotGit
	}

	return strings.TrimSpace(strings.TrimPrefix(lines[0], "gitdir: ")), nil

InvalidDotGit:
	return "", ErrInvalidDotGit
}

func switchWorkTree(ctx context.Context, repoDir, workTreeDir, commit string, withSubmodules bool) error {
	_, err := os.Stat(workTreeDir)
	switch {
	case os.IsNotExist(err):
		wtAddCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoDir}, "worktree", "add", "--force", "--force", "--detach", workTreeDir, commit)
		if err = wtAddCmd.Run(ctx); err != nil {
			return fmt.Errorf("git worktree add command failed: %w", err)
		}
	case err != nil:
		return fmt.Errorf("error accessing %s: %w", workTreeDir, err)
	default:
		checkoutCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "checkout", "--force", "--detach", commit)
		if err = checkoutCmd.Run(ctx); err != nil {
			return fmt.Errorf("git checkout command failed: %w", err)
		}
	}

	resetCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "reset", "--hard", commit)
	if err = resetCmd.Run(ctx); err != nil {
		return fmt.Errorf("git reset command failed: %w", err)
	}

	cleanCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "--work-tree", workTreeDir, "clean", "-d", "-f", "-f", "-x")
	if err = cleanCmd.Run(ctx); err != nil {
		return fmt.Errorf("git worktree clean command failed: %w", err)
	}

	if withSubmodules {
		if err = syncSubmodules(ctx, repoDir, workTreeDir); err != nil {
			return fmt.Errorf("cannot sync submodules: %w", err)
		}

		if err = updateSubmodules(ctx, repoDir, workTreeDir); err != nil {
			return fmt.Errorf("cannot update submodules: %w", err)
		}

		submResetArgs := []string{
			"--work-tree", workTreeDir, "submodule", "foreach", "--recursive",
		}
		submResetArgs = append(submResetArgs, append([]string{"git"}, append(getCommonGitOptions(), "reset", "--hard")...)...)

		submResetCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, submResetArgs...)
		if err = submResetCmd.Run(ctx); err != nil {
			return fmt.Errorf("git submodules reset commands failed: %w", err)
		}

		submCleanArgs := []string{
			"--work-tree", workTreeDir, "submodule", "foreach", "--recursive",
		}
		submCleanArgs = append(submCleanArgs, append([]string{"git"}, append(getCommonGitOptions(), "clean", "-d", "-f", "-f", "-x")...)...)

		submCleanCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, submCleanArgs...)
		if err = submCleanCmd.Run(ctx); err != nil {
			return fmt.Errorf("git submodules clean commands failed: %w", err)
		}
	}

	return nil
}

func ResolveRepoDir(ctx context.Context, repoDir string) (string, error) {
	revParseCmd := NewGitCmd(ctx, nil, "--git-dir", repoDir, "rev-parse", "--git-dir")
	if err := revParseCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git parse git-dir command failed: %w", err)
	}

	return strings.TrimSpace(revParseCmd.OutBuf.String()), nil
}

type WorktreeDescriptor struct {
	Path   string
	Head   string
	Branch string
}

func GetWorkTreeList(ctx context.Context, repoDir string) ([]WorktreeDescriptor, error) {
	wtListCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoDir}, "worktree", "list", "--porcelain")
	if err := wtListCmd.Run(ctx); err != nil {
		return nil, fmt.Errorf("git worktree list command failed: %w", err)
	}

	var worktreeDesc *WorktreeDescriptor
	var res []WorktreeDescriptor
	for _, line := range strings.Split(wtListCmd.OutBuf.String(), "\n") {
		if line == "" && worktreeDesc == nil {
			continue
		} else if worktreeDesc == nil {
			worktreeDesc = &WorktreeDescriptor{}
		}

		switch {
		case strings.HasPrefix(line, "worktree "):
			worktreeDesc.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			worktreeDesc.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			worktreeDesc.Branch = strings.TrimPrefix(line, "branch ")
		case line == "":
			res = append(res, *worktreeDesc)
			worktreeDesc = nil
		}
	}

	return res, nil
}
