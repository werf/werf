package true_git

import (
	"fmt"
	"os/exec"
)

func updateSubmodules(repoDir, workTreeDir string) error {
	fmt.Printf("Update submodules in work tree `%s` ...\n", workTreeDir)

	cmd := exec.Command(
		"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
		"submodule", "update", "--init", "--recursive",
	)

	cmd.Dir = workTreeDir // required for `git submodule` to work

	output := setCommandRecordingLiveOutput(cmd)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git submodule update failed: %s\n%s", err, output.String())
	}

	fmt.Printf("Update submodules in work tree `%s` OK\n", workTreeDir)

	return nil
}
