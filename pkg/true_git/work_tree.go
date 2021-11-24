package true_git

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/util/timestamps"
	"github.com/werf/werf/pkg/werf"
)

type WithWorkTreeOptions struct {
	HasSubmodules bool
}

func WithWorkTree(ctx context.Context, gitDir, workTreeCacheDir string, commit string, opts WithWorkTreeOptions, f func(workTreeDir string) error) error {
	return withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		var err error

		gitDir, err = filepath.Abs(gitDir)
		if err != nil {
			return fmt.Errorf("bad git dir %s: %s", gitDir, err)
		}

		workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
		if err != nil {
			return fmt.Errorf("bad work tree cache dir %s: %s", workTreeCacheDir, err)
		}

		workTreeDir, err := prepareWorkTree(ctx, gitDir, workTreeCacheDir, commit, opts.HasSubmodules)
		if err != nil {
			return fmt.Errorf("cannot prepare worktree: %s", err)
		}

		return f(workTreeDir)
	})
}

func withWorkTreeCacheLock(ctx context.Context, workTreeCacheDir string, f func() error) error {
	lockName := fmt.Sprintf("git_work_tree_cache %s", workTreeCacheDir)
	return werf.WithHostLock(ctx, lockName, lockgate.AcquireOptions{Timeout: 600 * time.Second}, f)
}

func prepareWorkTree(ctx context.Context, repoDir, workTreeCacheDir string, commit string, withSubmodules bool) (string, error) {
	if err := os.MkdirAll(workTreeCacheDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("unable to create dir %s: %s", workTreeCacheDir, err)
	}

	lastAccessAtPath := filepath.Join(workTreeCacheDir, "last_access_at")
	if err := timestamps.WriteTimestampFile(lastAccessAtPath, time.Now()); err != nil {
		return "", fmt.Errorf("error writing timestamp file %q: %s", lastAccessAtPath, err)
	}

	gitDirPath := filepath.Join(workTreeCacheDir, "git_dir")
	if _, err := os.Stat(gitDirPath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(gitDirPath, []byte(repoDir+"\n"), 0o644); err != nil {
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
			if filepath.ToSlash(workTreeDesc.Path) == filepath.ToSlash(workTreeDir) {
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
		logboek.Context(ctx).Info().LogFDetails("Removing unregistered work tree dir %s of repo %s\n", workTreeDir, repoDir)

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
	if err := logboek.Context(ctx).Info().LogProcess(logProcessMsg).DoError(func() error {
		logboek.Context(ctx).Info().LogFDetails("Work tree dir: %s\n", workTreeDir)
		logboek.Context(ctx).Info().LogFDetails("Commit: %s\n", commit)
		if currentCommit != "" {
			logboek.Context(ctx).Info().LogFDetails("Current commit: %s\n", currentCommit)
		}
		return switchWorkTree(ctx, repoDir, workTreeDir, commit, withSubmodules)
	}); err != nil {
		return "", fmt.Errorf("unable to switch work tree %s to commit %s: %s", workTreeDir, commit, err)
	}

	if err := ioutil.WriteFile(currentCommitPath, []byte(commit+"\n"), 0o644); err != nil {
		return "", fmt.Errorf("error writing %s: %s", currentCommitPath, err)
	}

	return workTreeDir, nil
}

func debugWorktreeSwitch() bool {
	return os.Getenv("WERF_TRUE_GIT_DEBUG_WORKTREE_SWITCH") == "1"
}

func switchWorkTree(ctx context.Context, repoDir, workTreeDir string, commit string, withSubmodules bool) error {
	var cmd *exec.Cmd
	var output *bytes.Buffer

	_, err := os.Stat(workTreeDir)
	switch {
	case os.IsNotExist(err):
		cmd = exec.Command(
			"git", append(getCommonGitOptions(), "-C", repoDir,
				"worktree", "add", "--force", "--detach", workTreeDir, commit)...,
		)

		output = SetCommandRecordingLiveOutput(ctx, cmd)
		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}

		if err = cmd.Run(); err != nil {
			return fmt.Errorf("git worktree add failed: %s\n%s", err, output.String())
		}
	case err != nil:
		return fmt.Errorf("error accessing %s: %s", workTreeDir, err)
	default:
		cmd = exec.Command("git", append(getCommonGitOptions(), "checkout", "--force", "--detach", commit)...)
		cmd.Dir = workTreeDir

		output = SetCommandRecordingLiveOutput(ctx, cmd)
		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}

		if err = cmd.Run(); err != nil {
			return fmt.Errorf("git checkout failed: %s\n%s", err, output.String())
		}
	}

	cmd = exec.Command("git", append(getCommonGitOptions(), "reset", "--hard", commit)...)
	cmd.Dir = workTreeDir

	output = SetCommandRecordingLiveOutput(ctx, cmd)
	if debugWorktreeSwitch() {
		fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
	}

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("git reset failed: %s\n%s", err, output.String())
	}

	cmd = exec.Command(
		"git", append(getCommonGitOptions(), "--work-tree", workTreeDir,
			"clean", "-d", "-f", "-f", "-x")...,
	)
	cmd.Dir = workTreeDir

	output = SetCommandRecordingLiveOutput(ctx, cmd)
	if debugWorktreeSwitch() {
		fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
	}

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("git clean failed: %s\n%s", err, output.String())
	}

	if withSubmodules {
		if err = syncSubmodules(ctx, repoDir, workTreeDir); err != nil {
			return fmt.Errorf("cannot sync submodules: %s", err)
		}

		if err = updateSubmodules(ctx, repoDir, workTreeDir); err != nil {
			return fmt.Errorf("cannot update submodules: %s", err)
		}

		gitArgs := append(getCommonGitOptions(), "--work-tree", workTreeDir, "submodule", "foreach", "--recursive")
		gitArgs = append(append(gitArgs, "git"), append(getCommonGitOptions(), "reset", "--hard")...)

		cmd = exec.Command("git", gitArgs...)
		cmd.Dir = workTreeDir // required for `git submodule` to work

		output = SetCommandRecordingLiveOutput(ctx, cmd)
		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}

		if err = cmd.Run(); err != nil {
			return fmt.Errorf("git submodules reset failed: %s\n%s", err, output.String())
		}

		gitArgs = append(getCommonGitOptions(), "--work-tree", workTreeDir, "submodule", "foreach", "--recursive")
		gitArgs = append(append(gitArgs, "git"), append(getCommonGitOptions(), "clean", "-d", "-f", "-f", "-x")...)

		cmd = exec.Command("git", gitArgs...)
		cmd.Dir = workTreeDir // required for `git submodule` to work

		output = SetCommandRecordingLiveOutput(ctx, cmd)
		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}

		if err = cmd.Run(); err != nil {
			return fmt.Errorf("git submodules clean failed: %s\n%s", err, output.String())
		}
	}

	return nil
}

