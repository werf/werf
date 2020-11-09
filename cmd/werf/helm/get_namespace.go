package helm

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/git_repo"
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
	common.SetupDisableDeterminism(&commonCmdData, cmd)
	common.SetupConfigPath(&getNamespaceCmdData, cmd)
	common.SetupConfigTemplatesDir(&getNamespaceCmdData, cmd)
	common.SetupTmpDir(&getNamespaceCmdData, cmd)
	common.SetupHomeDir(&getNamespaceCmdData, cmd)
	common.SetupEnvironment(&getNamespaceCmdData, cmd)
	common.SetupDockerConfig(&getNamespaceCmdData, cmd, "")

	common.SetupLogOptions(&getNamespaceCmdData, cmd)

	return cmd
}

func runGetNamespace() error {
	if err := werf.Init(*getNamespaceCmdData.TmpDir, *getNamespaceCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	projectDir, err := common.GetProjectDir(&getNamespaceCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	localGitRepo, err := git_repo.OpenLocalRepo("own", projectDir)
	if err != nil {
		return fmt.Errorf("unable to open local repo %s: %s", projectDir, err)
	}

	werfConfig, err := common.GetRequiredWerfConfig(common.BackgroundContext(), projectDir, &getNamespaceCmdData, localGitRepo, config.WerfConfigOptions{DisableDeterminism: *commonCmdData.DisableDeterminism})
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
