package cleanup

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/host_cleaning"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
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

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error { return runCleanup(ctx) })
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupDockerConfig(&commonCmdData, cmd, "")
	common.SetupProjectName(&commonCmdData, cmd, true)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupBackendStoragePath(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.Force, "force", "", util.GetBoolEnvironmentDefaultFalse("WERF_FORCE"), "Force deletion of images which are being used by some containers (default $WERF_FORCE)")

	return cmd
}

func runCleanup(ctx context.Context) error {
	projectName := *commonCmdData.ProjectName
	if projectName != "" {
		return fmt.Errorf("no functionality for cleaning a certain project is implemented (--project-name or WERF_PROJECT_NAME = %s)", projectName)
	}

	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
		InitProcessContainerBackend: true,
		InitGitDataManager:          true,
		InitManifestCache:           true,
		InitLRUImagesCache:          true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	logboek.LogOptionalLn()

	hostCleanupOptions := host_cleaning.HostCleanupOptions{
		DryRun: *commonCmdData.DryRun,
		Force:  cmdData.Force,
		AllowedBackendStorageVolumeUsagePercentage:       commonCmdData.AllowedBackendStorageVolumeUsage,
		AllowedBackendStorageVolumeUsageMarginPercentage: commonCmdData.AllowedBackendStorageVolumeUsageMargin,
		AllowedLocalCacheVolumeUsagePercentage:           commonCmdData.AllowedLocalCacheVolumeUsage,
		AllowedLocalCacheVolumeUsageMarginPercentage:     commonCmdData.AllowedLocalCacheVolumeUsageMargin,
		BackendStoragePath:                               commonCmdData.BackendStoragePath,
	}

	return host_cleaning.RunHostCleanup(ctx, commonManager.ContainerBackend(), hostCleanupOptions)
}
