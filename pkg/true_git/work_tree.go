package true_git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/lock"
)

type WithWorkTreeOptions struct {
	HasSubmodules bool
}

func WithWorkTree(gitDir, workTreeCacheDir string, commit string, opts WithWorkTreeOptions, f func(workTreeDir string) error) error {
	return withWorkTreeCacheLock(workTreeCacheDir, func() error {
		var err error

		gitDir, err = filepath.Abs(gitDir)
		if err != nil {
			return fmt.Errorf("bad git dir %s: %s", gitDir, err)
		}

		workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
		if err != nil {
			return fmt.Errorf("bad work tree cache dir %s: %s", workTreeCacheDir, err)
		}

		if opts.HasSubmodules {
			err := checkSubmoduleConstraint()
			if err != nil {
				return err
			}
		}

		workTreeDir, err := prepareWorkTree(gitDir, workTreeCacheDir, commit, opts.HasSubmodules)
		if err != nil {
			return fmt.Errorf("cannot prepare worktree: %s", err)
		}

		return f(workTreeDir)
	})
}

func withWorkTreeCacheLock(workTreeCacheDir string, f func() error) error {
	lockName := fmt.Sprintf("git_work_tree_cache %s", workTreeCacheDir)
	return lock.WithLock(lockName, lock.LockOptions{Timeout: 600 * time.Second}, f)
}

func prepareWorkTree(repoDir, workTreeCacheDir string, commit string, withSubmodules bool) (string, error) {
	workTreeDirByCommit := filepath.Join(workTreeCacheDir, commit)

	workTreeExists := true
	if _, err := os.Stat(workTreeDirByCommit); os.IsNotExist(err) {
		workTreeExists = false
	} else if err != nil {
		return "", fmt.Errorf("unable to access %s: %s", workTreeDirByCommit, err)
	}
	if workTreeExists {
		return workTreeDirByCommit, nil
	}

	tmpWorkTreeDir := filepath.Join(workTreeCacheDir, uuid.NewV4().String())

	currentWorkTreeDirLink := filepath.Join(workTreeCacheDir, "current")
	currentLinkExists := true
	currentCommit := ""
	if _, err := os.Stat(currentWorkTreeDirLink); os.IsNotExist(err) {
		currentLinkExists = false
	} else if err != nil {
		return "", fmt.Errorf("unable to access %s: %s", currentWorkTreeDirLink, err)
	}
	if currentLinkExists {
		// NOTICE: Ignore readlink and rename errors.
		// NOTICE: Work tree will be created from scratch
		// NOTICE: in the specified tmp dir in that case.
		if currentWorkTreeDir, err := os.Readlink(currentWorkTreeDirLink); err == nil {
			currentCommit = filepath.Base(currentWorkTreeDir)
			_ = os.Rename(currentWorkTreeDir, tmpWorkTreeDir)
		}

		if err := os.RemoveAll(currentWorkTreeDirLink); err != nil {
			return "", fmt.Errorf("unable to remove current work tree link %s: %s", currentWorkTreeDirLink, err)
		}
	}

	logProcessMsg := fmt.Sprintf("Switch work tree to commit %s", commit)
	if err := logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
		logboek.LogInfoF("Work tree dir: %s\n", tmpWorkTreeDir)
		if currentCommit != "" {
			logboek.LogInfoF("Current commit: %s\n", currentCommit)
		}

		return switchWorkTree(repoDir, tmpWorkTreeDir, commit, withSubmodules)
	}); err != nil {
		return "", fmt.Errorf("unable to switch work tree %s to commit %s: %s", tmpWorkTreeDir, commit, err)
	}

	if err := os.RemoveAll(workTreeDirByCommit); err != nil {
		return "", fmt.Errorf("unable to remove old dir %s: %s", workTreeDirByCommit, err)
	}
	if err := os.Rename(tmpWorkTreeDir, workTreeDirByCommit); err != nil {
		return "", fmt.Errorf("unable to rename %s to %s: %s", tmpWorkTreeDir, workTreeDirByCommit, err)
	}

	if err := os.Symlink(workTreeDirByCommit, currentWorkTreeDirLink); err != nil {
		return "", fmt.Errorf("unable to create symlink %s to %s: %s", currentWorkTreeDirLink, workTreeDirByCommit, err)
	}

	return workTreeDirByCommit, nil
}

func debugWorktreeSwitch() bool {
	return os.Getenv("WERF_TRUE_GIT_DEBUG_WORKTREE_SWITCH") == "1"
}

func switchWorkTree(repoDir, workTreeDir string, commit string, withSubmodules bool) error {
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
