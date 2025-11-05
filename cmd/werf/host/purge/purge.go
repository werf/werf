package reset

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/cleaning"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/host_cleaning"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	Force bool
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "purge",
		Short:                 "Purge werf images, cache and other data for all projects on host machine",
		Long:                  common.GetLongCommandDescription(GetPurgeDocs().Long),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.DocsLongMD: GetPurgeDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error { return runReset(ctx) })
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupDockerConfig(&commonCmdData, cmd, "")
	common.SetupProjectName(&commonCmdData, cmd, true)
	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupGiterminismOptions(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)
	cmd.Flags().BoolVarP(&cmdData.Force, "force", "", false, common.CleaningCommandsForceOptionDescription)

	return cmd
}

func runReset(ctx context.Context) error {
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                         &commonCmdData,
		InitWerf:                    true,
		InitGitDataManager:          true,
		InitManifestCache:           true,
		InitProcessContainerBackend: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	containerBackend := commonManager.ContainerBackend()

	projectName := *commonCmdData.ProjectName
	if projectName == "" {
		logboek.LogOptionalLn()
		hostPurgeOptions := host_cleaning.HostPurgeOptions{DryRun: *commonCmdData.DryRun, RmContainersThatUseWerfImages: cmdData.Force}
		if err := host_cleaning.HostPurge(ctx, containerBackend, hostPurgeOptions); err != nil {
			return err
		}
	} else {
		if _, ok := containerBackend.(*container_backend.DockerServerBackend); !ok {
			logboek.Context(ctx).Warn().LogF("Skip cleaning local storage with buildah backend (not implemented)\n")
			return nil
		}
		storageManager, err := common.NewStorageManagerWithOptions(ctx, &common.NewStorageManagerConfig{
			ProjectName:      projectName,
			ContainerBackend: containerBackend,
			CmdData:          &commonCmdData,
		}, common.WithHostPurge())
		if err != nil {
			return fmt.Errorf("unable to init storage manager: %w", err)
		}

		purgeOptions := cleaning.PurgeOptions{
			RmContainersThatUseWerfImages: cmdData.Force,
			DryRun:                        *commonCmdData.DryRun,
		}

		logboek.LogOptionalLn()
		if err := cleaning.Purge(ctx, projectName, storageManager, purgeOptions); err != nil {
			return err
		}
	}

	return nil
}
