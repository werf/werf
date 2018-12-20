package gc

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/dapp/cmd/dapp/common"
	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/cleanup"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/project_tmp_dir"
	"github.com/flant/dapp/pkg/true_git"
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
	if err := dapp.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
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

		if err := cleanup.RemoveLostTmpDappFiles(); err != nil {
			return fmt.Errorf("unable to remove lost tmp dapp files: %s", err)
		}

		return nil
	})
}