func ResolveRepoDir(repoDir string) (string, error) {
	gitArgs := append(getCommonGitOptions(), "--git-dir", repoDir, "rev-parse", "--git-dir")

	cmd := exec.Command("git", gitArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%q failed (%s): %s:\n%s", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "), repoDir, err, output)
	}

	return strings.TrimSpace(string(output)), nil
}

type WorktreeDescriptor struct {
	Path   string
	Head   string
	Branch string
}

func GetWorkTreeList(repoDir string) ([]WorktreeDescriptor, error) {
	gitArgs := append(getCommonGitOptions(), "-C", repoDir, "worktree", "list", "--porcelain")
	cmd := exec.Command("git", gitArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%q failed (%s): %s:\n%s", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "), repoDir, err, output)
	}

	var worktreeDesc *WorktreeDescriptor
	var res []WorktreeDescriptor
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" && worktreeDesc == nil {
			continue
		} else if worktreeDesc == nil {
			worktreeDesc = &WorktreeDescriptor{}
		}

		switch {
		case strings.HasPrefix(line, "worktree "):
			worktreeDesc.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			worktreeDesc.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			worktreeDesc.Branch = strings.TrimPrefix(line, "branch ")
		case line == "":
			res = append(res, *worktreeDesc)
			worktreeDesc = nil
		}
	}

	return res, nil
}
