package helm

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var getNamespaceCmdData common.CmdData

func NewGetNamespaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "get-namespace",
		DisableFlagsInUseLine: true,
		Short:                 "Print Kubernetes Namespace that will be used in current configuration with specified params",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&getNamespaceCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runGetNamespace()
		},
	}

	common.SetupDir(&getNamespaceCmdData, cmd)
	common.SetupGitWorkTree(&getNamespaceCmdData, cmd)
	common.SetupConfigTemplatesDir(&getNamespaceCmdData, cmd)
	common.SetupConfigPath(&getNamespaceCmdData, cmd)
	common.SetupEnvironment(&getNamespaceCmdData, cmd)

	common.SetupGiterminismInspectorOptions(&getNamespaceCmdData, cmd)

	common.SetupTmpDir(&getNamespaceCmdData, cmd)
	common.SetupHomeDir(&getNamespaceCmdData, cmd)
	common.SetupDockerConfig(&getNamespaceCmdData, cmd, "")

	common.SetupLogOptions(&getNamespaceCmdData, cmd)

	return cmd
}

func runGetNamespace() error {
	if err := werf.Init(*getNamespaceCmdData.TmpDir, *getNamespaceCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := common.InitGiterminismInspector(&getNamespaceCmdData); err != nil {
		return err
	}

	if err := git_repo.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{LiveGitOutput: *getNamespaceCmdData.LogVerbose || *getNamespaceCmdData.LogDebug}); err != nil {
		return err
	}

	giterminismManager, err := common.GetGiterminismManager(&getNamespaceCmdData)
	if err != nil {
		return err
	}

	werfConfig, err := common.GetRequiredWerfConfig(common.BackgroundContext(), &getNamespaceCmdData, giterminismManager, common.GetWerfConfigOptions(&getNamespaceCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	namespace, err := common.GetKubernetesNamespace("", *getNamespaceCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	fmt.Println(namespace)

	return nil
}
