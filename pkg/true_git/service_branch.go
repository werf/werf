package true_git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

type SyncSourceWorktreeWithServiceBranchOptions struct {
	ServiceBranch   string
	GlobExcludeList []string
}

func SyncSourceWorktreeWithServiceBranch(ctx context.Context, gitDir, sourceWorktreeDir, worktreeCacheDir, commit string, opts SyncSourceWorktreeWithServiceBranchOptions) (string, error) {
	var resultCommit string
	if err := withWorkTreeCacheLock(ctx, worktreeCacheDir, func() error {
		var err error
		if gitDir, err = filepath.Abs(gitDir); err != nil {
			return fmt.Errorf("bad git dir %s: %w", gitDir, err)
		}

		if worktreeCacheDir, err = filepath.Abs(worktreeCacheDir); err != nil {
			return fmt.Errorf("bad work tree cache dir %s: %w", worktreeCacheDir, err)
		}

		serviceWorktreeDir, err := prepareWorkTree(ctx, gitDir, worktreeCacheDir, commit, true)
		if err != nil {
			return fmt.Errorf("unable to prepare worktree for commit %v: %w", commit, err)
		}

		currentCommitPath := filepath.Join(worktreeCacheDir, "current_commit")
		if err := os.RemoveAll(currentCommitPath); err != nil {
			return fmt.Errorf("unable to remove %s: %w", currentCommitPath, err)
		}

		resultCommit, err = syncWorktreeWithServiceWorktreeBranch(ctx, sourceWorktreeDir, serviceWorktreeDir, commit, opts.ServiceBranch, opts.GlobExcludeList)
		if err != nil {
			return fmt.Errorf("unable to sync worktree with service branch %q: %w", opts.ServiceBranch, err)
		}

		return nil
	}); err != nil {
		return "", err
	}

	return resultCommit, nil
}

func syncWorktreeWithServiceWorktreeBranch(ctx context.Context, sourceWorktreeDir, serviceWorktreeDir, sourceCommit, branchName string, globExcludeList []string) (string, error) {
	if err := prepareAndCheckoutServiceBranch(ctx, serviceWorktreeDir, sourceCommit, branchName); err != nil {
		return "", fmt.Errorf("unable to get or prepare service branch head commit: %w", err)
	}

	serviceBranchHeadCommit, err := GetLastBranchCommitSHA(ctx, serviceWorktreeDir, branchName)
	if err != nil {
		return "", fmt.Errorf("unable to get service worktree commit SHA: %w", err)
	}

	revertedChangesExist, err := revertExcludedChangesInServiceWorktreeIndex(ctx, sourceWorktreeDir, serviceWorktreeDir, sourceCommit, serviceBranchHeadCommit, globExcludeList)
	if err != nil {
		return "", fmt.Errorf("unable to revert excluded changes in service worktree index: %w", err)
	}

	newChangesExist, err := checkNewChangesInSourceWorktreeDir(ctx, sourceWorktreeDir, serviceWorktreeDir, globExcludeList)
	if err != nil {
		return "", fmt.Errorf("unable to check new changes in source worktree: %w", err)
	}

	if !revertedChangesExist && !newChangesExist {
		return serviceBranchHeadCommit, nil
	}

	if err = addNewChangesInServiceWorktreeDir(ctx, sourceWorktreeDir, serviceWorktreeDir, globExcludeList); err != nil {
		return "", fmt.Errorf("unable to add new changes in service worktree: %w", err)
	}

	newCommit, err := commitNewChangesInServiceBranch(ctx, serviceWorktreeDir, branchName)
	if err != nil {
		return "", fmt.Errorf("unable to commit new changes in service branch: %w", err)
	}

	return newCommit, nil
}

func prepareAndCheckoutServiceBranch(ctx context.Context, serviceWorktreeDir, sourceCommit, branchName string) error {
	branchListCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "branch", "--list", branchName)
	if err := branchListCmd.Run(ctx); err != nil {
		return fmt.Errorf("git branch list command failed: %w", err)
	}

	if branchListCmd.OutBuf.Len() == 0 {
		checkoutCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "checkout", "-b", branchName, sourceCommit)
		if err := checkoutCmd.Run(ctx); err != nil {
			return fmt.Errorf("git checkout command failed: %w", err)
		}

		return nil
	}

	checkoutCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "checkout", branchName)
	if err := checkoutCmd.Run(ctx); err != nil {
		return fmt.Errorf("git checkout command failed: %w", err)
	}

	isSourceCommitInServiceBranch, err := IsAncestor(ctx, sourceCommit, branchName, serviceWorktreeDir)
	if err != nil {
		return fmt.Errorf("unable to detect whether sourceCommit %q is in service branch: %w", sourceCommit, err)
	}
	if isSourceCommitInServiceBranch {
		return nil
	}

	mergeCmd := NewGitCmd(
		ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir},
		"-c", "user.email=werf@werf.io", "-c", "user.name=werf",
		"merge", "--no-verify", "--no-edit", "--no-ff", "--allow-unrelated-histories", "-s", "ours", sourceCommit,
	)
	if err = mergeCmd.Run(ctx); err != nil {
		return fmt.Errorf("git merge of source commit %q into service branch %q failed: %w\nNOTE: To continue you can remove the service branch %q with \"git branch -D %s\", but we would also ask you to report this issue to https://github.com/werf/werf/issues", sourceCommit, branchName, err, branchName, branchName)
	}

	return nil
}

