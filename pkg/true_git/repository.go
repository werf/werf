package true_git

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/util"
)

func GitOpenWithCustomWorktreeDir(gitDir, worktreeDir string) (*git.Repository, error) {
	return git.PlainOpenWithOptions(worktreeDir, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
}

type FetchOptions struct {
	All       bool
	TagsOnly  bool
	Prune     bool
	PruneTags bool
	Unshallow bool
	RefSpecs  map[string]string
}

func Fetch(ctx context.Context, path string, options FetchOptions) error {
	command := "git"
	commandArgs := append(getCommonGitOptions(), "-C", path, "fetch")

	if options.Unshallow {
		commandArgs = append(commandArgs, "--unshallow")
	}

	if options.All {
		commandArgs = append(commandArgs, "--all")
	}

	if options.TagsOnly {
		commandArgs = append(commandArgs, "--tags")
	}

	if options.Prune || options.PruneTags {
		commandArgs = append(commandArgs, "--prune")

		if options.PruneTags && !gitVersion.LessThan(semver.MustParse("2.17.0")) {
			commandArgs = append(commandArgs, "--prune-tags")
		}
	}

	for remote, refSpec := range options.RefSpecs {
		commandArgs = append(commandArgs, remote, refSpec)
	}

	logboek.Context(ctx).Debug().LogLnDetails(command, strings.Join(commandArgs, " "))

	cmd := exec.Command(command, commandArgs...)
	cmd.Stdout = logboek.Context(ctx).OutStream()
	cmd.Stderr = logboek.Context(ctx).ErrStream()

	return cmd.Run()
}

func IsShallowClone(path string) (bool, error) {
	if gitVersion.LessThan(semver.MustParse("2.15.0")) {
		exist, err := util.FileExists(filepath.Join(path, ".git", "shallow"))
		if err != nil {
			return false, err
		}

		return exist, nil
	}

	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-shallow-repository")

	res, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(res)) == "true", nil
}
