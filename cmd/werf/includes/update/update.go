package update

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/true_git"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, common.SetCommandContext(ctx, &cobra.Command{
		Use:   "update",
		Short: "Update werf-includes.lock file",
		Long:  "Update werf-includes.lock file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
				Cmd:                &commonCmdData,
				InitWerf:           true,
				InitGitDataManager: true,
				InitTrueGitWithOptions: &common.InitTrueGitOptions{
					Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
				},
			})
			if err != nil {
				return fmt.Errorf("component init error: %w", err)
			}

			_, err = common.GetGiterminismManager(ctx, &commonCmdData)
			if err != nil {
				return err
			}

			return nil
		},
	}))

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupIntrospectAfterError(&commonCmdData, cmd)
	common.SetupIntrospectBeforeError(&commonCmdData, cmd)
	common.SetupIntrospectStage(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, true)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupSkipTLSVerifyKube(&commonCmdData, cmd)
	common.SetupKubeApiServer(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyHelmDependencies(&commonCmdData, cmd)
	common.SetupKubeCaPath(&commonCmdData, cmd)
	common.SetupKubeTlsServer(&commonCmdData, cmd)
	common.SetupKubeToken(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	common.SetupRelease(&commonCmdData, cmd, true)
	common.SetupNamespace(&commonCmdData, cmd, true)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddAnnotationSeparator(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)
	common.SetupAddLabelSeparator(&commonCmdData, cmd)

	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd, true)
	common.SetupSecretValues(&commonCmdData, cmd, true)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	commonCmdData.SetupDisableDefaultValues(cmd)
	commonCmdData.SetupDisableDefaultSecretValues(cmd)
	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)
	commonCmdData.SetupWithoutImages(cmd)

	common.SetupSaveBuildReport(&commonCmdData, cmd)
	common.SetupBuildReportPath(&commonCmdData, cmd)

	common.SetupSaveDeployReport(&commonCmdData, cmd)
	common.SetupDeployReportPath(&commonCmdData, cmd)

	common.SetupUseCustomTag(&commonCmdData, cmd)
	common.SetupAddCustomTag(&commonCmdData, cmd)
	common.SetupVirtualMerge(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)
	common.SetupRequireBuiltImages(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)
	common.SetupFollow(&commonCmdData, cmd)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupBackendStoragePath(&commonCmdData, cmd)

	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupKubeQpsLimit(&commonCmdData, cmd)
	common.SetupKubeBurstLimit(&commonCmdData, cmd)
	common.SetupDeployGraphPath(&commonCmdData, cmd)
	common.SetupRollbackGraphPath(&commonCmdData, cmd)

	common.SetupRenderSubchartNotes(&commonCmdData, cmd)

	commonCmdData.SetupCreateIncludesLockFile(true)

	commonCmdData.SetupUseIncludesLatestVersions(cmd)

	return cmd
}
