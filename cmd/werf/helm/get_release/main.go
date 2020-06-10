package get_release

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "get-release",
		DisableFlagsInUseLine: true,
		Short:                 "Print Helm Release name that will be used in current configuration with specified params",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runGetRelease()
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "")

	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}

func runGetRelease() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := docker.Init(*commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, &commonCmdData, false)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	release, err := common.GetHelmRelease("", *commonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	fmt.Println(release)

	return nil
}
