package true_git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func SwitchWorkTree(repoDir, workTreeDir string, commit string) error {
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
