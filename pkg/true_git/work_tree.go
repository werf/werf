package true_git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flant/logboek"
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

	err = switchWorkTree(gitDir, workTreeDir, commit, withSubmodules)
	if err != nil {
		return fmt.Errorf("cannot reset work tree `%s` to commit `%s`: %s", workTreeDir, commit, err)
	}

	return nil
}

func switchWorkTree(repoDir, workTreeDir string, commit string, withSubmodules bool) error {
	logProcessMsg := fmt.Sprintf("Switch work tree %s to commit %s", workTreeDir, commit)
	return logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
		return doSwitchWorkTree(repoDir, workTreeDir, commit, withSubmodules)
	})
}

func debugWorktreeSwitch() bool {
	return os.Getenv("WERF_TRUE_GIT_DEBUG_WORKTREE_SWITCH") == "1"
}

func doSwitchWorkTree(repoDir, workTreeDir string, commit string, withSubmodules bool) error {
	var err error

	err = os.MkdirAll(workTreeDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create work tree dir %s: %s", workTreeDir, err)
	}

	var cmd *exec.Cmd
	var output *bytes.Buffer

	cmd = exec.Command(
		"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
		"reset", "--hard", commit,
	)
	output = setCommandRecordingLiveOutput(cmd)
	if debugWorktreeSwitch() {
		fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
	}
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("git reset failed: %s\n%s", err, output.String())
	}

	cmd = exec.Command(
		"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
		"clean", "-d", "-f", "-f", "-x",
	)
	output = setCommandRecordingLiveOutput(cmd)
	if debugWorktreeSwitch() {
		fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
	}
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("git clean failed: %s\n%s", err, output.String())
	}

	if withSubmodules {
		var err error

		err = syncSubmodules(repoDir, workTreeDir)
		if err != nil {
			return fmt.Errorf("cannot sync submodules: %s", err)
		}

		err = updateSubmodules(repoDir, workTreeDir)
		if err != nil {
			return fmt.Errorf("cannot update submodules: %s", err)
		}

		cmd = exec.Command(
			"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
			"submodule", "foreach", "--recursive",
			"git", "reset", "--hard",
		)
		cmd.Dir = workTreeDir // required for `git submodule` to work
		output = setCommandRecordingLiveOutput(cmd)
		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("git submodules reset failed: %s\n%s", err, output.String())
		}

		cmd = exec.Command(
			"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
			"submodule", "foreach", "--recursive",
			"git", "clean", "-d", "-f", "-f", "-x",
		)
		cmd.Dir = workTreeDir // required for `git submodule` to work
		output = setCommandRecordingLiveOutput(cmd)
		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("git submodules clean failed: %s\n%s", err, output.String())
		}
	}

	return nil
}
