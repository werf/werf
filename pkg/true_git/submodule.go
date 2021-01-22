package true_git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/werf/logboek"
)

func syncSubmodules(ctx context.Context, repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Sync submodules in work tree %q", workTreeDir)
	return logboek.Context(ctx).Info().LogProcess(logProcessMsg).DoError(func() error {
		cmd := exec.Command(
			"git", "-c", "core.autocrlf=false",
			"submodule", "sync", "--recursive",
		)

		cmd.Dir = workTreeDir // required for `git submodule` to work

		output := setCommandRecordingLiveOutput(ctx, cmd)

		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("`git submodule sync` failed: %s\n%s", err, output.String())
		}

		return nil
	})
}

func updateSubmodules(ctx context.Context, repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Update submodules in work tree %q", workTreeDir)
	return logboek.Context(ctx).Info().LogProcess(logProcessMsg).DoError(func() error {
		cmd := exec.Command(
			"git", "-c", "core.autocrlf=false",
			"submodule", "update", "--checkout", "--force", "--init", "--recursive",
		)

		cmd.Dir = workTreeDir // required for `git submodule` to work

		output := setCommandRecordingLiveOutput(ctx, cmd)

		if debugWorktreeSwitch() {
			fmt.Printf("[DEBUG WORKTREE SWITCH] %s\n", strings.Join(append([]string{cmd.Path}, cmd.Args[1:]...), " "))
		}

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("`git submodule update` failed: %s\n%s", err, output.String())
		}

		return nil
	})
}
