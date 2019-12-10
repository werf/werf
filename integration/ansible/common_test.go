// +build integration

package ansible_test

import (
	"fmt"

	"github.com/flant/werf/integration/utils/werfexec"
)

func werfBuild(dir string, opts werfexec.CommandOptions, extraArgs ...string) error {
	return werfexec.ExecWerfCommand(dir, werfBinPath, opts, append([]string{"build", "--stages-storage", ":local"}, extraArgs...)...)
}

func werfPurge(dir string, opts werfexec.CommandOptions, extraArgs ...string) error {
	return werfexec.ExecWerfCommand(dir, werfBinPath, opts, append([]string{"stages", "purge", "--stages-storage", ":local"}, extraArgs...)...)
}

func setGitRepoState(workTreeDir, repoDir string, commitMessage string) error {
	if err := werfexec.ExecWerfCommand(".", "git", werfexec.CommandOptions{}, append([]string{"init", workTreeDir, "--separate-git-dir", repoDir})...); err != nil {
		return fmt.Errorf("unable to init git repo %s with work tree %s: %s", repoDir, workTreeDir, err)
	}
	if err := werfexec.ExecWerfCommand(".", "git", werfexec.CommandOptions{}, append([]string{"-C", workTreeDir, "add", "."})...); err != nil {
		return fmt.Errorf("unable to add work tree %s files to git repo %s: %s", workTreeDir, repoDir, err)
	}
	if err := werfexec.ExecWerfCommand(".", "git", werfexec.CommandOptions{}, append([]string{"-C", workTreeDir, "commit", "-m", commitMessage})...); err != nil {
		return fmt.Errorf("unable to commit work tree %s files to git repo %s: %s", workTreeDir, repoDir, err)
	}
	return nil
}
