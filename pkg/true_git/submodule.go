package true_git

import (
	"fmt"
	"os/exec"

	"github.com/flant/logboek"
)

func deinitSubmodules(repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Deinit submodules in work tree '%s'", workTreeDir)
	return logboek.LogProcessInline(logProcessMsg, logboek.LogProcessInlineOptions{}, func() error {
		cmd := exec.Command(
			"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
			"submodule", "deinit", "--all", "--force",
		)

		cmd.Dir = workTreeDir // required for `git submodule` to work

		output := setCommandRecordingLiveOutput(cmd)

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("`git submodule deinit` failed: %s\n%s", err, output.String())
		}

		return nil
	})
}

func syncSubmodules(repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Sync submodules in work tree '%s'", workTreeDir)
	return logboek.LogProcessInline(logProcessMsg, logboek.LogProcessInlineOptions{}, func() error {
		cmd := exec.Command(
			"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
			"submodule", "sync", "--recursive",
		)

		cmd.Dir = workTreeDir // required for `git submodule` to work

		output := setCommandRecordingLiveOutput(cmd)

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("`git submodule sync` failed: %s\n%s", err, output.String())
		}

		return nil
	})
}

func updateSubmodules(repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Update submodules in work tree '%s'", workTreeDir)
	return logboek.LogProcessInline(logProcessMsg, logboek.LogProcessInlineOptions{}, func() error {
		cmd := exec.Command(
			"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
			"submodule", "update", "--checkout", "--force", "--init", "--recursive",
		)

		cmd.Dir = workTreeDir // required for `git submodule` to work

		output := setCommandRecordingLiveOutput(cmd)

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("`git submodule update` failed: %s\n%s", err, output.String())
		}

		return nil
	})
}
