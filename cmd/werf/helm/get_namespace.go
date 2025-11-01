package helm

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/config/deploy_params"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
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

			global_warnings.SuppressGlobalWarnings = true

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
	common.SetupConfigRenderDir(&commonCmdData, cmd)
	common.SetupConfigPath(&getNamespaceCmdData, cmd)
	common.SetupGiterminismConfigPath(&getNamespaceCmdData, cmd)
	common.SetupEnvironment(&getNamespaceCmdData, cmd)

	common.SetupGiterminismOptions(&getNamespaceCmdData, cmd)

	common.SetupTmpDir(&getNamespaceCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&getNamespaceCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupDockerConfig(&getNamespaceCmdData, cmd, "")

	common.SetupLogOptions(&getNamespaceCmdData, cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	return cmd
}

func runGetNamespace(ctx context.Context) error {
	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &getNamespaceCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *getNamespaceCmdData.LogVerbose || *getNamespaceCmdData.LogDebug},
		},
		InitWerf:           true,
		InitGitDataManager: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &getNamespaceCmdData)
	if err != nil {
		return err
	}

	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &getNamespaceCmdData, giterminismManager, common.GetWerfConfigOptions(&getNamespaceCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	namespace, err := deploy_params.GetKubernetesNamespace("", getNamespaceCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	fmt.Println(namespace)

	return nil
}
