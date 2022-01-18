package list

import (
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
		imagesOnly bool
	}
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "list",
		DisableFlagsInUseLine: true,
		Short:                 "List image and artifact names defined in werf.yaml",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return run()
		},
	}

	cmd.Flags().BoolVarP(&cmdData.imagesOnly, "images-only", "", false, "Show image names without artifacts")

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}

func run() error {
	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %s", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	_, werfConfig, err := common.GetRequiredWerfConfig(common.BackgroundContext(), &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return err
	}

	for _, image := range werfConfig.StapelImages {
		if image.Name == "" {
			fmt.Println("~")
		} else {
			fmt.Println(image.Name)
		}
	}

	for _, image := range werfConfig.ImagesFromDockerfile {
		if image.Name == "" {
			fmt.Println("~")
		} else {
			fmt.Println(image.Name)
		}
	}

	if !cmdData.imagesOnly {
		for _, image := range werfConfig.Artifacts {
			fmt.Println(image.Name)
		}
	}

	return nil
}
