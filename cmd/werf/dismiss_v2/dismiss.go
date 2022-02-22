package dismiss_v2

import (
	"context"
	"fmt"

	cmd_helm "helm.sh/helm/v3/cmd/helm"

	"github.com/werf/werf/pkg/deploy_v2/lock_manager"
	"github.com/werf/werf/pkg/storage/lrumeta"

	"github.com/werf/werf/pkg/deploy_v2/werf_chart"

	"github.com/spf13/cobra"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	WithNamespace bool
	WithHooks     bool
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dismiss",
		Short: "Delete application from Kubernetes",
		Long: common.GetLongCommandDescription(`Delete application from Kubernetes.

Helm Release will be purged and optionally Kubernetes Namespace.

Environment is a required param for the dismiss by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command`),
		Example: `  # Dismiss project named 'myproject' previously deployed app from 'dev' environment; helm release name and namespace will be named as 'myproject-dev'
  $ werf dismiss --env dev

  # Dismiss project with namespace
  $ werf dismiss --env my-feature-branch --with-namespace

  # Dismiss project using specified helm release name and namespace
  $ werf dismiss --release myrelease --namespace myns`,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer werf.PrintGlobalWarnings(common.BackgroundContext())

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runDismiss()
			})
		},
	}

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupDir(&commonCmdData, cmd)

	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupRelease(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "")

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.WithNamespace, "with-namespace", "", common.GetBoolEnvironmentDefaultFalse("WERF_WITH_NAMESPACE"), "Delete Kubernetes Namespace after purging Helm Release (default $WERF_WITH_NAMESPACE)")
	cmd.Flags().BoolVarP(&cmdData.WithHooks, "with-hooks", "", common.GetBoolEnvironmentDefaultTrue("WERF_WITH_HOOKS"), "Delete Helm Release hooks getting from existing revisions (default $WERF_WITH_HOOKS or true)")

	return cmd
}

func runDismiss() error {
	tmp_manager.AutoGCEnabled = true
	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := lrumeta.Init(); err != nil {
		return err
	}

	common.LogKubeContext(kube.Context)

	if err := docker.Init(ctx, *commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return err
	}
	ctx = ctxWithDockerCli

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&commonCmdData, projectDir)

	werfConfig, err := common.GetRequiredWerfConfig(ctx, projectDir, &commonCmdData, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}
	logboek.LogOptionalLn()

	err = kube.Init(kube.InitOptions{kube.KubeConfigOptions{
		Context:             *commonCmdData.KubeContext,
		ConfigPath:          *commonCmdData.KubeConfig,
		ConfigDataBase64:    *commonCmdData.KubeConfigBase64,
		ConfigPathMergeList: *commonCmdData.KubeConfigPathMergeList,
	}})
	if err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	if err := common.InitKubedog(ctx); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	releaseName, err := common.GetHelmRelease(*commonCmdData.Release, *commonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	namespace, err := common.GetKubernetesNamespace(*commonCmdData.Namespace, *commonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	var lockManager *lock_manager.LockManager
	if m, err := lock_manager.NewLockManager(namespace); err != nil {
		return fmt.Errorf("unable to create lock manager: %s", err)
	} else {
		lockManager = m
	}

	wc := werf_chart.NewWerfChart(werf_chart.WerfChartOptions{
		ReleaseName: releaseName,
		LockManager: lockManager,
	})

	actionConfig, err := common.NewActionConfig(ctx, namespace, &commonCmdData)
	if err != nil {
		return err
	}
	helmUninstallCmd := cmd_helm.NewUninstallCmd(actionConfig, logboek.ProxyOutStream())

	if err := wc.SetEnv(*commonCmdData.Environment); err != nil {
		return err
	}

	if err := wc.SetWerfConfig(werfConfig); err != nil {
		return err
	}

	return wc.WrapUninstall(context.Background(), func() error {
		return helmUninstallCmd.RunE(helmUninstallCmd, []string{releaseName})
	})
}
