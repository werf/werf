package true_git

import (
	"fmt"
	"os/exec"
)

func deinitSubmodules(repoDir, workTreeDir string) error {
	fmt.Printf("Deinit submodules in work tree `%s` ...\n", workTreeDir)

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

	fmt.Printf("Deinit submodules in work tree `%s` OK\n", workTreeDir)

	return nil
}

func syncSubmodules(repoDir, workTreeDir string) error {
	fmt.Printf("Sync submodules in work tree `%s` ...\n", workTreeDir)

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

	fmt.Printf("Sync submodules in work tree `%s` OK\n", workTreeDir)

	return nil
}

func updateSubmodules(repoDir, workTreeDir string) error {
	fmt.Printf("Update submodules in work tree `%s` ...\n", workTreeDir)

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

	fmt.Printf("Update submodules in work tree `%s` OK\n", workTreeDir)

	return nil
}
