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
	"github.com/flant/shluz"
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
	return shluz.WithLock(lockName, shluz.LockOptions{Timeout: 600 * time.Second}, f)
}

func checkIsWorkTreeValid(repoDir, workTreeDir, repoToCacheLinkFilePath string) (bool, error) {
	if _, err := os.Stat(repoToCacheLinkFilePath); err == nil {
		if data, err := ioutil.ReadFile(repoToCacheLinkFilePath); err != nil {
			return false, fmt.Errorf("error reading %s: %s", repoToCacheLinkFilePath, err)
		} else if strings.TrimSpace(string(data)) == workTreeDir {
			return true, nil
		}
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("error accessing %s: %s", repoToCacheLinkFilePath, err)
	}

	return false, nil
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
	workTreeDir := filepath.Join(workTreeCacheDir, "worktree")

	realRepoDir, err := GetRealRepoDir(repoDir)
	if err != nil {
		return "", fmt.Errorf("unable to get real repo dir of repo %s: %s", repoDir, err)
	}
	repoToCacheLinkFilePath := filepath.Join(realRepoDir, "werf_work_tree_cache_dir")

	currentCommit := ""

	isWorkTreeDirExist := false
	isWorkTreeValid := true
	if _, err := os.Stat(workTreeDir); err == nil {
		isWorkTreeDirExist = true
		if isValid, err := checkIsWorkTreeValid(repoDir, workTreeDir, repoToCacheLinkFilePath); err != nil {
			return "", fmt.Errorf("unable to check work tree %s validity with repo %s: %s", workTreeDir, repoDir, err)
		} else {
			isWorkTreeValid = isValid
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("error accessing %s: %s", workTreeDir, err)
	}

	currentCommitExists := true
	if _, err := os.Stat(currentCommitPath); os.IsNotExist(err) {
		currentCommitExists = false
	} else if err != nil {
		return "", fmt.Errorf("unable to access %s: %s", currentCommitPath, err)
	}
	if currentCommitExists {
		if data, err := ioutil.ReadFile(currentCommitPath); err == nil {
			currentCommit = strings.TrimSpace(string(data))

			if currentCommit == commit && isWorkTreeDirExist && isWorkTreeValid {
				return workTreeDir, nil
			}
		} else {
			return "", fmt.Errorf("error reading %s: %s", currentCommitPath, err)
		}

		if err := os.RemoveAll(currentCommitPath); err != nil {
			return "", fmt.Errorf("unable to remove %s: %s", currentCommitPath, err)
		}
	}

	if isWorkTreeDirExist && !isWorkTreeValid {
		logboek.Default.LogFDetails("Removing invalidated work tree dir %s of repo %s\n", workTreeDir, repoDir)

		if err := os.RemoveAll(workTreeDir); err != nil {
			return "", fmt.Errorf("unable to remove invalidated work tree dir %s: %s", workTreeDir, err)
		}
	}

	// Switch worktree state to the desired commit.
	// If worktree already exists — it will be used as a cache.
	logProcessMsg := fmt.Sprintf("Switch work tree %s to commit %s", workTreeDir, commit)
	if err := logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
		logboek.Default.LogFDetails("Work tree dir: %s\n", workTreeDir)
		if currentCommit != "" {
			logboek.Default.LogFDetails("Current commit: %s\n", currentCommit)
		}
		return switchWorkTree(repoDir, workTreeDir, commit, withSubmodules)
	}); err != nil {
		return "", fmt.Errorf("unable to switch work tree %s to commit %s: %s", workTreeDir, commit, err)
	}

	if err := ioutil.WriteFile(repoToCacheLinkFilePath, []byte(workTreeDir+"\n"), 0644); err != nil {
		return "", fmt.Errorf("unable to write %s: %s", repoToCacheLinkFilePath, err)
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
		"git", "-c", "core.autocrlf=false", "--git-dir", repoDir, "--work-tree", workTreeDir,
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
			"git", "-c", "core.autocrlf=false", "reset", "--hard",
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

func GetRealRepoDir(repoDir string) (string, error) {
	gitArgs := []string{"--git-dir", repoDir, "rev-parse", "--git-dir"}

	cmd := exec.Command("git", gitArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("'git --git-dir %s rev-parse --git-dir' failed: %s:\n%s", repoDir, err, output)
	}

	return strings.TrimSpace(string(output)), nil
}
