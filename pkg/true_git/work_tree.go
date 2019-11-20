package true_git

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	if err := os.MkdirAll(workTreeCacheDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("unable to create dir %s: %s", workTreeCacheDir, err)
	}

	gitDirPath := filepath.Join(workTreeCacheDir, "git_dir")
	if _, err := os.Stat(gitDirPath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(gitDirPath, []byte(repoDir+"\n"), 0644); err != nil {
			return "", fmt.Errorf("error writing %s: %s", gitDirPath, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("unable to access %s: %s", gitDirPath, err)
	}

	currentCommitPath := filepath.Join(workTreeCacheDir, "current_commit")
	worktreePath := filepath.Join(workTreeCacheDir, "worktree")
	currentCommit := ""

	currentCommitExists := true
	if _, err := os.Stat(currentCommitPath); os.IsNotExist(err) {
		currentCommitExists = false
	} else if err != nil {
		return "", fmt.Errorf("unable to access %s: %s", currentCommitPath, err)
	}
	if currentCommitExists {
		if data, err := ioutil.ReadFile(currentCommitPath); err == nil {
			currentCommit = strings.TrimSpace(string(data))

			if currentCommit == commit {
				return worktreePath, nil
			}
		} else {
			return "", fmt.Errorf("error reading %s: %s", currentCommitPath, err)
		}

		if err := os.RemoveAll(currentCommitPath); err != nil {
			return "", fmt.Errorf("unable to remove %s: %s", currentCommitPath, err)
		}
	}

	// Switch worktree state to the desired commit.
	// If worktree already exists — it will be used as a cache.
	workTreeDir := filepath.Join(workTreeCacheDir, "worktree")
	logProcessMsg := fmt.Sprintf("Switch work tree %s to commit %s", workTreeDir, commit)
	if err := logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
		logboek.LogInfoF("Work tree dir: %s\n", workTreeDir)
		if currentCommit != "" {
			logboek.LogInfoF("Current commit: %s\n", currentCommit)
		}

		return switchWorkTree(repoDir, workTreeDir, commit, withSubmodules)
	}); err != nil {
		return "", fmt.Errorf("unable to switch work tree %s to commit %s: %s", workTreeDir, commit, err)
	}

	if err := ioutil.WriteFile(currentCommitPath, []byte(commit+"\n"), 0644); err != nil {
		return "", fmt.Errorf("error writing %s: %s", currentCommitPath, err)
	}

	return workTreeDir, nil
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
