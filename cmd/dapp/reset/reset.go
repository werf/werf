package reset

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/dapp/cmd/dapp/common"
	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/cleanup"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
)

var CmdData struct {
	OnlyCacheVersion bool

	DryRun bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Delete images, containers, and cache files for all projects created by dapp on the host",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runReset()
			if err != nil {
				return fmt.Errorf("reset failed: %s", err)
			}
			return nil
		},
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	//cmd.PersistentFlags().BoolVarP(&CmdData.OnlyDevModeCache, "only-dev-mode-cache", "", false, "delete stages cache, images, and containers created in developer mode")
	cmd.PersistentFlags().BoolVarP(&CmdData.OnlyCacheVersion, "only-cache-version", "", false, "Only delete stages cache, images, and containers created by another dapp version")

	cmd.PersistentFlags().BoolVarP(&CmdData.DryRun, "dry-run", "", false, "Indicate what the command would do without actually doing that")

	return cmd
}

func runReset() error {
	if err := dapp.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	commonOptions := cleanup.CommonOptions{DryRun: CmdData.DryRun}
	if CmdData.OnlyCacheVersion {
		return cleanup.ResetCacheVersion(commonOptions)
	} else {
		return cleanup.ResetAll(commonOptions)
	}

	return nil
}
