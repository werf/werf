package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
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

	return cmd
}

func run(ctx context.Context) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogDebug}); err != nil {
		return err
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return err
	}

	for _, imageName := range werfConfig.GetImageNameList(cmdData.finalImagesOnly) {
		if imageName == "" {
			fmt.Println("~")
		} else {
			fmt.Println(imageName)
		}
	}

	return nil
}
