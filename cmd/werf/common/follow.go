package common

import (
	"context"
	"fmt"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/giterminism_manager"
)

func FollowGitHead(ctx context.Context, cmdData *CmdData, taskFunc func(ctx context.Context, iterGiterminismManager giterminism_manager.Interface) error) error {
	var waitMessage string
	if *cmdData.Dev {
		waitMessage = "Waiting for new changes ..."
	} else {
		waitMessage = "Waiting for the new commit ..."
	}

	var savedHeadCommit string
	iterFunc := func() error {
		giterminismManager, err := GetGiterminismManager(ctx, cmdData)
		if err != nil {
			return fmt.Errorf("unable to get giterminism manager: %w", err)
		}

		currentHeadCommit := giterminismManager.HeadCommit()
		if savedHeadCommit != currentHeadCommit {
			savedHeadCommit = currentHeadCommit

			if err := logboek.Context(ctx).LogProcess("Commit %q", savedHeadCommit).
				Options(func(options types.LogProcessOptionsInterface) {
					options.Style(style.Highlight())
				}).
				DoError(func() error {
					return taskFunc(ctx, giterminismManager)
				}); err != nil {
				return err
			}

			logboek.Context(ctx).LogLn(waitMessage)
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
			logboek.Context(ctx).LogLn(waitMessage)
			logboek.Context(ctx).LogOptionalLn()
		}
	}
}
