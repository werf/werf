package main

import (
	"fmt"
	"os"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/kubedog/pkg/kube"
	"github.com/spf13/cobra"
)

var dismissCmdData struct {
	HelmReleaseName string

	Namespace     string
	KubeContext   string
	WithNamespace bool
}

func newDismissCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "dismiss HELM_RELEASE_NAME",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dismissCmdData.HelmReleaseName = args[0]

			err := runDismiss()
			if err != nil {
				return fmt.Errorf("dismiss failed: %s", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&dismissCmdData.Namespace, "namespace", "", "", "Kubernetes namespace")
	cmd.PersistentFlags().StringVarP(&dismissCmdData.KubeContext, "kube-context", "", "", "Kubernetes config context")
	cmd.PersistentFlags().BoolVarP(&dismissCmdData.WithNamespace, "with-namespace", "", false, "Delete namespace after purging helm release")

	return cmd
}

func runDismiss() error {
	if err := dapp.Init(rootCmdData.TmpDir, rootCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	kubeContext := os.Getenv("KUBECONTEXT")
	if kubeContext == "" {
		kubeContext = dismissCmdData.KubeContext
	}
	err := kube.Init(kube.InitOptions{KubeContext: kubeContext})
	if err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	namespace := getNamespace(dismissCmdData.Namespace)

	return deploy.RunDismiss(dismissCmdData.HelmReleaseName, namespace, kubeContext, deploy.DismissOptions{
		WithNamespace: dismissCmdData.WithNamespace,
	})
}
