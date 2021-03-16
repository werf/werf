package common

import (
	"context"
	"fmt"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager"
)

func FollowGitHead(ctx context.Context, cmdData *CmdData, taskFunc func(ctx context.Context, headCommitGiterminismManager giterminism_manager.Interface) error) error {
	workingDir := GetWorkingDir(cmdData)
	gitWorkTree, err := GetGitWorkTree(cmdData, workingDir)
	if err != nil {
		return err
	}

	var savedHeadCommit string
	var savedIndexChecksum string
	iterFunc := func() error {
		var currentHeadCommit string
		var currentIndexChecksum string

		repo, err := git_repo.OpenLocalRepo(BackgroundContext(), "own", gitWorkTree, git_repo.OpenLocalRepoOptions{})
		if err != nil {
			return err
		}

		currentHeadCommit, _ = repo.HeadCommit(ctx)
		if *cmdData.Dev {
			currentIndexChecksum, err = repo.StatusIndexChecksum(ctx)
			if err != nil {
				return err
			}
		}

		if savedHeadCommit != currentHeadCommit || savedIndexChecksum != currentIndexChecksum {
			savedHeadCommit = currentHeadCommit
			savedIndexChecksum = currentIndexChecksum

			var header string
			var waitMessage string
			if *cmdData.Dev {
				header = fmt.Sprintf("Commit %q IndexStatusChecksum %q", savedHeadCommit, savedIndexChecksum)
				waitMessage = "Waiting for the new commit or staged changes ..."
			} else {
				header = fmt.Sprintf("Commit %q", savedHeadCommit)
				waitMessage = "Waiting for the new commit ..."
			}

			if err := logboek.Context(ctx).LogProcess(header).
				Options(func(options types.LogProcessOptionsInterface) {
					options.Style(style.Highlight())
				}).
				DoError(func() error {
					giterminismManager, err := GetGiterminismManager(cmdData)
					if err != nil {
						return err
					}

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
		}
	}
}
