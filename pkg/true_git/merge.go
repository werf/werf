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
)

type CreateDetachedMergeCommitOptions struct {
	HasSubmodules bool
}

func CreateDetachedMergeCommit(ctx context.Context, gitDir, workTreeCacheDir, commitToMerge, mergeIntoCommit string, opts CreateDetachedMergeCommitOptions) (string, error) {
	var resCommit string

	if err := withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		var err error

		gitDir, err = filepath.Abs(gitDir)
		if err != nil {
			return fmt.Errorf("bad git dir %s: %s", gitDir, err)
		}

		workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
		if err != nil {
			return fmt.Errorf("bad work tree cache dir %s: %s", workTreeCacheDir, err)
		}

		if workTreeDir, err := prepareWorkTree(ctx, gitDir, workTreeCacheDir, mergeIntoCommit, opts.HasSubmodules); err != nil {
			return fmt.Errorf("unable to prepare worktree for commit %v: %s", mergeIntoCommit, err)
		} else {
			var err error
			var cmd *exec.Cmd
			var output *bytes.Buffer

			currentCommitPath := filepath.Join(workTreeCacheDir, "current_commit")
			if err := os.RemoveAll(currentCommitPath); err != nil {
				return fmt.Errorf("unable to remove %s: %s", currentCommitPath, err)
			}

			cmd = exec.Command("git", append(getCommonGitOptions(), "-c", "user.email=werf@werf.io", "-c", "user.name=werf", "merge", "--no-edit", "--no-ff", commitToMerge)...)
			cmd.Dir = workTreeDir
			output = SetCommandRecordingLiveOutput(ctx, cmd)
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("git merge %q failed: %s\n%s", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "), err, output.String())
			}
			if debugMerge() {
				fmt.Printf("[DEBUG MERGE] %s\n%s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "), output)
			}

			cmd = exec.Command("git", append(getCommonGitOptions(), "rev-parse", "HEAD")...)
			cmd.Dir = workTreeDir
			output = SetCommandRecordingLiveOutput(ctx, cmd)
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("git merge failed: %s\n%s", err, output.String())
			}
			if debugMerge() {
				fmt.Printf("[DEBUG MERGE] %s\n%s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "), output)
			}

			resCommit = strings.TrimSpace(output.String())

			if err := ioutil.WriteFile(currentCommitPath, []byte(resCommit+"\n"), 0644); err != nil {
				return fmt.Errorf("unable to write %s: %s", currentCommitPath, err)
			}
		}

		return nil
	}); err != nil {
		return "", err
	}

	return resCommit, nil
}

func debugMerge() bool {
	return os.Getenv("WERF_TRUE_GIT_MERGE_DEBUG") == "1"
}
