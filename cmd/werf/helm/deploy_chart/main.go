package deploy_chart

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"

	"github.com/flant/werf/cmd/werf/common"
	helm_common "github.com/flant/werf/cmd/werf/helm/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/deploy/werf_chart"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	Namespace string
	Timeout   int
}

var CommonCmdData common.CmdData
var HelmCmdData helm_common.HelmCmdData
var DownloadChartOptions helm_common.DownloadChartOptions

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-chart CHART_DIR|CHART_REFERENCE RELEASE_NAME",
		Short: "Deploy Helm chart specified by path",
		Example: `  # Deploy raw helm chart from current directory
  $ werf helm deploy-chart . myrelease

  # Deploy helm chart by chart reference
  $ werf helm deploy-chart stable/nginx-ingress myrelease
`,
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
	common.SetupStatusProgressPeriod(&CommonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&CommonCmdData, cmd)

	common.SetupLogOptions(&CommonCmdData, cmd)

	common.SetupSet(&CommonCmdData, cmd)
	common.SetupSetString(&CommonCmdData, cmd)
	common.SetupValues(&CommonCmdData, cmd)

	common.SetupThreeWayMergeMode(&CommonCmdData, cmd)

	helm_common.SetupHelmHome(&HelmCmdData, cmd)

	f := cmd.Flags()
	f.StringVarP(&CmdData.Namespace, "namespace", "", "", "Namespace to install release into")
	f.IntVarP(&CmdData.Timeout, "timeout", "t", 0, "Resources tracking timeout in seconds")

	downloadChartOptionsExtraUsage := " (if using CHART as a chart reference)"

	f.BoolVar(&DownloadChartOptions.Verify, "verify", false, "verify the package against its signature"+downloadChartOptionsExtraUsage)
	f.BoolVar(&DownloadChartOptions.VerifyLater, "prov", false, "fetch the provenance file, but don't perform verification"+downloadChartOptionsExtraUsage)
	f.StringVar(&DownloadChartOptions.Version, "version", "", "specific version of a chart. Without this, the latest version is fetched"+downloadChartOptionsExtraUsage)
	f.StringVar(&DownloadChartOptions.Keyring, "keyring", helm_common.DefaultKeyring(), "keyring containing public keys"+downloadChartOptionsExtraUsage)
	f.StringVar(&DownloadChartOptions.RepoURL, "repo", "", "chart repository url where to locate the requested chart"+downloadChartOptionsExtraUsage)
	f.StringVar(&DownloadChartOptions.CertFile, "cert-file", "", "identify HTTPS client using this SSL certificate file"+downloadChartOptionsExtraUsage)
	f.StringVar(&DownloadChartOptions.KeyFile, "key-file", "", "identify HTTPS client using this SSL key file"+downloadChartOptionsExtraUsage)
	f.StringVar(&DownloadChartOptions.CaFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle"+downloadChartOptionsExtraUsage)
	f.BoolVar(&DownloadChartOptions.Devel, "devel", false, "use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored"+downloadChartOptionsExtraUsage)
	f.StringVar(&DownloadChartOptions.Username, "username", "", "chart repository username"+downloadChartOptionsExtraUsage)
	f.StringVar(&DownloadChartOptions.Password, "password", "", "chart repository password"+downloadChartOptionsExtraUsage)

	return cmd
}

func runDeployChart(chartDirOrChartReference string, releaseName string) error {
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

	threeWayMergeMode, err := common.GetThreeWayMergeMode(*CommonCmdData.ThreeWayMergeMode)
	if err != nil {
		return err
	}

	deployInitOptions := deploy.InitOptions{
		HelmInitOptions: helm.InitOptions{
			KubeConfig:                  *CommonCmdData.KubeConfig,
			KubeContext:                 *CommonCmdData.KubeContext,
			HelmReleaseStorageNamespace: *CommonCmdData.HelmReleaseStorageNamespace,
			HelmReleaseStorageType:      helmReleaseStorageType,
			StatusProgressPeriod:        common.GetStatusProgressPeriod(&CommonCmdData),
			HooksStatusProgressPeriod:   common.GetHooksStatusProgressPeriod(&CommonCmdData),
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

	exist, err := util.DirExists(chartDirOrChartReference)
	if err != nil {
		return err
	}

	var chartDir string
	if exist {
		chartDir = chartDirOrChartReference
	} else {
		chartReferenceParts := strings.Split(chartDirOrChartReference, "/")
		if len(chartReferenceParts) == 2 {
			helm_common.InitHelmSettings(&HelmCmdData)

			destDir, err := tmp_manager.CreateHelmTmpChartDestDir()
			if err != nil {
				return err
			}
			defer os.RemoveAll(destDir)

			DownloadChartOptions.ChartRef = chartDirOrChartReference
			DownloadChartOptions.DestDir = destDir
			DownloadChartOptions.Untar = true

			if err := helm_common.DownloadChart(&DownloadChartOptions); err != nil {
				return fmt.Errorf("\n- chart directory %[1]s is not found\n- unable to download chart %[1]s: %s", chartDirOrChartReference, err)
			}

			chartDir = filepath.Join(destDir, chartReferenceParts[1])
		} else {
			return fmt.Errorf("chart directory %s is not found", chartDirOrChartReference)
		}
	}

	logboek.LogOptionalLn()
	werfChart := &werf_chart.WerfChart{ChartDir: chartDir}
	if err := werfChart.Deploy(releaseName, namespace, helm.ChartOptions{
		Timeout: time.Duration(CmdData.Timeout) * time.Second,
		ChartValuesOptions: helm.ChartValuesOptions{
			Set:       *CommonCmdData.Set,
			SetString: *CommonCmdData.SetString,
			Values:    *CommonCmdData.Values,
		},
		ThreeWayMergeMode: threeWayMergeMode,
	}); err != nil {
		replaceOld := fmt.Sprintf("%s/", werfChart.Name)
		replaceNew := fmt.Sprintf("%s/", strings.TrimRight(werfChart.ChartDir, "/"))
		errMsg := strings.Replace(err.Error(), replaceOld, replaceNew, -1)
		return errors.New(errMsg)
	}

	return nil
}
