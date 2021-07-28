package utils

import (
	"fmt"
	"strings"

	"github.com/werf/werf/integration/pkg/utils/liveexec"
)

func SetGitRepoState(workTreeDir, repoDir string, commitMessage string) error {
	if err := liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, append([]string{"init", workTreeDir, "--separate-git-dir", repoDir})...); err != nil {
		return fmt.Errorf("unable to init git repo %s with work tree %s: %s", repoDir, workTreeDir, err)
	}
	if err := liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, append([]string{"-C", workTreeDir, "add", "."})...); err != nil {
		return fmt.Errorf("unable to add work tree %s files to git repo %s: %s", workTreeDir, repoDir, err)
	}
	if err := liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, append([]string{"-C", workTreeDir, "commit", "-m", commitMessage})...); err != nil {
		return fmt.Errorf("unable to commit work tree %s files to git repo %s: %s", workTreeDir, repoDir, err)
	}
	return nil
}

func GetHeadCommit(workTreeDir string) string {
	out := SucceedCommandOutputString(
		workTreeDir,
		"git",
		"rev-parse", "HEAD",
	)

	return strings.TrimSpace(out)
}
