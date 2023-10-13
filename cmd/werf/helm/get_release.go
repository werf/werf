package helm

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/config/deploy_params"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var getReleaseCmdData common.CmdData

func NewGetReleaseCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &getReleaseCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "get-release",
		DisableFlagsInUseLine: true,
		Short:                 "Print Helm Release name that will be used in current configuration with specified params.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&getReleaseCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runGetRelease(ctx)
		},
	})

	common.SetupDir(&getReleaseCmdData, cmd)
	common.SetupGitWorkTree(&getReleaseCmdData, cmd)
	common.SetupConfigTemplatesDir(&getReleaseCmdData, cmd)
	common.SetupConfigPath(&getReleaseCmdData, cmd)
	common.SetupGiterminismConfigPath(&getReleaseCmdData, cmd)
	common.SetupNamespace(&getReleaseCmdData, cmd)
	common.SetupEnvironment(&getReleaseCmdData, cmd)

	common.SetupGiterminismOptions(&getReleaseCmdData, cmd)

	common.SetupTmpDir(&getReleaseCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&getReleaseCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupDockerConfig(&getReleaseCmdData, cmd, "")

	common.SetupLogOptions(&getReleaseCmdData, cmd)

	return cmd
}

func runGetRelease(ctx context.Context) error {
	if err := werf.Init(*getReleaseCmdData.TmpDir, *getReleaseCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *getReleaseCmdData.LogVerbose || *getReleaseCmdData.LogDebug}); err != nil {
		return err
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &getReleaseCmdData)
	if err != nil {
		return err
	}

	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &getReleaseCmdData, giterminismManager, common.GetWerfConfigOptions(&getReleaseCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	namespace, err := deploy_params.GetKubernetesNamespace(*getReleaseCmdData.Namespace, *getReleaseCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	release, err := deploy_params.GetHelmRelease("", *getReleaseCmdData.Environment, namespace, werfConfig)
	if err != nil {
		return err
	}

	fmt.Println(release)

	return nil
}
