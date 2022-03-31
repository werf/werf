package graph

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "graph [IMAGE_NAME...]",
		DisableFlagsInUseLine: true,
		Short:                 "Print dependency graph for images in werf.yaml",
		Example: `  # Print dependency graph
  $ werf config graph
  - image: app1
    dependsOn:
      from: baseImage
  - image: app2
    dependsOn:
      from: baseImage
      import:
      - app1
  - image: baseImage

  # Print dependency graph for a certain image
  $ werf config graph app
  - image: app2
    dependsOn:
      from: baseImage
      import:
      - app1
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := common.GetContext()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

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

			if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
				return err
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

			_, werfConfig, err := config.GetWerfConfig(common.GetContext(), customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, giterminismManager, configOpts)
			if err != nil {
				return err
			}

			graphList, err := werfConfig.GetImageGraphList(args)
			if err != nil {
				return err
			}

			data, err := yaml.Marshal(graphList)
			if err != nil {
				return err
			}

			fmt.Println(strings.TrimSpace(string(data)))
			return nil
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}
