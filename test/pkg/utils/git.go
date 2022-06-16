package utils

import (
	"fmt"
	"strings"

	"github.com/werf/werf/test/pkg/utils/liveexec"
)

func SetGitRepoState(workTreeDir, repoDir, commitMessage string) error {
	if err := liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, []string{"init", workTreeDir, "--separate-git-dir", repoDir}...); err != nil {
		return fmt.Errorf("unable to init git repo %s with work tree %s: %w", repoDir, workTreeDir, err)
	}
	if err := liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, []string{"-C", workTreeDir, "add", "."}...); err != nil {
		return fmt.Errorf("unable to add work tree %s files to git repo %s: %w", workTreeDir, repoDir, err)
	}
	if err := liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, []string{"-C", workTreeDir, "commit", "-m", commitMessage}...); err != nil {
		return fmt.Errorf("unable to commit work tree %s files to git repo %s: %w", workTreeDir, repoDir, err)
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
