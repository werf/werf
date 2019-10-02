package cleanup

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/lock"
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

It is safe to run this command periodically by automated cleanup job in parallel with other werf commands such as build, deploy, stages and images cleanup.`),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&CommonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runGC()
			})
		},
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "")
	common.SetupInsecureRegistry(&CommonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&CommonCmdData, cmd)

	common.SetupLogOptions(&CommonCmdData, cmd)

	common.SetupDryRun(&CommonCmdData, cmd)

	return cmd
}

func runGC() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream()}); err != nil {
		return err
	}

	if err := docker_registry.Init(docker_registry.Options{InsecureRegistry: *CommonCmdData.InsecureRegistry, SkipTlsVerifyRegistry: *CommonCmdData.SkipTlsVerifyRegistry}); err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	logboek.LogOptionalLn()
	hostCleanupOptions := cleaning.HostCleanupOptions{DryRun: *CommonCmdData.DryRun}
	if err := cleaning.HostCleanup(hostCleanupOptions); err != nil {
		return err
	}

	return nil
}
