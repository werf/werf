package apply

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/werf/werf/pkg/deploy/lock_manager"

	"github.com/werf/werf/pkg/deploy/werf_chart"

	"helm.sh/helm/v3/pkg/cli/values"

	"github.com/werf/kubedog/pkg/kube"

	uuid "github.com/satori/go.uuid"

	"github.com/werf/werf/pkg/werf/global_warnings"

	"github.com/werf/werf/pkg/deploy/helm"

	cmd_helm "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	Tag          string
	Timeout      int
	AutoRollback bool
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "apply",
		Short:                 "Apply bundle into Kubernetes",
		Long:                  common.GetLongCommandDescription(`Take latest bundle from the specified container registry using specified version tag or version mask and apply it as a helm chart into Kubernetes cluster.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs, common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			defer global_warnings.PrintGlobalWarnings(common.BackgroundContext())

			logboek.Streams().Mute()
			logboek.SetAcceptedLevel(level.Error)

			if err := common.ProcessLogOptionsDefaultQuiet(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runApply()
			})
		},
	}

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupStagesStorageOptions(&commonCmdData, cmd) // FIXME

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd)
	common.SetupSecretValues(&commonCmdData, cmd)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupRelease(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd)
	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Provide exact tag version or semver-based pattern, werf will install or upgrade to the latest version of the specified bundle ($WERF_TAG or latest by default)")

	return cmd
}

func runApply() error {
	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := common.DockerRegistryInit(&commonCmdData); err != nil {
		return err
	}

	common.SetupOndemandKubeInitializer(*commonCmdData.KubeContext, *commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64)
	if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
		return err
	}

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return err
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return err
	}

	repoAddress, err := common.GetStagesStorageAddress(&commonCmdData)
	if err != nil {
		return err
	}

	cmd_helm.Settings.Debug = *commonCmdData.LogDebug

	actionConfig := new(action.Configuration)
	if err := helm.InitActionConfig(ctx, common.GetOndemandKubeInitializer(), *commonCmdData.Namespace, cmd_helm.Settings, actionConfig, helm.InitActionConfigOptions{
		StatusProgressPeriod:      time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
		HooksStatusProgressPeriod: time.Duration(*commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second,
		KubeConfigOptions: kube.KubeConfigOptions{
			Context:          *commonCmdData.KubeContext,
			ConfigPath:       *commonCmdData.KubeConfig,
			ConfigDataBase64: *commonCmdData.KubeConfigBase64,
		},
		ReleasesHistoryMax: *commonCmdData.ReleasesHistoryMax,
	}); err != nil {
		return err
	}

	loader.GlobalLoadOptions = &loader.LoadOptions{}

	// FIXME: support semver-pattern
	bundleRef := fmt.Sprintf("%s:%s", repoAddress, cmdData.Tag)

	bundleTmpDir := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewV4().String())
	defer os.RemoveAll(bundleTmpDir)

	if err := logboek.Context(ctx).LogProcess("Pulling bundle %q", bundleRef).DoError(func() error {
		if cmd := cmd_helm.NewChartPullCmd(actionConfig, logboek.ProxyOutStream()); cmd != nil {
			if err := cmd.RunE(cmd, []string{bundleRef}); err != nil {
				return fmt.Errorf("error saving bundle to the local chart helm cache: %s", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Exporting bundle %q", bundleRef).DoError(func() error {
		if cmd := cmd_helm.NewChartExportCmd(actionConfig, logboek.ProxyOutStream(), cmd_helm.ChartExportCmdOptions{Destination: bundleTmpDir}); cmd != nil {
			if err := cmd.RunE(cmd, []string{bundleRef}); err != nil {
				return fmt.Errorf("error pushing bundle %q: %s", bundleRef, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	namespace := common.GetNamespace(&commonCmdData)
	releaseName, err := common.GetRequiredRelease(&commonCmdData)
	if err != nil {
		return err
	}

	var lockManager *lock_manager.LockManager
	if m, err := lock_manager.NewLockManager(namespace); err != nil {
		return fmt.Errorf("unable to create lock manager: %s", err)
	} else {
		lockManager = m
	}

	bundle := werf_chart.NewBundle(bundleTmpDir, lockManager)

	postRenderer, err := bundle.GetPostRenderer()
	if err != nil {
		return err
	}

	postRenderer.Add(userExtraAnnotations, userExtraLabels)
	if *commonCmdData.Environment != "" {
		postRenderer.Add(map[string]string{"project.werf.io/env": *commonCmdData.Environment}, nil)
	}

	loader.GlobalLoadOptions = &loader.LoadOptions{
		ChartExtender: bundle,
	}

	helmUpgradeCmd, _ := cmd_helm.NewUpgradeCmd(actionConfig, logboek.ProxyOutStream(), cmd_helm.UpgradeCmdOptions{
		PostRenderer: postRenderer,
		ValueOpts: &values.Options{
			ValueFiles:   *commonCmdData.Values,
			StringValues: *commonCmdData.SetString,
			Values:       *commonCmdData.Set,
			FileValues:   *commonCmdData.SetFile,
		},
		CreateNamespace: common.NewBool(true),
		Install:         common.NewBool(true),
		Wait:            common.NewBool(true),
		Atomic:          common.NewBool(cmdData.AutoRollback),
		Timeout:         common.NewDuration(time.Duration(cmdData.Timeout) * time.Second),
	})

	/*
	 * TODO: rework WerfChart and Bundle, use shared common code
	 */

	return bundle.WrapUpgrade(ctx, releaseName, func() error {
		return helmUpgradeCmd.RunE(helmUpgradeCmd, []string{releaseName, bundle.Dir})
	})
}
