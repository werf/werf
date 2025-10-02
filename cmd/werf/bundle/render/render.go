package render

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	ioutil "io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/3p-helm/pkg/engine"
	"github.com/werf/3p-helm/pkg/werf/helmopts"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	Tag          string
	BundleDir    string
	RenderOutput string
	Validate     bool
	IncludeCRDs  bool
	ShowOnly     []string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "render",
		Short:                 "Render Kubernetes manifests from bundle",
		Long:                  common.GetLongCommandDescription(`Take locally extracted bundle or download bundle from the specified container registry using specified version tag or version mask and render it as Kubernetes manifests.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runRender(ctx) })
		},
	})

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, false)
	common.SetupSkipTLSVerifyKube(&commonCmdData, cmd)
	common.SetupKubeApiServer(&commonCmdData, cmd)
	common.SetupSQLConnectionString(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyHelmDependencies(&commonCmdData, cmd)
	common.SetupKubeCaPath(&commonCmdData, cmd)
	common.SetupKubeTlsServer(&commonCmdData, cmd)
	common.SetupKubeToken(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptionsDefaultQuiet(&commonCmdData, cmd)
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

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupRelease(&commonCmdData, cmd, false)
	common.SetupNamespace(&commonCmdData, cmd, false)

	common.SetupKubeVersion(&commonCmdData, cmd)

	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupKubeQpsLimit(&commonCmdData, cmd)
	common.SetupKubeBurstLimit(&commonCmdData, cmd)
	common.SetupForceAdoption(&commonCmdData, cmd)

	commonCmdData.SetupDebugTemplates(cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}

	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag,
		"Provide exact tag version or semver-based pattern, werf will render the latest version of the specified bundle ($WERF_TAG or latest by default)")
	cmd.Flags().BoolVarP(&cmdData.IncludeCRDs, "include-crds", "", util.GetBoolEnvironmentDefaultTrue("WERF_INCLUDE_CRDS"),
		"Include CRDs in the templated output (default $WERF_INCLUDE_CRDS)")
	cmd.Flags().StringVarP(&cmdData.RenderOutput, "output", "", os.Getenv("WERF_RENDER_OUTPUT"),
		"Write render output to the specified file instead of stdout ($WERF_RENDER_OUTPUT by default)")
	cmd.Flags().StringVarP(&cmdData.BundleDir, "bundle-dir", "b", os.Getenv(("WERF_BUNDLE_DIR")),
		"Get extracted bundle from directory instead of registry (default $WERF_BUNDLE_DIR)")

	cmd.Flags().BoolVarP(&cmdData.Validate, "validate", "", util.GetBoolEnvironmentDefaultFalse("WERF_VALIDATE"), "Validate your manifests against the Kubernetes cluster you are currently pointing at (default $WERF_VALIDATE)")
	cmd.Flags().StringArrayVarP(&cmdData.ShowOnly, "show-only", "s", []string{}, "only show manifests rendered from the given templates")

	return cmd
}

func runRender(ctx context.Context) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning(ctx)

	var isLocalBundle bool
	switch {
	case cmdData.BundleDir != "":
		if *commonCmdData.Repo.Address != "" {
			return fmt.Errorf("only one of --bundle-dir or --repo should be specified, but both provided")
		}

		isLocalBundle = true
	case *commonCmdData.Repo.Address == storage.LocalStorageAddress:
		return fmt.Errorf("--repo %s is not allowed, specify remote storage address", storage.LocalStorageAddress)
	case *commonCmdData.Repo.Address != "":
		isLocalBundle = false
	default:
		return fmt.Errorf("either --bundle-dir or --repo required")
	}

	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                &commonCmdData,
		InitDockerRegistry: !isLocalBundle,
		InitWerf:           true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	releaseNamespace := common.GetNamespace(&commonCmdData)
	releaseName := common.GetOptionalRelease(&commonCmdData)
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

	var bundlePath string
	if isLocalBundle {
		bundlePath = cmdData.BundleDir
	} else {
		bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
		if err != nil {
			return fmt.Errorf("construct bundles registry client: %w", err)
		}

		repoAddress, err := commonCmdData.Repo.GetAddress()
		if err != nil {
			return fmt.Errorf("get repo address: %w", err)
		}

		bundlePath = filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
		defer os.RemoveAll(bundlePath)

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
	}

	serviceAnnotations, extraAnnotations, extraLabels, err := getAnnotationsAndLabels(bundlePath)
	if err != nil {
		return fmt.Errorf("get annotations and labels: %w", err)
	}

	// TODO(v3): get rid of forcing color mode via ci-env and use color mode detection logic from
	// Nelm instead. Until then, color will be always off here.
	ctx = action.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultChartRenderLogLevel), action.SetupLoggingOptions{
		ColorMode:      action.LogColorModeOff,
		LogIsParseable: true,
	})
	engine.Debug = *commonCmdData.DebugTemplates

	if _, err := action.ChartRender(ctx, action.ChartRenderOptions{
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
		LocalKubeVersion:             *commonCmdData.KubeVersion,
		LogRegistryStreamOut:         os.Stdout,
		NetworkParallelism:           *commonCmdData.NetworkParallelism,
		OutputFilePath:               cmdData.RenderOutput,
		RegistryCredentialsPath:      registryCredentialsPath,
		ReleaseName:                  releaseName,
		ReleaseNamespace:             releaseNamespace,
		ReleaseStorageDriver:         os.Getenv("HELM_DRIVER"),
		SQLConnectionString:  *commonCmdData.SQLConnectionString,
		Remote:                       cmdData.Validate,
		SecretKeyIgnore:              *commonCmdData.IgnoreSecretKey,
		SecretValuesPaths:            common.GetSecretValues(&commonCmdData),
		SecretWorkDir:                secretWorkDir,
		ShowCRDs:                     cmdData.IncludeCRDs,
		ShowOnlyFiles:                append(util.PredefinedValuesByEnvNamePrefix("WERF_SHOW_ONLY"), cmdData.ShowOnly...),
		ValuesFileSets:               common.GetSetFile(&commonCmdData),
		ValuesFilesPaths:             common.GetValues(&commonCmdData),
		ValuesSets:                   common.GetSet(&commonCmdData),
		ValuesStringSets:             common.GetSetString(&commonCmdData),
	}); err != nil {
		return fmt.Errorf("chart render: %w", err)
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
