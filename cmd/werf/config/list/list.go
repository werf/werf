package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var (
	commonCmdData common.CmdData
	cmdData       struct {
		finalImagesOnly bool
	}
)

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "list",
		DisableFlagsInUseLine: true,
		Short:                 GetListDocs().Short,
		Annotations: map[string]string{
			common.DocsLongMD: GetListDocs().ShortMD,
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			global_warnings.SuppressGlobalWarnings = true

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return run(ctx)
		},
	})

	// Setup final-images-only flag.
	{
		name := "final-images-only"
		deprecatedName := "images-only"
		for _, n := range []string{name, deprecatedName} {
			// FIXME: it should be default behavior.
			cmd.Flags().BoolVarP(&cmdData.finalImagesOnly, n, "", false, "Show only final images")
		}

		if err := cmd.Flags().MarkHidden(deprecatedName); err != nil {
			panic(fmt.Errorf("error marking flag hidden: %w", err))
		}

		if err := cmd.Flags().MarkDeprecated(deprecatedName, fmt.Sprintf("use --%s instead", name)); err != nil {
			panic(fmt.Errorf("error marking flag deprecated: %w", err))
		}
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupLogOptions(&commonCmdData, cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	return cmd
}

func run(ctx context.Context) error {
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

	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return err
	}

	imagesToProcess, err := config.NewImagesToProcess(werfConfig, nil, cmdData.finalImagesOnly, false)
	if err != nil {
		return err
	}

	for _, imageName := range imagesToProcess.ImageNameList {
		if imageName == "" {
			fmt.Println("~")
		} else {
			fmt.Println(imageName)
		}
	}

	return nil
}
