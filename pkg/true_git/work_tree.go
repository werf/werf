package true_git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func PrepareWorkTree(gitDir, workTreeDir string, commit string) error {
	return prepareWorkTree(gitDir, workTreeDir, commit, false)
}

func PrepareWorkTreeWithSubmodules(gitDir, workTreeDir string, commit string) error {
	return prepareWorkTree(gitDir, workTreeDir, commit, true)
}

func prepareWorkTree(gitDir, workTreeDir string, commit string, withSubmodules bool) error {
	var err error

	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return fmt.Errorf("bad git dir `%s`: %s", gitDir, err)
	}

	workTreeDir, err = filepath.Abs(workTreeDir)
	if err != nil {
		return fmt.Errorf("bad work tree dir `%s`: %s", workTreeDir, err)
	}

	if withSubmodules {
		err := checkSubmoduleConstraint()
		if err != nil {
			return err
		}
	}

	err = switchWorkTree(gitDir, workTreeDir, commit)
	if err != nil {
		return fmt.Errorf("cannot reset work tree `%s` to commit `%s`: %s", workTreeDir, commit, err)
	}

	if withSubmodules {
		var err error

		err = deinitSubmodules(gitDir, workTreeDir)
		if err != nil {
			return fmt.Errorf("cannot deinit submodules: %s", err)
		}

		err = updateSubmodules(gitDir, workTreeDir)
		if err != nil {
			return fmt.Errorf("cannot update submodules: %s", err)
		}
	}

	return nil
}

func switchWorkTree(repoDir, workTreeDir string, commit string) error {
	fmt.Printf("Switch work tree `%s` to commit `%s` ...\n", workTreeDir, commit)

	var err error

	err = os.MkdirAll(workTreeDir, os.ModePerm)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	var output *bytes.Buffer

	cmd = exec.Command(
		"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
		"reset", "--hard", commit,
	)
	output = setCommandRecordingLiveOutput(cmd)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("git reset failed: %s\n%s", err, output.String())
	}

	cmd = exec.Command(
		"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
		"clean", "-d", "-f", "-f", "-x",
	)
	output = setCommandRecordingLiveOutput(cmd)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("git clean failed: %s\n%s", err, output.String())
	}

	fmt.Printf("Switch work tree `%s` to commit `%s` OK\n", workTreeDir, commit)

	return nil
}
