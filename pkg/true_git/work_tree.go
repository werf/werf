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

	err = switchWorkTree(gitDir, workTreeDir, commit, withSubmodules)
	if err != nil {
		return fmt.Errorf("cannot reset work tree `%s` to commit `%s`: %s", workTreeDir, commit, err)
	}

	if withSubmodules {
		var err error

		err = syncSubmodules(gitDir, workTreeDir)
		if err != nil {
			return fmt.Errorf("cannot sync submodules: %s", err)
		}

		err = updateSubmodules(gitDir, workTreeDir)
		if err != nil {
			return fmt.Errorf("cannot update submodules: %s", err)
		}
	}

	return nil
}

func switchWorkTree(repoDir, workTreeDir string, commit string, withSubmodules bool) error {
	var err error

	err = os.RemoveAll(workTreeDir)
	if err != nil {
		return fmt.Errorf("unable to remove old work tree dir %s: %s", workTreeDir, err)
	}

	err = os.MkdirAll(workTreeDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create work tree dir %s: %s", workTreeDir, err)
	}

	// TODO: to allow caching of work trees we need a way to switch worktrees safely

	//// NOTE: git clean -dffx does not clean .git files: so delete files manually
	//servicePathsToRemove := []string{}
	//err = filepath.Walk(workTreeDir, func(path string, info os.FileInfo, pathErr error) error {
	//	if pathErr != nil {
	//		return fmt.Errorf("error accessing path %s: %s", path, pathErr)
	//	}
	//
	//	if info.Name() == ".git" {
	//		servicePathsToRemove = append(servicePathsToRemove, path)
	//	}
	//
	//	return nil
	//})
	//if err != nil {
	//	return fmt.Errorf("error walking the path %s: %s", workTreeDir, err)
	//}
	//
	//for _, path := range servicePathsToRemove {
	//	logboek.LogF("Removing old service file %s\n", path)
	//	if err := os.RemoveAll(path); err != nil {
	//		fmt.Errorf("error removing %s: %s", path, err)
	//	}
	//}

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

	//if withSubmodules {
	//	cmd = exec.Command(
	//		"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
	//		"submodule", "foreach", "--recursive",
	//		"git", "reset", "--hard", commit,
	//	)
	//
	//	cmd.Dir = workTreeDir // required for `git submodule` to work
	//
	//	output = setCommandRecordingLiveOutput(cmd)
	//	err = cmd.Run()
	//	if err != nil {
	//		return fmt.Errorf("git submodules reset failed: %s\n%s", err, output.String())
	//	}
	//}
	//
	//cmd = exec.Command(
	//	"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
	//	"clean", "-d", "-f", "-f", "-x",
	//)
	//output = setCommandRecordingLiveOutput(cmd)
	//err = cmd.Run()
	//if err != nil {
	//	return fmt.Errorf("git clean failed: %s\n%s", err, output.String())
	//}
	//
	//if withSubmodules {
	//	cmd = exec.Command(
	//		"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
	//		"submodule", "foreach", "--recursive",
	//		"git", "clean", "-d", "-f", "-f", "-x",
	//	)
	//
	//	cmd.Dir = workTreeDir // required for `git submodule` to work
	//
	//	output = setCommandRecordingLiveOutput(cmd)
	//	err = cmd.Run()
	//	if err != nil {
	//		return fmt.Errorf("git submodules clean failed: %s\n%s", err, output.String())
	//	}
	//}

	return nil
}
