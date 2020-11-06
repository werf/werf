package common

import (
	"context"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
)

func FollowGitHead(ctx context.Context, cmdData *CmdData, taskFunc func(ctx context.Context) error) error {
	projectDir, err := GetProjectDir(cmdData)
	if err != nil {
		return fmt.Errorf("unable to get project dir: %s", err)
	}

	repo, err := git.PlainOpen(projectDir)
	if err != nil {
		return fmt.Errorf("unable to get project git repository: %s", err)
	}

	var refHash string
	iterFunc := func() error {
		ref, err := repo.Head()
		if err != nil {
			return err
		}

		if refHash != ref.Hash().String() {
			refHash = ref.Hash().String()

			if err := logboek.Context(ctx).LogProcess("Commit %s", refHash).
				Options(func(options types.LogProcessOptionsInterface) {
					options.Style(style.Highlight())
				}).
				DoError(func() error {
					return taskFunc(ctx)
				}); err != nil {
				return err
			}

			logboek.Context(ctx).LogLn("Waiting for new commit ...")
			logboek.Context(ctx).LogOptionalLn()
		} else {
			time.Sleep(1 * time.Second)
		}

		return nil
	}

	if err := iterFunc(); err != nil {
		return err
	}

	for {
		if err := iterFunc(); err != nil {
			logboek.Context(ctx).Warn().LogLn(err)
		}
	}
}
