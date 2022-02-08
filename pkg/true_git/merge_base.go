package true_git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func IsAncestor(ctx context.Context, ancestorCommit, descendantCommit string, gitDir string) (bool, error) {
	mergeBaseCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: gitDir}, "merge-base", "--is-ancestor", ancestorCommit, descendantCommit)
	if err := mergeBaseCmd.Run(ctx); err != nil {
		var errExit *exec.ExitError
		if errors.As(err, &errExit) {
			if errExit.ExitCode() == 1 {
				return false, nil
			}
			if errExit.ExitCode() == 128 && strings.HasPrefix(mergeBaseCmd.ErrBuf.String(), "fatal: Not a valid commit name ") {
				return false, nil
			}
		}

		return false, fmt.Errorf("git merge-base command failed: %s", err)
	}

	return true, nil
}
