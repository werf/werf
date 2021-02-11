package true_git

import (
	"fmt"
	"os/exec"
	"strings"
)

func IsAncestor(ancestorCommit, descendantCommit string, gitDir string) (bool, error) {
	gitArgs := append(getCommonGitOptions(), "-C", gitDir, "merge-base", "--is-ancestor", ancestorCommit, descendantCommit)
	cmd := exec.Command("git", gitArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, nil
			}
			if exitError.ExitCode() == 128 && strings.HasPrefix(string(output), "fatal: Not a valid commit name ") {
				return false, nil
			}
		}
		return false, fmt.Errorf("'git merge-base' failed: %s:\n%s", err, output)
	}
	return true, nil
}
