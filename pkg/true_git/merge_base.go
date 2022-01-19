package true_git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func IsAncestor(ctx context.Context, ancestorCommit, descendantCommit string, gitDir string) (bool, error) {
	mergeBaseCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: gitDir}, "merge-base", "--is-ancestor", ancestorCommit, descendantCommit)
	err := mergeBaseCmd.Run(ctx)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, nil
			}
			if exitError.ExitCode() == 128 && strings.HasPrefix(mergeBaseCmd.ErrBuf.String(), "fatal: Not a valid commit name ") {
				return false, nil
			}
		}

		return false, fmt.Errorf("git merge-base command failed: %s", err)
	}

	return true, nil
}
