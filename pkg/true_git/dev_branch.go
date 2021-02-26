package true_git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func SyncDevBranchWithStagedFiles(ctx context.Context, gitDir, workTreeCacheDir, commit string) (string, error) {
	var resCommit string

	if err := withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		var err error
		if gitDir, err = filepath.Abs(gitDir); err != nil {
			return fmt.Errorf("bad git dir %s: %s", gitDir, err)
		}

		if workTreeCacheDir, err = filepath.Abs(workTreeCacheDir); err != nil {
			return fmt.Errorf("bad work tree cache dir %s: %s", workTreeCacheDir, err)
		}

		workTreeDir, err := prepareWorkTree(ctx, gitDir, workTreeCacheDir, commit, true)
		if err != nil {
			return fmt.Errorf("unable to prepare worktree for commit %v: %s", commit, err)
		}

		currentCommitPath := filepath.Join(workTreeCacheDir, "current_commit")
		if err := os.RemoveAll(currentCommitPath); err != nil {
			return fmt.Errorf("unable to remove %s: %s", currentCommitPath, err)
		}

		devBranchName := fmt.Sprintf("werf-dev-%s", commit)
		var isDevBranchExist bool
		if output, err := runGitCmd(ctx, []string{"branch", "--list", devBranchName}, workTreeDir, runGitCmdOptions{}); err != nil {
			return err
		} else {
			isDevBranchExist = output.Len() != 0
		}

		var devHeadCommit string
		if isDevBranchExist {
			if _, err := runGitCmd(ctx, []string{"checkout", devBranchName}, workTreeDir, runGitCmdOptions{}); err != nil {
				return err
			}

			if output, err := runGitCmd(ctx, []string{"rev-parse", devBranchName}, workTreeDir, runGitCmdOptions{}); err != nil {
				return err
			} else {
				devHeadCommit = strings.TrimSpace(output.String())
			}
		} else {
			if _, err := runGitCmd(ctx, []string{"checkout", "-b", devBranchName, commit}, workTreeDir, runGitCmdOptions{}); err != nil {
				return err
			}

			devHeadCommit = commit
		}

		gitDiffArgs := []string{
			"-c", "diff.renames=false",
			"-c", "core.quotePath=false",
			"diff",
			"--full-index",
			"--binary",
			"--cached",
			devHeadCommit,
		}
		if diffOutput, err := runGitCmd(ctx, gitDiffArgs, gitDir, runGitCmdOptions{}); err != nil {
			return err
		} else if len(diffOutput.Bytes()) == 0 {
			resCommit = devHeadCommit
		} else {
			if _, err := runGitCmd(ctx, []string{"apply", "--binary", "--index"}, workTreeDir, runGitCmdOptions{stdin: diffOutput}); err != nil {
				return err
			}

			gitArgs := []string{"-c", "user.email=werf@werf.io", "-c", "user.name=werf", "commit", "-m", time.Now().String()}
			if _, err := runGitCmd(ctx, gitArgs, workTreeDir, runGitCmdOptions{}); err != nil {
				return err
			}

			if output, err := runGitCmd(ctx, []string{"rev-parse", devBranchName}, workTreeDir, runGitCmdOptions{}); err != nil {
				return err
			} else {
				newDevCommit := strings.TrimSpace(output.String())
				resCommit = newDevCommit
			}
		}

		if _, err := runGitCmd(ctx, []string{"checkout", "--force", "--detach", resCommit}, workTreeDir, runGitCmdOptions{}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	return resCommit, nil
}

type runGitCmdOptions struct {
	stdin io.Reader
}

func runGitCmd(ctx context.Context, args []string, dir string, opts runGitCmdOptions) (*bytes.Buffer, error) {
	allArgs := append(getCommonGitOptions(), args...)
	cmd := exec.Command("git", allArgs...)
	cmd.Dir = dir

	if opts.stdin != nil {
		cmd.Stdin = opts.stdin
	}

	output := setCommandRecordingLiveOutput(ctx, cmd)

	err := cmd.Run()

	cmdWithArgs := strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " ")
	if debug() {
		fmt.Printf("[DEBUG] %s\n%s\n", cmdWithArgs, output)
	}

	if err != nil {
		return nil, fmt.Errorf("git command %s failed: %s\n%s", cmdWithArgs, err, output)
	}

	return output, err
}

func debug() bool {
	return os.Getenv("WERF_DEBUG_TRUE_GIT") == "1"
}
