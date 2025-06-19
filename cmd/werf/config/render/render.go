package render

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "render [IMAGE_NAME...]",
		DisableFlagsInUseLine: true,
		Short:                 GetRenderDocs().Short,
		Annotations: map[string]string{
			common.DocsLongMD: GetRenderDocs().ShortMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			global_warnings.SuppressGlobalWarnings = true
			if *commonCmdData.LogVerbose || *commonCmdData.LogDebug {
				global_warnings.SuppressGlobalWarnings = false
			}
			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
				Cmd: &commonCmdData,
				InitTrueGitWithOptions: &common.InitTrueGitOptions{
					Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
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

			configOpts := common.GetWerfConfigOptions(&commonCmdData, false)

			customWerfConfigRelPath, err := common.GetCustomWerfConfigRelPath(giterminismManager, &commonCmdData)
			if err != nil {
				return err
			}

			customWerfConfigTemplatesDirRelPath, err := common.GetCustomWerfConfigTemplatesDirRelPath(giterminismManager, &commonCmdData)
			if err != nil {
				return err
			}

			return config.RenderWerfConfig(ctx, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, args, giterminismManager, configOpts)
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupLogOptionsDefaultQuiet(&commonCmdData, cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	return cmd
}
