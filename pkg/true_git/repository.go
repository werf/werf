package true_git

import (
	"os/exec"

	"github.com/go-git/go-git/v5"
)

func GitOpenWithCustomWorktreeDir(gitDir, worktreeDir string) (*git.Repository, error) {
	return git.PlainOpenWithOptions(worktreeDir, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
}

type FetchOptions struct {
	All       bool
	TagsOnly  bool
	Prune     bool
	PruneTags bool
}

func Fetch(gitDir string, options FetchOptions) error {
	command := "git"
	commandArgs := []string{"-C", gitDir, "fetch"}

	if options.All {
		commandArgs = append(commandArgs, "--all")
	}

	if options.TagsOnly {
		commandArgs = append(commandArgs, "--tags")
	}

	if options.Prune || options.PruneTags {
		commandArgs = append(commandArgs, "--prune")

		if options.PruneTags {
			commandArgs = append(commandArgs, "--prune-tags")
		}
	}

	cmd := exec.Command(command, commandArgs...)
	cmd.Stdout = outStream
	cmd.Stderr = errStream

	return cmd.Run()
}
