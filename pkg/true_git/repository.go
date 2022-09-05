package true_git

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"

	"github.com/werf/werf/pkg/util"
)

func GitOpenWithCustomWorktreeDir(gitDir, worktreeDir string) (*git.Repository, error) {
	return git.PlainOpenWithOptions(worktreeDir, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
}

type FetchOptions struct {
	All       bool
	TagsOnly  bool
	Prune     bool
	PruneTags bool
	Unshallow bool
	RefSpecs  map[string]string
}

func IsShallowFileChangedSinceWeReadIt(err error) bool {
	return err != nil && strings.Contains(err.Error(), "shallow file has changed since we read it")
}

func Fetch(ctx context.Context, path string, options FetchOptions) error {
	commandArgs := []string{"fetch"}

	if options.Unshallow {
		commandArgs = append(commandArgs, "--unshallow")
	}

	if options.All {
		commandArgs = append(commandArgs, "--all")
	}

	if options.TagsOnly {
		commandArgs = append(commandArgs, "--tags")
	}

	if options.Prune || options.PruneTags {
		commandArgs = append(commandArgs, "--prune")

		if options.PruneTags && !gitVersion.LessThan(semver.MustParse("2.17.0")) {
			commandArgs = append(commandArgs, "--prune-tags")
		}
	}

	for remote, refSpec := range options.RefSpecs {
		commandArgs = append(commandArgs, remote, refSpec)
	}

	gitCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: path}, commandArgs...)

	return gitCmd.Run(ctx)
}

func GetLastBranchCommitSHA(ctx context.Context, repoPath, branch string) (string, error) {
	revParseCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoPath}, "rev-parse", branch)
	if err := revParseCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git rev parse branch command failed: %w", err)
	}

	return strings.TrimSpace(revParseCmd.OutBuf.String()), nil
}

func IsShallowClone(ctx context.Context, path string) (bool, error) {
	if gitVersion.LessThan(semver.MustParse("2.15.0")) {
		exist, err := util.FileExists(filepath.Join(path, ".git", "shallow"))
		if err != nil {
			return false, err
		}

		return exist, nil
	}

	checkShallowCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: path}, "rev-parse", "--is-shallow-repository")
	if err := checkShallowCmd.Run(ctx); err != nil {
		return false, fmt.Errorf("git shallow repository check command failed: %w", err)
	}

	return strings.TrimSpace(checkShallowCmd.OutBuf.String()) == "true", nil
}
