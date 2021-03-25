package cleanup

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/host_cleaning"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

var cmdData struct {
	Force bool
}

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

			return common.LogRunningTime(runGC)
		},
	}

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "")
	common.SetupProjectName(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedDockerStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedDockerStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupDockerServerStoragePath(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.Force, "force", "", common.GetBoolEnvironmentDefaultFalse("WERF_FORCE"), "Force deletion of images which are being used by some containers (default $WERF_FORCE)")

	return cmd
}

func runGC() error {
	ctx := common.BackgroundContext()

	projectName := *commonCmdData.ProjectName
	if projectName != "" {
		return fmt.Errorf("no functionality for cleaning a certain project is implemented (--project-name=%s)", projectName)
	}

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %s", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := lrumeta.Init(); err != nil {
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

	hostCleanupOptions := host_cleaning.HostCleanupOptions{
		DryRun: *commonCmdData.DryRun,
		Force:  cmdData.Force,
		AllowedDockerStorageVolumeUsagePercentage:       commonCmdData.AllowedDockerStorageVolumeUsage,
		AllowedDockerStorageVolumeUsageMarginPercentage: commonCmdData.AllowedDockerStorageVolumeUsageMargin,
		AllowedLocalCacheVolumeUsagePercentage:          commonCmdData.AllowedLocalCacheVolumeUsage,
		AllowedLocalCacheVolumeUsageMarginPercentage:    commonCmdData.AllowedLocalCacheVolumeUsageMargin,
		DockerServerStoragePath:                         *commonCmdData.DockerServerStoragePath,
	}

	return host_cleaning.RunHostCleanup(ctx, hostCleanupOptions)
}
