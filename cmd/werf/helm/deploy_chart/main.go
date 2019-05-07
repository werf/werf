package deploy_chart

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/deploy/werf_chart"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

var CmdData struct {
	Namespace string
	Timeout   int
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-chart PATH RELEASE_NAME",
		Short: "Deploy Helm chart specified by path",
		Long: common.GetLongCommandDescription(`Deploy Helm chart specified by path.

If specified Helm chart is a Werf chart with additional values and contains werf-chart.yaml, then werf will pass all additinal values and data into helm`),
		Example: `  # Deploy raw helm chart from current directory
  $ werf helm deploy-chart . myrelease`,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ValidateArgumentCount(2, args, cmd); err != nil {
				return err
			}

			if err := common.ProcessLogOptions(&CommonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return common.LogRunningTime(func() error {
				return runDeployChart(args[0], args[1])
			})
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupKubeConfig(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)
	common.SetupHelmReleaseStorageNamespace(&CommonCmdData, cmd)
	common.SetupHelmReleaseStorageType(&CommonCmdData, cmd)

	common.SetupLogOptions(&CommonCmdData, cmd)

	common.SetupSet(&CommonCmdData, cmd)
	common.SetupSetString(&CommonCmdData, cmd)
	common.SetupValues(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.Namespace, "namespace", "", "", "Namespace to install release into")
	cmd.Flags().IntVarP(&CmdData.Timeout, "timeout", "t", 0, "Resources tracking timeout in seconds")

	return cmd
}

func runDeployChart(chartDir string, releaseName string) error {
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

	if err := kube.Init(kube.InitOptions{KubeContext: *CommonCmdData.KubeContext, KubeConfig: *CommonCmdData.KubeConfig}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	common.LogKubeContext(kube.Context)

	if err := common.InitKubedog(); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	namespace := CmdData.Namespace
	if namespace == "" {
		namespace = kube.DefaultNamespace
	}

	werfChart, err := werf_chart.LoadWerfChart(chartDir)
	if err != nil {
		return fmt.Errorf("unable to load chart %s: %s", chartDir, err)
	}

	if err := werfChart.Deploy(releaseName, namespace, helm.ChartOptions{
		Timeout: time.Duration(CmdData.Timeout) * time.Second,
		ChartValuesOptions: helm.ChartValuesOptions{
			Set:       *CommonCmdData.Set,
			SetString: *CommonCmdData.SetString,
			Values:    *CommonCmdData.Values,
		},
	}); err != nil {
		replaceOld := fmt.Sprintf("%s/", werfChart.Name)
		replaceNew := fmt.Sprintf("%s/", strings.TrimRight(werfChart.ChartDir, "/"))
		errMsg := strings.Replace(err.Error(), replaceOld, replaceNew, -1)
		return errors.New(errMsg)
	}

	return nil
}
