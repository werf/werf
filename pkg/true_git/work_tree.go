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

	"github.com/flant/lockgate"

	"github.com/werf/werf/pkg/werf"

	"github.com/flant/logboek"
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
	return werf.WithHostLock(lockName, lockgate.AcquireOptions{Timeout: 600 * time.Second}, f)
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

	workTreeDir := filepath.Join(workTreeCacheDir, "worktree")

	isWorkTreeDirExist := false
	if _, err := os.Stat(workTreeDir); err == nil {
		isWorkTreeDirExist = true
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("error accessing %s: %s", workTreeDir, err)
	}

	isWorkTreeRegistered := false
	if workTreeList, err := GetWorkTreeList(repoDir); err != nil {
		return "", fmt.Errorf("unable to get worktree list for repo %s: %s", repoDir, err)
	} else {
		for _, workTreeDesc := range workTreeList {
			if workTreeDesc.Path == workTreeDir {
				isWorkTreeRegistered = true
			}
		}
	}

	currentCommit := ""
	currentCommitPath := filepath.Join(workTreeCacheDir, "current_commit")
	currentCommitPathExists := true
	if _, err := os.Stat(currentCommitPath); os.IsNotExist(err) {
		currentCommitPathExists = false
	} else if err != nil {
		return "", fmt.Errorf("unable to access %s: %s", currentCommitPath, err)
	}

	if isWorkTreeDirExist && !isWorkTreeRegistered {
		logboek.Info.LogFDetails("Removing unregistered work tree dir %s of repo %s\n", workTreeDir, repoDir)

		if err := os.RemoveAll(currentCommitPath); err != nil {
			return "", fmt.Errorf("unable to remove %s: %s", currentCommitPath, err)
		}
		currentCommitPathExists = false

		if err := os.RemoveAll(workTreeDir); err != nil {
			return "", fmt.Errorf("unable to remove invalidated work tree dir %s: %s", workTreeDir, err)
		}
		isWorkTreeDirExist = false
	} else if isWorkTreeDirExist && currentCommitPathExists {
		if data, err := ioutil.ReadFile(currentCommitPath); err == nil {
			currentCommit = strings.TrimSpace(string(data))

			if currentCommit == commit {
				return workTreeDir, nil
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
	logProcessMsg := fmt.Sprintf("Switch work tree %s to commit %s", workTreeDir, commit)
	if err := logboek.Info.LogProcess(logProcessMsg, logboek.LevelLogProcessOptions{}, func() error {
		logboek.Info.LogFDetails("Work tree dir: %s\n", workTreeDir)
		logboek.Info.LogFDetails("Commit: %s\n", commit)
		if currentCommit != "" {
			logboek.Info.LogFDetails("Current commit: %s\n", currentCommit)
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
	var cmd *exec.Cmd
	var output *bytes.Buffer

	if _, err := os.Stat(workTreeDir); os.IsNotExist(err) {
		cmd = exec.Command(
			"git", "-C", repoDir,
			"worktree", "add", "--force", "--detach", "--no-checkout", workTreeDir,
		)
		output = setCommandRecordingLiveOutput(cmd)
		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("git worktree add failed: %s\n%s", err, output.String())
		}
	} else if err != nil {
		return fmt.Errorf("error accessing %s: %s", workTreeDir, err)
	}

	cmd = exec.Command(
		"git", "-c", "core.autocrlf=false",
		"reset", "--hard", commit,
	)
	cmd.Dir = workTreeDir
	output = setCommandRecordingLiveOutput(cmd)
	if debugWorktreeSwitch() {
		fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
	}
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("git reset failed: %s\n%s", err, output.String())
	}

	cmd = exec.Command(
		"git", "--work-tree", workTreeDir,
		"clean", "-d", "-f", "-f", "-x",
	)
	cmd.Dir = workTreeDir
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
			"git", "--work-tree", workTreeDir,
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
			"git", "--work-tree", workTreeDir,
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
		return "", fmt.Errorf("'%s' failed: %s:\n%s", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "), repoDir, err, output)
	}

	return strings.TrimSpace(string(output)), nil
}

type WorktreeDescriptor struct {
	Path   string
	Head   string
	Branch string
}

func GetWorkTreeList(repoDir string) ([]WorktreeDescriptor, error) {
	gitArgs := []string{"-C", repoDir, "worktree", "list", "--porcelain"}

	cmd := exec.Command("git", gitArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("'%s' failed: %s:\n%s", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "), repoDir, err, output)
	}

	var worktreeDesc *WorktreeDescriptor
	var res []WorktreeDescriptor
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" && worktreeDesc == nil {
			continue
		} else if worktreeDesc == nil {
			worktreeDesc = &WorktreeDescriptor{}
		}

		if strings.HasPrefix(line, "worktree ") {
			worktreeDesc.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "HEAD ") {
			worktreeDesc.Head = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") {
			worktreeDesc.Branch = strings.TrimPrefix(line, "branch ")
		} else if line == "" {
			res = append(res, *worktreeDesc)
			worktreeDesc = nil
		}
	}

	return res, nil
}
