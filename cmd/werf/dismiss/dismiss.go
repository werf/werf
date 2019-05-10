package dismiss

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	WithNamespace bool
	WithHooks     bool
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

			return common.LogRunningTime(func() error {
				return runDismiss()
			})
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
	common.SetupHelmReleaseStorageNamespace(&CommonCmdData, cmd)
	common.SetupHelmReleaseStorageType(&CommonCmdData, cmd)

	common.SetupDockerConfig(&CommonCmdData, cmd, "")

	common.SetupLogProjectDir(&CommonCmdData, cmd)

	cmd.Flags().BoolVarP(&CmdData.WithNamespace, "with-namespace", "", false, "Delete Kubernetes Namespace after purging Helm Release")
	cmd.Flags().BoolVarP(&CmdData.WithHooks, "with-hooks", "", true, "Delete Helm Release hooks getting from existing revisions")

	return cmd
}

func runDismiss() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	helmReleaseStorageType, err := common.GetHelmReleaseStorageType(*CommonCmdData.HelmReleaseStorageType)
	if err != nil {
		return err
	}

	deployInitOptions := deploy.InitOptions{
		HelmInitOptions: helm.InitOptions{
			KubeConfig:                  *CommonCmdData.KubeConfig,
			KubeContext:                 *CommonCmdData.KubeContext,
			HelmReleaseStorageNamespace: *CommonCmdData.HelmReleaseStorageNamespace,
			HelmReleaseStorageType:      helmReleaseStorageType,
		},
	}
	if err := deploy.Init(deployInitOptions); err != nil {
		return err
	}

	common.LogKubeContext(kube.Context)

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&CommonCmdData, projectDir)

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("bad config: %s", err)
	}

	err = kube.Init(kube.InitOptions{KubeContext: *CommonCmdData.KubeContext, KubeConfig: *CommonCmdData.KubeConfig})
	if err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	if err := common.InitKubedog(); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	release, err := common.GetHelmRelease(*CommonCmdData.Release, *CommonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	namespace, err := common.GetKubernetesNamespace(*CommonCmdData.Namespace, *CommonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	logboek.LogF("Using helm release storage namespace: %s\n", *CommonCmdData.HelmReleaseStorageNamespace)
	logboek.LogF("Using helm release storage type: %s\n", helmReleaseStorageType)
	logboek.LogF("Using helm release name: %s\n", release)
	logboek.LogF("Using kubernetes namespace: %s\n", namespace)

	return deploy.RunDismiss(release, namespace, *CommonCmdData.KubeContext, deploy.DismissOptions{
		WithNamespace: CmdData.WithNamespace,
		WithHooks:     CmdData.WithHooks,
	})
}
