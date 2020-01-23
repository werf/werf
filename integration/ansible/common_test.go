package ansible_test

import (
	"fmt"

	"github.com/flant/werf/pkg/testing/utils"
	"github.com/flant/werf/pkg/testing/utils/liveexec"
)

func werfBuild(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, utils.WerfBinArgs(append([]string{"build", "--stages-storage", ":local"}, extraArgs...)...)...)
}

func werfPurge(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, utils.WerfBinArgs(append([]string{"stages", "purge", "--stages-storage", ":local"}, extraArgs...)...)...)
}

func setGitRepoState(workTreeDir, repoDir string, commitMessage string) error {
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
