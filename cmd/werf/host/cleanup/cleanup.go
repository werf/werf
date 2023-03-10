package cleanup

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/host_cleaning"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

var cmdData struct {
	Force bool
}

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "cleanup",
		Short:                 "Cleanup old unused werf cache and data of all projects on host machine.",
		Long:                  common.GetLongCommandDescription(GetCleanupDocs().Long),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.DocsLongMD: GetCleanupDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
				return fmt.Errorf("initialization error: %w", err)
			}

			return common.WithContext(true, func(ctx context.Context) error {
				defer global_warnings.PrintGlobalWarnings(ctx)

				if err := common.ProcessLogOptions(&commonCmdData); err != nil {
					common.PrintHelp(cmd)
					return err
				}
				common.LogVersion()

				return common.LogRunningTime(func() error { return runCleanup(ctx) })
			})
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
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
	commonCmdData.SetupPlatform(cmd)
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.Force, "force", "", util.GetBoolEnvironmentDefaultFalse("WERF_FORCE"), "Force deletion of images which are being used by some containers (default $WERF_FORCE)")

	return cmd
}

func runCleanup(ctx context.Context) error {
	containerBackend, processCtx, err := common.InitProcessContainerBackend(ctx, &commonCmdData)
	if err != nil {
		return err
	}
	ctx = processCtx

	projectName := *commonCmdData.ProjectName
	if projectName != "" {
		return fmt.Errorf("no functionality for cleaning a certain project is implemented (--project-name=%s)", projectName)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
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

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogDebug}); err != nil {
		return err
	}

	logboek.LogOptionalLn()

	cleanupDockerServer := false
	if _, match := containerBackend.(*container_backend.DockerServerBackend); match {
		cleanupDockerServer = true
	}

	hostCleanupOptions := host_cleaning.HostCleanupOptions{
		DryRun:              *commonCmdData.DryRun,
		Force:               cmdData.Force,
		CleanupDockerServer: cleanupDockerServer,
		AllowedDockerStorageVolumeUsagePercentage:       commonCmdData.AllowedDockerStorageVolumeUsage,
		AllowedDockerStorageVolumeUsageMarginPercentage: commonCmdData.AllowedDockerStorageVolumeUsageMargin,
		AllowedLocalCacheVolumeUsagePercentage:          commonCmdData.AllowedLocalCacheVolumeUsage,
		AllowedLocalCacheVolumeUsageMarginPercentage:    commonCmdData.AllowedLocalCacheVolumeUsageMargin,
		DockerServerStoragePath:                         commonCmdData.DockerServerStoragePath,
	}

	return host_cleaning.RunHostCleanup(ctx, containerBackend, hostCleanupOptions)
}
