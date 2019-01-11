package dismiss

import (
	"fmt"
	"os"

	"github.com/flant/dapp/cmd/dapp/common"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/kubedog/pkg/kube"
	"github.com/spf13/cobra"
)

var CmdData struct {
	HelmReleaseName string

	Namespace     string
	KubeContext   string
	WithNamespace bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Short: "<COMMAND DESCRIPTION HERE>",
		Use:   "dismiss HELM_RELEASE_NAME",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			CmdData.HelmReleaseName = args[0]

			err := runDismiss()
			if err != nil {
				return fmt.Errorf("dismiss failed: %s", err)
			}

			return nil
		},
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	cmd.PersistentFlags().StringVarP(&CmdData.Namespace, "namespace", "", "", "Kubernetes namespace")
	cmd.PersistentFlags().StringVarP(&CmdData.KubeContext, "kube-context", "", "", "Kubernetes config context")
	cmd.PersistentFlags().BoolVarP(&CmdData.WithNamespace, "with-namespace", "", false, "Delete namespace after purging helm release")

	return cmd
}

func runDismiss() error {
	if err := dapp.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := deploy.Init(); err != nil {
		return err
	}

	kubeContext := os.Getenv("KUBECONTEXT")
	if kubeContext == "" {
		kubeContext = CmdData.KubeContext
	}
	err := kube.Init(kube.InitOptions{KubeContext: kubeContext})
	if err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	namespace := common.GetNamespace(CmdData.Namespace)

	return deploy.RunDismiss(CmdData.HelmReleaseName, namespace, kubeContext, deploy.DismissOptions{
		WithNamespace: CmdData.WithNamespace,
	})
}
