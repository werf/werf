package apply

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/3p-helm/pkg/engine"
	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/3p-helm/pkg/werf/helmopts"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	Tag          string
	Timeout      int
	AutoRollback bool
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "apply",
		Short:                 "Apply bundle into Kubernetes",
		Long:                  common.GetLongCommandDescription(`Take latest bundle from the specified container registry using specified version tag or version mask and apply it as a helm chart into Kubernetes cluster.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runApply(ctx) })
		},
	})

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupGiterminismOptions(&commonCmdData, cmd)
	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)

	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, false)
	common.SetupSkipTLSVerifyKube(&commonCmdData, cmd)
	common.SetupKubeApiServer(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyHelmDependencies(&commonCmdData, cmd)
	common.SetupKubeCaPath(&commonCmdData, cmd)
	common.SetupKubeTlsServer(&commonCmdData, cmd)
	common.SetupKubeToken(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupNoPodLogs(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd, false)
	common.SetupSecretValues(&commonCmdData, cmd, false)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	commonCmdData.SetupDisableDefaultValues(cmd)
	commonCmdData.SetupDisableDefaultSecretValues(cmd)
	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)

	common.SetupSaveDeployReport(&commonCmdData, cmd)
	common.SetupDeployReportPath(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupRelease(&commonCmdData, cmd, false)
	common.SetupNamespace(&commonCmdData, cmd, false)
	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupKubeQpsLimit(&commonCmdData, cmd)
	common.SetupKubeBurstLimit(&commonCmdData, cmd)
	common.SetupDeployGraphPath(&commonCmdData, cmd)
	common.SetupRollbackGraphPath(&commonCmdData, cmd)

	common.SetupRenderSubchartNotes(&commonCmdData, cmd)
	common.SetupNoInstallCRDs(&commonCmdData, cmd)
	common.SetupReleaseLabel(&commonCmdData, cmd)
	common.SetupForceAdoption(&commonCmdData, cmd)

	commonCmdData.SetupDebugTemplates(cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Provide exact tag version or semver-based pattern, werf will install or upgrade to the latest version of the specified bundle ($WERF_TAG or latest by default)")

	defaultTimeout, err := util.GetIntEnvVar("WERF_TIMEOUT")
	if err != nil || defaultTimeout == nil {
		defaultTimeout = new(int64)
	}
	cmd.Flags().IntVarP(&cmdData.Timeout, "timeout", "t", int(*defaultTimeout), "Resources tracking timeout in seconds ($WERF_TIMEOUT by default)")

	cmd.Flags().BoolVarP(&cmdData.AutoRollback, "auto-rollback", "R", util.GetBoolEnvironmentDefaultFalse("WERF_AUTO_ROLLBACK"), "Enable auto rollback of the failed release to the previous deployed release version when current deploy process have failed ($WERF_AUTO_ROLLBACK by default)")
	cmd.Flags().BoolVarP(&cmdData.AutoRollback, "atomic", "", util.GetBoolEnvironmentDefaultFalse("WERF_ATOMIC"), "Enable auto rollback of the failed release to the previous deployed release version when current deploy process have failed ($WERF_ATOMIC by default)")

	return cmd
}

func runApply(ctx context.Context) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning(ctx)

	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                &commonCmdData,
		InitDockerRegistry: true,
		InitWerf:           true,
		InitGitDataManager: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	repoAddress, err := commonCmdData.Repo.GetAddress()
	if err != nil {
		return fmt.Errorf("get repo address: %w", err)
	}

	releaseNamespace := common.GetNamespace(&commonCmdData)
	releaseName, err := common.GetRequiredRelease(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get release name: %w", err)
	}

	var deployReportPath string
	if common.GetSaveDeployReport(&commonCmdData) {
		deployReportPath, err = common.GetDeployReportPath(&commonCmdData)
		if err != nil {
			return fmt.Errorf("get deploy report path: %w", err)
		}
	}

	bundlePath := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
	defer os.RemoveAll(bundlePath)

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return fmt.Errorf("construct bundles registry client: %w", err)
	}

	registryCredentialsPath := docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig)

	serviceValues, err := helpers.GetBundleServiceValues(ctx, helpers.ServiceValuesOptions{
		Env:                      *commonCmdData.Environment,
		Namespace:                releaseNamespace,
		SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
		DockerConfigPath:         filepath.Dir(registryCredentialsPath),
	})
	if err != nil {
		return fmt.Errorf("get service values: %w", err)
	}

	secretWorkDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current working directory: %w", err)
	}

	if err := bundles.Pull(ctx, fmt.Sprintf("%s:%s", repoAddress, cmdData.Tag), bundlePath, bundlesRegistryClient, helmopts.HelmOptions{
		ChartLoadOpts: helmopts.ChartLoadOptions{
			NoDecryptSecrets:      *commonCmdData.IgnoreSecretKey,
			NoDefaultSecretValues: *commonCmdData.DisableDefaultSecretValues,
			NoDefaultValues:       *commonCmdData.DisableDefaultValues,
			SecretValuesFiles:     common.GetSecretValues(&commonCmdData),
			SecretsWorkingDir:     secretWorkDir,
			ExtraValues:           serviceValues,
		},
	}); err != nil {
		return fmt.Errorf("pull bundle: %w", err)
	}

	serviceAnnotations, extraAnnotations, extraLabels, err := getAnnotationsAndLabels(bundlePath)
	if err != nil {
		return fmt.Errorf("get annotations and labels: %w", err)
	}

	releaseLabels, err := common.GetReleaseLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get release labels: %w", err)
	}

	ctx = action.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultReleaseInstallLogLevel), action.SetupLoggingOptions{
		ColorMode: *commonCmdData.LogColorMode,
	})
	engine.Debug = *commonCmdData.DebugTemplates

	// Для bundle команд не нужен giterminism manager, используем простой file reader
	bundleFileReader := common.NewBundleFileReader(bundlePath)
	file.ChartFileReader = bundleFileReader

	if err := action.ReleaseInstall(ctx, releaseName, releaseNamespace, action.ReleaseInstallOptions{
		AutoRollback:                 cmdData.AutoRollback,
		ChartDirPath:                 bundlePath,
		ChartRepositoryInsecure:      *commonCmdData.InsecureHelmDependencies,
		ChartRepositorySkipTLSVerify: *commonCmdData.SkipTlsVerifyHelmDependencies,
		ChartRepositorySkipUpdate:    *commonCmdData.SkipDependenciesRepoRefresh,
		DefaultSecretValuesDisable:   *commonCmdData.DisableDefaultSecretValues,
		DefaultValuesDisable:         *commonCmdData.DisableDefaultValues,
		ExtraAnnotations:             extraAnnotations,
		ExtraLabels:                  extraLabels,
		ExtraRuntimeAnnotations:      serviceAnnotations,
		ForceAdoption:                *commonCmdData.ForceAdoption,
		InstallGraphPath:             common.GetDeployGraphPath(&commonCmdData),
		InstallReportPath:            deployReportPath,
		KubeAPIServerName:            *commonCmdData.KubeApiServer,
		KubeBurstLimit:               *commonCmdData.KubeBurstLimit,
		KubeCAPath:                   *commonCmdData.KubeCaPath,
		KubeConfigBase64:             *commonCmdData.KubeConfigBase64,
		KubeConfigPaths:              append([]string{*commonCmdData.KubeConfig}, *commonCmdData.KubeConfigPathMergeList...),
		KubeContext:                  *commonCmdData.KubeContext,
		KubeQPSLimit:                 *commonCmdData.KubeQpsLimit,
		KubeSkipTLSVerify:            *commonCmdData.SkipTlsVerifyKube,
		KubeTLSServerName:            *commonCmdData.KubeTlsServer,
		KubeToken:                    *commonCmdData.KubeToken,
		LegacyChartType:              helmopts.ChartTypeBundle,
		LegacyExtraValues:            serviceValues,
		LogRegistryStreamOut:         os.Stdout,
		NetworkParallelism:           common.GetNetworkParallelism(&commonCmdData),
		NoPodLogs:                    *commonCmdData.NoPodLogs,
		NoInstallCRDs:                *commonCmdData.NoInstallCRDs,
		NoProgressTablePrint:         *commonCmdData.StatusProgressPeriodSeconds == -1,
		ProgressTablePrintInterval:   time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
		RegistryCredentialsPath:      registryCredentialsPath,
		ReleaseHistoryLimit:          *commonCmdData.ReleasesHistoryMax,
		ReleaseInfoAnnotations:       serviceAnnotations,
		ReleaseLabels:                releaseLabels,
		ReleaseStorageDriver:         os.Getenv("HELM_DRIVER"),
		RollbackGraphPath:            common.GetRollbackGraphPath(&commonCmdData),
		SecretKeyIgnore:              *commonCmdData.IgnoreSecretKey,
		SecretValuesPaths:            common.GetSecretValues(&commonCmdData),
		SecretWorkDir:                secretWorkDir,
		SubNotes:                     *commonCmdData.RenderSubchartNotes,
		TrackCreationTimeout:         time.Duration(cmdData.Timeout) * time.Second,
		TrackDeletionTimeout:         time.Duration(cmdData.Timeout) * time.Second,
		TrackReadinessTimeout:        time.Duration(cmdData.Timeout) * time.Second,
		ValuesFileSets:               common.GetSetFile(&commonCmdData),
		ValuesFilesPaths:             common.GetValues(&commonCmdData),
		ValuesSets:                   common.GetSet(&commonCmdData),
		ValuesStringSets:             common.GetSetString(&commonCmdData),
	}); err != nil {
		return fmt.Errorf("release install: %w", err)
	}

	return nil
}

func getAnnotationsAndLabels(bundleDir string) (map[string]string, map[string]string, map[string]string, error) {
	bundleExtraAnnotations, err := readBundleJsonMap(filepath.Join(bundleDir, "extra_annotations.json"))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read bundle extra_annotations.json: %w", err)
	}

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get user extra annotations: %w", err)
	}

	serviceAnnotations := map[string]string{}
	extraAnnotations := map[string]string{}
	for key, value := range lo.Assign(bundleExtraAnnotations, userExtraAnnotations) {
		if strings.HasPrefix(key, "project.werf.io/") ||
			strings.Contains(key, "ci.werf.io/") ||
			key == "werf.io/release-channel" {
			serviceAnnotations[key] = value
		} else {
			extraAnnotations[key] = value
		}
	}

	serviceAnnotations["werf.io/version"] = werf.Version
	if *commonCmdData.Environment != "" {
		serviceAnnotations["project.werf.io/env"] = *commonCmdData.Environment
	}

	bundleExtraLabels, err := readBundleJsonMap(filepath.Join(bundleDir, "extra_labels.json"))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read bundle extra_labels.json: %w", err)
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get user extra labels: %w", err)
	}

	extraLabels := lo.Assign(bundleExtraLabels, userExtraLabels)

	return serviceAnnotations, extraAnnotations, extraLabels, nil
}

func readBundleJsonMap(path string) (map[string]string, error) {
	var res map[string]string
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error accessing %q: %w", path, err)
	} else if data, err := ioutil.ReadFile(path); err != nil {
		return nil, fmt.Errorf("error reading %q: %w", path, err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("error unmarshalling json from %q: %w", path, err)
	} else {
		return res, nil
	}
}
