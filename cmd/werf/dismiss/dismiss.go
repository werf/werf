package dismiss

import (
	"fmt"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/docker"
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
		Use:   "dismiss",
		Short: "Delete application from Kubernetes",
		Long: common.GetLongCommandDescription(`Delete application from Kubernetes.

Helm Release will be purged and optionally Kubernetes Namespace.

Environment is a required param for the dismiss by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.

Read more info about Helm Release name, Kubernetes Namespace and how to change it: https://werf.io/reference/deploy/deploy_to_kubernetes.html`),
		Example: `  # Dismiss project named 'myproject' previously deployed app from 'dev' environment; helm release name and namespace will be named as 'myproject-dev'
  $ werf dismiss --env dev --stages-storage :local

  # Dismiss project with namespace
  $ werf dismiss --env my-feature-branch --stages-storage :local --with-namespace

  # Dismiss project using specified helm release name and namespace
  $ werf dismiss --release myrelease --namespace myns --stages-storage :local`,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			common.LogVersion()

			return runDismiss()
		},
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupDir(&CommonCmdData, cmd)

	common.SetupEnvironment(&CommonCmdData, cmd)
	common.SetupRelease(&CommonCmdData, cmd)
	common.SetupNamespace(&CommonCmdData, cmd)

	common.SetupKubeConfig(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "")

	cmd.Flags().BoolVarP(&CmdData.WithNamespace, "with-namespace", "", false, "Delete Kubernetes Namespace after purging Helm Release")

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

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}
	common.LogProjectDir(projectDir)

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("bad config: %s", err)
	}

	err = kube.Init(kube.InitOptions{KubeContext: *CommonCmdData.KubeContext, KubeConfig: *CommonCmdData.KubeConfig})
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

	return deploy.RunDismiss(release, namespace, *CommonCmdData.KubeContext, deploy.DismissOptions{
		WithNamespace: CmdData.WithNamespace,
	})
}
