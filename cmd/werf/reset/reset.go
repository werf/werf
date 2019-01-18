package reset

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/docker_authorizer"
	"github.com/flant/werf/pkg/cleanup"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	OnlyCacheVersion bool

	DryRun bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Delete images, containers, and cache files for all projects created by werf on the host",
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
	cmd.PersistentFlags().BoolVarP(&CmdData.OnlyCacheVersion, "only-cache-version", "", false, "Only delete stages cache, images, and containers created by another werf version")

	cmd.PersistentFlags().BoolVarP(&CmdData.DryRun, "dry-run", "", false, "Indicate what the command would do without actually doing that")

	return cmd
}

func runReset() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
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
