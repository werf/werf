package main

import (
	"github.com/spf13/cobra"

	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/cleanup"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
)

var resetCmdData struct {
	OnlyCacheVersion bool

	DryRun bool
}

func newResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Delete images, containers, and cache files for all projects created by dapp on the host",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReset()
		},
	}

	//cmd.PersistentFlags().BoolVarP(&resetCmdData.OnlyDevModeCache, "only-dev-mode-cache", "", false, "delete stages cache, images, and containers created in developer mode")
	cmd.PersistentFlags().BoolVarP(&resetCmdData.OnlyCacheVersion, "only-cache-version", "", false, "only delete stages cache, images, and containers created by another dapp version")

	cmd.PersistentFlags().BoolVarP(&resetCmdData.DryRun, "dry-run", "", false, "indicate what the command would do without actually doing that")

	return cmd
}

func runReset() error {
	if err := lock.Init(); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	commonOptions := cleanup.CommonOptions{DryRun: resetCmdData.DryRun}
	if resetCmdData.OnlyCacheVersion {
		return cleanup.ResetCacheVersion(commonOptions)
	} else {
		return cleanup.ResetAll(commonOptions)
	}

	return nil
}
