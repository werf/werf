package deploy_chart

import (
	"fmt"
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

			if err := common.ApplyLogOptions(&CommonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			return runDeployChart(args[0], args[1])
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupKubeConfig(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)

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

	if err := deploy.Init(*CommonCmdData.KubeContext); err != nil {
		return err
	}

	if err := kube.Init(kube.InitOptions{KubeContext: *CommonCmdData.KubeContext, KubeConfig: *CommonCmdData.KubeConfig}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	namespace := CmdData.Namespace
	if namespace == "" {
		namespace = kube.DefaultNamespace
	}

	werfChart, err := werf_chart.LoadWerfChart(chartDir)
	if err != nil {
		return fmt.Errorf("unable to load chart %s: %s", chartDir, err)
	}

	return werfChart.Deploy(releaseName, namespace, helm.HelmChartOptions{
		Timeout: time.Duration(CmdData.Timeout) * time.Second,
		HelmChartValuesOptions: helm.HelmChartValuesOptions{
			Set:       *CommonCmdData.Set,
			SetString: *CommonCmdData.SetString,
			Values:    *CommonCmdData.Values,
		},
	})
}
