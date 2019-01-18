package gc

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/docker_authorizer"
	"github.com/flant/werf/pkg/cleanup"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/project_tmp_dir"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "gc",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runGC()
			if err != nil {
				return fmt.Errorf("gc failed: %s", err)
			}
			return nil
		},
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	return cmd
}

func runGC() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	return lock.WithLock("gc", lock.LockOptions{}, func() error {
		err := project_tmp_dir.GC()
		if err != nil {
			return fmt.Errorf("project tmp dir gc failed: %s", err)
		}

		if err := cleanup.RemoveLostTmpWerfFiles(); err != nil {
			return fmt.Errorf("unable to remove lost tmp werf files: %s", err)
		}

		return nil
	})
}
