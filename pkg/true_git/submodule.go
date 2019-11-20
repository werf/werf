package true_git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/flant/logboek"
)

func syncSubmodules(repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Sync submodules in work tree '%s'", workTreeDir)
	return logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
		cmd := exec.Command(
			"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
			"submodule", "sync", "--recursive",
		)

		cmd.Dir = workTreeDir // required for `git submodule` to work

		output := setCommandRecordingLiveOutput(cmd)

		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("`git submodule sync` failed: %s\n%s", err, output.String())
		}

		return nil
	})
}

func updateSubmodules(repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Update submodules in work tree '%s'", workTreeDir)
	return logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
		cmd := exec.Command(
			"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
			"submodule", "update", "--checkout", "--force", "--init", "--recursive",
		)

		cmd.Dir = workTreeDir // required for `git submodule` to work

		output := setCommandRecordingLiveOutput(cmd)

		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("`git submodule update` failed: %s\n%s", err, output.String())
		}

		return nil
	})
}
