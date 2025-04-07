package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/werf/werf/v2/test/pkg/utils/liveexec"
)

func SetGitRepoState(ctx context.Context, workTreeDir, repoDir, commitMessage string) error {
	if err := liveexec.ExecCommand(ctx, ".", "git", liveexec.ExecCommandOptions{}, []string{"init", workTreeDir, "--separate-git-dir", repoDir}...); err != nil {
		return fmt.Errorf("unable to init git repo %s with work tree %s: %w", repoDir, workTreeDir, err)
	}
	if err := liveexec.ExecCommand(ctx, ".", "git", liveexec.ExecCommandOptions{}, []string{"-C", workTreeDir, "add", "."}...); err != nil {
		return fmt.Errorf("unable to add work tree %s files to git repo %s: %w", workTreeDir, repoDir, err)
	}
	if err := liveexec.ExecCommand(ctx, ".", "git", liveexec.ExecCommandOptions{}, []string{"-C", workTreeDir, "commit", "-m", commitMessage}...); err != nil {
		return fmt.Errorf("unable to commit work tree %s files to git repo %s: %w", workTreeDir, repoDir, err)
	}
	return nil
}

func GetHeadCommit(ctx context.Context, workTreeDir string) string {
	out := SucceedCommandOutputString(
		ctx,
		workTreeDir,
		"git",
		"rev-parse", "HEAD",
	)

	return strings.TrimSpace(out)
}

// LookupRepoAbsPath returns the absolute path to the git repository that contains the current directory.
// same function from true_git.go can't be used due to the cyclic dependency
func LookupRepoAbsPath(ctx context.Context) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get current directory: %w", err)
	}

	repo, err := git.PlainOpenWithOptions(currentDir, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return "", fmt.Errorf("unable to open repo: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("unable to get worktree: %w", err)
	}

	return worktree.Filesystem.Root(), nil
}
