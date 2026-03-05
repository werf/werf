package chart_ts_init

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/nelm/pkg/action"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/slug"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "init [PATH]",
		Short:                 "Initialize a new TypeScript chart project",
		Long:                  common.GetLongCommandDescription("Initialize a new werf project with TypeScript chart scaffolding. Creates werf.yaml and .helm/ directory with TypeScript boilerplate. If PATH is not specified, uses the current directory."),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			var targetDir string
			if len(args) > 0 {
				targetDir = args[0]
			} else {
				targetDir = "."
			}

			common.LogVersion()

			return runChartTSInit(ctx, targetDir)
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}

func runChartTSInit(ctx context.Context, targetDir string) error {
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("get absolute path: %w", err)
	}

	werfConfigPath := filepath.Join(absTargetDir, "werf.yaml")
	if _, err := os.Stat(werfConfigPath); err == nil {
		return fmt.Errorf("werf.yaml already exists in %s, project is already initialized", absTargetDir)
	}

	if err := os.MkdirAll(absTargetDir, 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	projectName := slug.Project(strings.ToLower(filepath.Base(absTargetDir)))

	werfConfigContent := fmt.Sprintf("configVersion: 1\nproject: %s\n", projectName)
	if err := os.WriteFile(werfConfigPath, []byte(werfConfigContent), 0o644); err != nil {
		return fmt.Errorf("create werf.yaml: %w", err)
	}

	helmDir := filepath.Join(absTargetDir, ".helm")

	ctx = log.SetupLogging(ctx, common.GetNelmLogLevel(&commonCmdData), log.SetupLoggingOptions{
		ColorMode: *commonCmdData.LogColorMode,
	})

	if err := action.ChartTSInit(ctx, action.ChartTSInitOptions{
		ChartDirPath: helmDir,
		ChartName:    projectName,
	}); err != nil {
		if removeErr := os.Remove(werfConfigPath); removeErr != nil {
			return fmt.Errorf("chart ts init failed: %w (also failed to cleanup werf.yaml: %v)", err, removeErr)
		}
		return fmt.Errorf("chart ts init: %w", err)
	}

	return nil
}
