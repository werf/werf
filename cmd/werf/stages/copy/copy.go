package copy

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build/stages"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	From string
	To   string
	All  bool
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "copy",
		Short:                 "Copy stages between container registry and archive",
		Long:                  common.GetLongCommandDescription(GetCopyDocs().Long),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
			common.DocsLongMD: GetCopyDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runCopy(ctx) })
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})
	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repos")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupFinalRepo(&commonCmdData, cmd)
	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	commonCmdData.SetupFinalImagesOnly(cmd, false)
	commonCmdData.SetupPlatform(cmd)

	cmd.Flags().StringVarP(&cmdData.From, "from", "", "", "Source address to copy stages from. Use archive:PATH for stage archive or [docker://]REPO for container registry.")
	cmd.Flags().StringVarP(&cmdData.To, "to", "", "", "Destination address to copy stages to. Use archive:PATH for stage archive or [docker://]REPO for container registry.")
	cmd.Flags().BoolVarP(&cmdData.All, "all", "", true, "Copy all project stages (default: true). If false, copy only stages for current build.")

	return cmd
}

func runCopy(ctx context.Context) error {
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                         &commonCmdData,
		InitWerf:                    true,
		InitGitDataManager:          true,
		InitDockerRegistry:          true,
		InitProcessContainerBackend: true,
		InitManifestCache:           true,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	if cmdData.From == "" {
		return fmt.Errorf("--from=ADDRESS param required")
	}

	if cmdData.To == "" {
		return fmt.Errorf("--to=ADDRESS param required")
	}

	if cmdData.From == cmdData.To {
		return fmt.Errorf("--from=ADDRESS and --to=ADDRESS must be different")
	}

	fromAddrRaw := cmdData.From
	toAddrRaw := cmdData.To

	fromAddr, err := ref.ParseAddr(fromAddrRaw)
	if err != nil {
		return fmt.Errorf("invalid from add %q: %w", fromAddrRaw, err)
	}

	toAddr, err := ref.ParseAddr(toAddrRaw)
	if err != nil {
		return fmt.Errorf("invalid to addr %q: %w", toAddrRaw, err)
	}

	containerBackend := commonManager.ContainerBackend()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}
	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, config.WerfConfigOptions{LogRenderedFilePath: true, Env: commonCmdData.Environment})
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}

	return logboek.Context(ctx).LogProcess("Copy stages").DoError(func() error {
		logboek.Context(ctx).Info().LogFDetails("From: %s\n", fromAddr)
		logboek.Context(ctx).Info().LogFDetails("To: %s\n", toAddr.String())

		return stages.Copy(ctx, fromAddr, toAddr, stages.CopyOptions{
			AllStages:          cmdData.All,
			ProjectName:        projectName,
			BaseTmpDir:         projectTmpDir,
			ContainerBackend:   containerBackend,
			CommonCmdData:      &commonCmdData,
			WerfConfig:         werfConfig,
			GiterminismManager: giterminismManager,
		})
	})
}
