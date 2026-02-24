package build

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/true_git"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "build [PATH]",
		Short:                 "Build TypeScript chart",
		Long:                  common.GetLongCommandDescription("Build bundle for TypeScript chart."),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			var chartDir string
			if len(args) > 0 {
				chartDir = args[0]
			} else {
				chartDir = "."
			}

			common.LogVersion()

			return runChartTSBuild(ctx, chartDir)
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigRenderPath(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupGiterminismOptions(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}

func runChartTSBuild(ctx context.Context, chartDir string) error {
	if *commonCmdData.Dir == "" && chartDir != "." {
		commonCmdData.Dir = &chartDir
	}

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

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	file.ChartFileReader = giterminismManager.FileManager

	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	relChartPath, err := common.GetHelmChartDir(werfConfigPath, werfConfig, giterminismManager)
	if err != nil {
		return fmt.Errorf("get helm chart dir: %w", err)
	}

	chartPath := filepath.Join(giterminismManager.ProjectDir(), relChartPath)

	ctx = log.SetupLogging(ctx, common.GetNelmLogLevel(&commonCmdData), log.SetupLoggingOptions{
		ColorMode: *commonCmdData.LogColorMode,
	})

	if err := action.ChartTSBuild(ctx, action.ChartTSBuildOptions{
		ChartDirPath: chartPath,
	}); err != nil {
		return fmt.Errorf("chart ts build: %w", err)
	}

	return nil
}
