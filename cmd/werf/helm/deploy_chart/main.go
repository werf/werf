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
	Values    []string
	Set       []string
	SetString []string
	Namespace string
	Timeout   int
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-chart PATH RELEASE_NAME",
		Short: "Deploy Helm chart specified by path",
		Long: common.GetLongCommandDescription(`Deploy Helm chart specified by path.

If specified Helm chart is a Werf chart with additional values and contains werf-chart.yaml, then werf will pass all additinal values and data into helm.`),
		DisableFlagsInUseLine: true,
		Args: cobra.MinimumNArgs(2),
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfHome, common.WerfTmp),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runDeployChart(args[0], args[1]); err != nil {
				return fmt.Errorf("deploy-chart failed: %s", err)
			}

			return nil
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	cmd.Flags().StringArrayVarP(&CmdData.Values, "values", "", []string{}, "Additional helm values")
	cmd.Flags().StringArrayVarP(&CmdData.Set, "set", "", []string{}, "Additional helm sets")
	cmd.Flags().StringArrayVarP(&CmdData.SetString, "set-string", "", []string{}, "Additional helm STRING sets")
	cmd.Flags().StringVarP(&CmdData.Namespace, "namespace", "", "", "Namespace to install release into")
	cmd.Flags().IntVarP(&CmdData.Timeout, "timeout", "t", 0, "Resources tracking timeout in seconds")

	common.SetupKubeContext(&CommonCmdData, cmd)

	return cmd
}

func runDeployChart(chartDir string, releaseName string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := deploy.Init(); err != nil {
		return err
	}

	kubeContext := common.GetKubeContext(*CommonCmdData.KubeContext)
	if err := kube.Init(kube.InitOptions{KubeContext: kubeContext}); err != nil {
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
		CommonHelmOptions: helm.CommonHelmOptions{KubeContext: kubeContext},
		Timeout:           time.Duration(CmdData.Timeout) * time.Second,
		Set:               CmdData.Set,
		SetString:         CmdData.SetString,
		Values:            CmdData.Values,
	})
}
