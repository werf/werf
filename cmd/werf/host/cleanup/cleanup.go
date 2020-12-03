package cleanup

import (
	"fmt"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_inspector"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/werf/global_warnings"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/host_cleaning"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup old unused werf cache and data of all projects on host machine",
		Long: common.GetLongCommandDescription(`Cleanup old unused werf cache and data of all projects on host machine.

The data include:
* Lost docker containers and images from interrupted builds.
* Old service tmp dirs, which werf creates during every build, converge and other commands.
* Local cache:
  * Remote git clones cache.
  * Git worktree cache.

It is safe to run this command periodically by automated cleanup job in parallel with other werf commands such as build, converge and cleanup.`),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer global_warnings.PrintGlobalWarnings(common.BackgroundContext())

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runGC()
			})
		},
	}

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "")

	common.SetupDisableGiterminism(&commonCmdData, cmd)
	common.SetupNonStrictGiterminismInspection(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	return cmd
}

func runGC() error {
	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := giterminism_inspector.Init(giterminism_inspector.InspectionOptions{DisableGiterminism: *commonCmdData.DisableGiterminism, NonStrict: *commonCmdData.NonStrictGiterminismInspection}); err != nil {
		return err
	}

	if err := git_repo.Init(); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := docker.Init(ctx, *commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return err
	}
	ctx = ctxWithDockerCli

	logboek.LogOptionalLn()
	hostCleanupOptions := host_cleaning.HostCleanupOptions{DryRun: *commonCmdData.DryRun}
	if err := host_cleaning.HostCleanup(ctx, hostCleanupOptions); err != nil {
		return err
	}

	return nil
}
