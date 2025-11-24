package secret

import (
	"cmp"
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/cmd/werf/docs/replacers/helm"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "rotate-secret-key [EXTRA_SECRET_VALUES_FILE_PATH...]",
		DisableFlagsInUseLine: true,
		Short:                 "Regenerate secret files with new secret key",
		Long:                  common.GetLongCommandDescription(helm.GetHelmSecretRotateSecretKeyDocs().Long),
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey, common.WerfOldSecretKey),
			common.DocsLongMD: helm.GetHelmSecretRotateSecretKeyDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runRotateSecretKey(ctx, cmd, args...)
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigRenderPath(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	return cmd
}

func runRotateSecretKey(
	ctx context.Context,
	cmd *cobra.Command,
	secretValuesPaths ...string,
) error {
	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug},
		},
		InitWerf:           true,
		InitGitDataManager: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	relChartPath, err := common.GetHelmChartDir(werfConfigPath, werfConfig, giterminismManager)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %w", err)
	}

	chartPath := filepath.Join(giterminismManager.ProjectDir(), relChartPath)

	ctx = log.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultSecretKeyRotateLogLevel), log.SetupLoggingOptions{
		ColorMode: *commonCmdData.LogColorMode,
	})

	if err := action.SecretKeyRotate(ctx, action.SecretKeyRotateOptions{
		ChartDirPath:      chartPath,
		SecretValuesFiles: secretValuesPaths,
		SecretWorkDir:     giterminismManager.ProjectDir(),
	}); err != nil {
		return fmt.Errorf("rotate secret key: %w", err)
	}

	return nil
}
