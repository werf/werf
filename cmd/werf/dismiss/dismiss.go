package dismiss

import (
	"fmt"
	"os"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

var CmdData struct {
	WithNamespace bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "dismiss",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runDismiss()
			if err != nil {
				return fmt.Errorf("dismiss failed: %s", err)
			}

			return nil
		},
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupDir(&CommonCmdData, cmd)

	cmd.PersistentFlags().BoolVarP(&CmdData.WithNamespace, "with-namespace", "", false, "Delete Kubernetes Namespace after purging Helm Release")

	common.SetupEnvironment(&CommonCmdData, cmd)
	common.SetupRelease(&CommonCmdData, cmd)
	common.SetupNamespace(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)

	return cmd
}

func runDismiss() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := deploy.Init(); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("cannot parse werf config: %s", err)
	}

	kubeContext := os.Getenv("KUBECONTEXT")
	if kubeContext == "" {
		kubeContext = *CommonCmdData.KubeContext
	}
	err = kube.Init(kube.InitOptions{KubeContext: kubeContext})
	if err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	release, err := common.GetHelmRelease(*CommonCmdData.Release, *CommonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	namespace, err := common.GetKubernetesNamespace(*CommonCmdData.Namespace, *CommonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	return deploy.RunDismiss(release, namespace, kubeContext, deploy.DismissOptions{
		WithNamespace: CmdData.WithNamespace,
	})
}
