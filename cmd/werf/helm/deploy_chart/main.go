package deploy_chart

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/werf/werf/pkg/image"

	"github.com/spf13/cobra"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"

	"github.com/werf/werf/cmd/werf/common"
	helm_common "github.com/werf/werf/cmd/werf/helm/common"
	"github.com/werf/werf/pkg/deploy"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/werf_chart"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	Namespace string
	Timeout   int
}

var commonCmdData common.CmdData
var helmCmdData helm_common.HelmCmdData
var downloadChartOptions helm_common.DownloadChartOptions

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

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return common.LogRunningTime(func() error {
				return runDeployChart(args[0], args[1])
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageNamespace(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageType(&commonCmdData, cmd)
	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd)

	common.SetupThreeWayMergeMode(&commonCmdData, cmd)

	helm_common.SetupHelmHome(&helmCmdData, cmd)

	f := cmd.Flags()
	f.StringVarP(&cmdData.Namespace, "namespace", "", "", "Namespace to install release into")
	f.IntVarP(&cmdData.Timeout, "timeout", "t", 0, "Resources tracking timeout in seconds")

	downloadChartOptionsExtraUsage := " (if using CHART as a chart reference)"

	f.BoolVar(&downloadChartOptions.Verify, "verify", false, "verify the package against its signature"+downloadChartOptionsExtraUsage)
	f.BoolVar(&downloadChartOptions.VerifyLater, "prov", false, "fetch the provenance file, but don't perform verification"+downloadChartOptionsExtraUsage)
	f.StringVar(&downloadChartOptions.Version, "version", "", "specific version of a chart. Without this, the latest version is fetched"+downloadChartOptionsExtraUsage)
	f.StringVar(&downloadChartOptions.Keyring, "keyring", helm_common.DefaultKeyring(), "keyring containing public keys"+downloadChartOptionsExtraUsage)
	f.StringVar(&downloadChartOptions.RepoURL, "repo", "", "chart repository url where to locate the requested chart"+downloadChartOptionsExtraUsage)
	f.StringVar(&downloadChartOptions.CertFile, "cert-file", "", "identify HTTPS client using this SSL certificate file"+downloadChartOptionsExtraUsage)
	f.StringVar(&downloadChartOptions.KeyFile, "key-file", "", "identify HTTPS client using this SSL key file"+downloadChartOptionsExtraUsage)
	f.StringVar(&downloadChartOptions.CaFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle"+downloadChartOptionsExtraUsage)
	f.BoolVar(&downloadChartOptions.Devel, "devel", false, "use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored"+downloadChartOptionsExtraUsage)
	f.StringVar(&downloadChartOptions.Username, "username", "", "chart repository username"+downloadChartOptionsExtraUsage)
	f.StringVar(&downloadChartOptions.Password, "password", "", "chart repository password"+downloadChartOptionsExtraUsage)

	return cmd
}

func runDeployChart(chartDirOrChartReference string, releaseName string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := image.Init(); err != nil {
		return err
	}

	helmReleaseStorageType, err := common.GetHelmReleaseStorageType(*commonCmdData.HelmReleaseStorageType)
	if err != nil {
		return err
	}

	threeWayMergeMode, err := common.GetThreeWayMergeMode(*commonCmdData.ThreeWayMergeMode)
	if err != nil {
		return err
	}

	deployInitOptions := deploy.InitOptions{
		HelmInitOptions: helm.InitOptions{
			KubeConfig:                  *commonCmdData.KubeConfig,
			KubeContext:                 *commonCmdData.KubeContext,
			HelmReleaseStorageNamespace: *commonCmdData.HelmReleaseStorageNamespace,
			HelmReleaseStorageType:      helmReleaseStorageType,
			StatusProgressPeriod:        common.GetStatusProgressPeriod(&commonCmdData),
			HooksStatusProgressPeriod:   common.GetHooksStatusProgressPeriod(&commonCmdData),
			InitNamespace:               true,
		},
	}
	if err := deploy.Init(deployInitOptions); err != nil {
		return err
	}

	if err := kube.Init(kube.InitOptions{KubeContext: *commonCmdData.KubeContext, KubeConfig: *commonCmdData.KubeConfig}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	common.LogKubeContext(kube.Context)

	if err := common.InitKubedog(); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	namespace := cmdData.Namespace
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
			helm_common.InitHelmSettings(&helmCmdData)

			destDir, err := tmp_manager.CreateHelmTmpChartDestDir()
			if err != nil {
				return err
			}
			defer os.RemoveAll(destDir)

			downloadChartOptions.ChartRef = chartDirOrChartReference
			downloadChartOptions.DestDir = destDir
			downloadChartOptions.Untar = true

			if err := helm_common.DownloadChart(&downloadChartOptions); err != nil {
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
		Timeout: time.Duration(cmdData.Timeout) * time.Second,
		ChartValuesOptions: helm.ChartValuesOptions{
			Set:       *commonCmdData.Set,
			SetString: *commonCmdData.SetString,
			Values:    *commonCmdData.Values,
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
