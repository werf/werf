package true_git

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type CreateDetachedMergeCommitOptions struct {
	HasSubmodules bool
}

func CreateDetachedMergeCommit(ctx context.Context, gitDir, workTreeCacheDir, commitToMerge, mergeIntoCommit string, opts CreateDetachedMergeCommitOptions) (string, error) {
	var resCommit string

	if err := withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		var err error

		gitDir, err = filepath.Abs(gitDir)
		if err != nil {
			return fmt.Errorf("bad git dir %s: %w", gitDir, err)
		}

		workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
		if err != nil {
			return fmt.Errorf("bad work tree cache dir %s: %w", workTreeCacheDir, err)
		}

		if workTreeDir, err := prepareWorkTree(ctx, gitDir, workTreeCacheDir, mergeIntoCommit, opts.HasSubmodules); err != nil {
			return fmt.Errorf("unable to prepare worktree for commit %v: %w", mergeIntoCommit, err)
		} else {
			currentCommitPath := filepath.Join(workTreeCacheDir, "current_commit")
			if err := os.RemoveAll(currentCommitPath); err != nil {
				return fmt.Errorf("unable to remove %s: %w", currentCommitPath, err)
			}

			mergeCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "-c", "user.email=werf@werf.io", "-c", "user.name=werf", "merge", "--no-verify", "--no-edit", "--no-ff", commitToMerge)
			if err := mergeCmd.Run(ctx); err != nil {
				return fmt.Errorf("git merge failed: %w", err)
			}

			getHeadCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "rev-parse", "HEAD")
			if err := getHeadCmd.Run(ctx); err != nil {
				return fmt.Errorf("getting HEAD rev during git merge failed: %w", err)
			}

			resCommit = strings.TrimSpace(getHeadCmd.OutBuf.String())

			if err := ioutil.WriteFile(currentCommitPath, []byte(resCommit+"\n"), 0o644); err != nil {
				return fmt.Errorf("unable to write %s: %w", currentCommitPath, err)
			}
		}

		return nil
	}); err != nil {
		return "", err
	}

	return resCommit, nil
}