func revertExcludedChangesInServiceWorktreeIndex(ctx context.Context, sourceWorktreeDir, serviceWorktreeDir, sourceCommit, serviceBranchHeadCommit string, globExcludeList []string) (bool, error) {
	if len(globExcludeList) == 0 || serviceBranchHeadCommit == sourceCommit {
		return false, nil
	}

	gitDiffArgs := []string{
		"-c", "diff.renames=false",
		"-c", "core.quotePath=false",
		"diff",
		"--binary",
		serviceBranchHeadCommit, sourceCommit,
		"--",
	}
	gitDiffArgs = append(gitDiffArgs, globExcludeList...)

	diffCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, gitDiffArgs...)
	if err := diffCmd.Run(ctx); err != nil {
		return false, fmt.Errorf("git diff command failed: %w", err)
	}

	if diffCmd.OutBuf.Len() == 0 {
		return false, nil
	}

	applyCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "apply", "--binary", "--index")
	applyCmd.Stdin = diffCmd.OutBuf
	if err := applyCmd.Run(ctx); err != nil {
		return false, fmt.Errorf("git apply command failed: %w", err)
	}

	return true, nil
}

func checkNewChangesInSourceWorktreeDir(ctx context.Context, sourceWorktreeDir, serviceWorktreeDir string, globExcludeList []string) (bool, error) {
	output, err := runGitAddCmd(ctx, sourceWorktreeDir, serviceWorktreeDir, globExcludeList, true)
	if err != nil {
		return false, err
	}

	return len(output.Bytes()) != 0, nil
}

func addNewChangesInServiceWorktreeDir(ctx context.Context, sourceWorktreeDir, serviceWorktreeDir string, globExcludeList []string) error {
	_, err := runGitAddCmd(ctx, sourceWorktreeDir, serviceWorktreeDir, globExcludeList, false)
	return err
}

func runGitAddCmd(ctx context.Context, sourceWorktreeDir, serviceWorktreeDir string, globExcludeList []string, dryRun bool) (*bytes.Buffer, error) {
	gitAddArgs := []string{
		"--work-tree",
		sourceWorktreeDir,
		"add",
	}

	if dryRun {
		gitAddArgs = append(gitAddArgs, "--dry-run", "--ignore-missing")
	}

	var pathSpecList []string
	{
		pathSpecList = append(pathSpecList, ":.")
		for _, glob := range globExcludeList {
			pathSpecList = append(pathSpecList, ":!"+glob)
		}
	}

	var pathSpecFileBuf *bytes.Buffer
	if gitVersion.LessThan(semver.MustParse("2.25.0")) {
		gitAddArgs = append(gitAddArgs, "--")
		gitAddArgs = append(gitAddArgs, pathSpecList...)
	} else {
		gitAddArgs = append(gitAddArgs, "--pathspec-from-file=-", "--pathspec-file-nul")
		pathSpecFileBuf = bytes.NewBufferString(strings.Join(pathSpecList, "\000"))
	}

	addCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, gitAddArgs...)
	if pathSpecFileBuf != nil {
		addCmd.Stdin = pathSpecFileBuf
	}
	if err := addCmd.Run(ctx); err != nil {
		return nil, err
	}

	return addCmd.OutBuf, nil
}

func commitNewChangesInServiceBranch(ctx context.Context, serviceWorktreeDir, branchName string) (string, error) {
	commitCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "-c", "user.email=werf@werf.io", "-c", "user.name=werf", "commit", "--no-verify", "-m", time.Now().String())
	if err := commitCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git commit command failed: %w", err)
	}

	revParseCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "rev-parse", branchName)
	if err := revParseCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git rev parse branch command failed: %w", err)
	}

	serviceNewCommit := strings.TrimSpace(revParseCmd.OutBuf.String())

	checkoutCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "checkout", "--force", "--detach", serviceNewCommit)
	if err := checkoutCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git checkout command failed: %w", err)
	}

	return serviceNewCommit, nil
}
