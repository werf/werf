package reset

import (
	"context"
	"errors"
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
	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupGiterminismOptions(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupSaveCleanupReport(&commonCmdData, cmd)
	common.SetupCleanupReportPath(&commonCmdData, cmd)
	common.SetupSaveHostCleanupReport(&commonCmdData, cmd)
	common.SetupHostCleanupReportPath(&commonCmdData, cmd)

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
		if common.GetSaveCleanupReport(&commonCmdData) {
			logboek.Context(ctx).Warn().LogF("--save-cleanup-report has no effect without --project-name, no registry cleanup report will be written\n")
		}

		report, reportPath, err := common.NewHostCleanupReport(ctx, &commonCmdData, "host purge", *commonCmdData.DryRun)
		if err != nil {
			return err
		}

		logboek.LogOptionalLn()
		hostPurgeOptions := host_cleaning.HostPurgeOptions{DryRun: *commonCmdData.DryRun, RmContainersThatUseWerfImages: cmdData.Force, Report: report}
		runErr := host_cleaning.HostPurge(ctx, containerBackend, hostPurgeOptions)

		return errors.Join(runErr, report.Save(ctx, reportPath))
	} else {
		if common.GetSaveHostCleanupReport(&commonCmdData) {
			logboek.Context(ctx).Warn().LogF("--save-host-cleanup-report has no effect with --project-name, no host cleanup report will be written\n")
		}

		if _, ok := containerBackend.(*container_backend.DockerServerBackend); !ok {
			if common.GetSaveCleanupReport(&commonCmdData) {
				logboek.Context(ctx).Warn().LogF("No cleanup report will be written: cleaning local storage is not implemented for the buildah backend\n")
			}
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

		report, reportPath, err := common.NewCleanupReport(ctx, &commonCmdData, "host purge", *commonCmdData.DryRun, storageManager)
		if err != nil {
			return err
		}

		purgeOptions := cleaning.PurgeOptions{
			RmContainersThatUseWerfImages: cmdData.Force,
			DryRun:                        *commonCmdData.DryRun,
			Report:                        report,
		}

		logboek.LogOptionalLn()
		runErr := cleaning.Purge(ctx, projectName, storageManager, purgeOptions)

		return errors.Join(runErr, report.Save(ctx, reportPath))
	}
}
