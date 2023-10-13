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

var getNamespaceCmdData common.CmdData

func NewGetNamespaceCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &getNamespaceCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "get-namespace",
		DisableFlagsInUseLine: true,
		Short:                 "Print Kubernetes Namespace that will be used in current configuration with specified params.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&getNamespaceCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runGetNamespace(ctx)
		},
	})

	common.SetupDir(&getNamespaceCmdData, cmd)
	common.SetupGitWorkTree(&getNamespaceCmdData, cmd)
	common.SetupConfigTemplatesDir(&getNamespaceCmdData, cmd)
	common.SetupConfigPath(&getNamespaceCmdData, cmd)
	common.SetupGiterminismConfigPath(&getNamespaceCmdData, cmd)
	common.SetupEnvironment(&getNamespaceCmdData, cmd)

	common.SetupGiterminismOptions(&getNamespaceCmdData, cmd)

	common.SetupTmpDir(&getNamespaceCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&getNamespaceCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupDockerConfig(&getNamespaceCmdData, cmd, "")

	common.SetupLogOptions(&getNamespaceCmdData, cmd)

	return cmd
}

func runGetNamespace(ctx context.Context) error {
	if err := werf.Init(*getNamespaceCmdData.TmpDir, *getNamespaceCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *getNamespaceCmdData.LogVerbose || *getNamespaceCmdData.LogDebug}); err != nil {
		return err
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &getNamespaceCmdData)
	if err != nil {
		return err
	}

	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &getNamespaceCmdData, giterminismManager, common.GetWerfConfigOptions(&getNamespaceCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	namespace, err := deploy_params.GetKubernetesNamespace("", *getNamespaceCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	fmt.Println(namespace)

	return nil
}
