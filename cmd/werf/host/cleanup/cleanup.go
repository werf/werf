package gc

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleanup"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup old unused werf cache and data of all projects on host machine",
		Long: common.GetLongCommandDescription(`Cleanup old unused werf cache and data of all projects on host machine.

The data include:
* Lost docker containers and images from interrupted builds.
* Old service tmp dirs, which werf creates during every build, publish, deploy and other commands.
* Local cache:
  * Remote git clones cache.
  * Git worktree cache.

It is safe to run this command periodically by automated cleanup job.`),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runGC()
			})
		},
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd)

	return cmd
}

func runGC() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logger.GetOutStream(), Err: logger.GetErrStream()}); err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	return lock.WithLock("gc", lock.LockOptions{}, func() error {
		if err := tmp_manager.GC(); err != nil {
			return fmt.Errorf("project tmp dir gc failed: %s", err)
		}

		commonOptions := cleanup.CommonOptions{DryRun: false} // TODO: DryRun
		if err := cleanup.HostCleanup(commonOptions); err != nil {
			return fmt.Errorf("project tmp dir gc failed: %s", err)
		}

		return nil
	})
}
