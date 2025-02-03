package render

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/chart/loader"
	"github.com/werf/3p-helm/pkg/chartutil"
	"github.com/werf/3p-helm/pkg/cli"
	"github.com/werf/3p-helm/pkg/registry"
	"github.com/werf/3p-helm/pkg/werf/secrets"
	"github.com/werf/3p-helm/pkg/werf/secrets/runtimedata"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender"
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
	global_warnings.PostponeMultiwerfNotUpToDateWarning()

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

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get user extra annotations: %w", err)
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get user extra labels: %w", err)
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

		if err := bundles.Pull(ctx, fmt.Sprintf("%s:%s", repoAddress, cmdData.Tag), bundlePath, bundlesRegistryClient); err != nil {
			return fmt.Errorf("pull bundle: %w", err)
		}
	}

	bundle, err := chart_extender.NewBundle(ctx, bundlePath, chart_extender.BundleOptions{
		SecretValueFiles:           common.GetSecretValues(&commonCmdData),
		BuildChartDependenciesOpts: chart.BuildChartDependenciesOptions{},
		ExtraAnnotations:           userExtraAnnotations,
		ExtraLabels:                userExtraLabels,
	})
	if err != nil {
		return fmt.Errorf("construct bundle: %w", err)
	}

	serviceAnnotations := map[string]string{}
	extraAnnotations := map[string]string{}
	for key, value := range bundle.GetExtraAnnotations() {
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

	extraLabels := bundle.GetExtraLabels()

	if err := action.Render(ctx, action.RenderOptions{
		ChartDirPath:                 bundle.Dir,
		ChartRepositoryInsecure:      *commonCmdData.InsecureHelmDependencies,
		ChartRepositorySkipTLSVerify: *commonCmdData.SkipTlsVerifyHelmDependencies,
		ChartRepositorySkipUpdate:    *commonCmdData.SkipDependenciesRepoRefresh,
		DefaultSecretValuesDisable:   *commonCmdData.DisableDefaultSecretValues,
		DefaultValuesDisable:         *commonCmdData.DisableDefaultValues,
		ExtraAnnotations:             extraAnnotations,
		ExtraLabels:                  extraLabels,
		ExtraRuntimeAnnotations:      serviceAnnotations,
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
		Local:                        !cmdData.Validate,
		LocalKubeVersion:             *commonCmdData.KubeVersion,
		LogDebug:                     *commonCmdData.LogDebug,
		LogRegistryStreamOut:         os.Stdout,
		NetworkParallelism:           *commonCmdData.NetworkParallelism,
		OutputFilePath:               cmdData.RenderOutput,
		OutputFileSave:               cmdData.RenderOutput != "",
		RegistryCredentialsPath:      docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig),
		ReleaseName:                  releaseName,
		ReleaseNamespace:             releaseNamespace,
		ReleaseStorageDriver:         action.ReleaseStorageDriver(os.Getenv("HELM_DRIVER")),
		SecretKeyIgnore:              *commonCmdData.IgnoreSecretKey,
		SecretValuesPaths:            common.GetSecretValues(&commonCmdData),
		ShowCRDs:                     cmdData.IncludeCRDs,
		ShowOnlyFiles:                append(util.PredefinedValuesByEnvNamePrefix("WERF_SHOW_ONLY"), cmdData.ShowOnly...),
		ValuesFileSets:               common.GetSetFile(&commonCmdData),
		ValuesFilesPaths:             common.GetValues(&commonCmdData),
		ValuesSets:                   common.GetSet(&commonCmdData),
		ValuesStringSets:             common.GetSetString(&commonCmdData),
		LegacyPreRenderHook: func(
			ctx context.Context,
			releaseNamespace string,
			helmRegistryClient *registry.Client,
			registryCredentialsPath string,
			chartRepositorySkipUpdate bool,
			secretValuesPaths []string,
			extraAnnotations map[string]string,
			extraLabels map[string]string,
			defaultValuesDisable bool,
			defaultSecretValuesDisable bool,
			helmSettings *cli.EnvSettings,
		) error {
			if vals, err := helpers.GetBundleServiceValues(ctx, helpers.ServiceValuesOptions{
				Env:                      *commonCmdData.Environment,
				Namespace:                releaseNamespace,
				SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
				DockerConfigPath:         filepath.Dir(registryCredentialsPath),
			}); err != nil {
				return fmt.Errorf("get service values: %w", err)
			} else {
				bundle.SetServiceValues(vals)
			}

			loader.GlobalLoadOptions = &chart.LoadOptions{
				ChartExtender: bundle,
				SubchartExtenderFactoryFunc: func() chart.ChartExtender {
					return chart_extender.NewWerfSubchart(ctx, chart_extender.WerfSubchartOptions{
						DisableDefaultSecretValues: *commonCmdData.DisableDefaultSecretValues,
					})
				},
				SecretsRuntimeDataFactoryFunc: func() runtimedata.RuntimeData {
					return secrets.NewSecretsRuntimeData()
				},
			}
			secrets.CoalesceTablesFunc = chartutil.CoalesceTables

			return nil
		},
	}); err != nil {
		return fmt.Errorf("render manifests: %w", err)
	}

	return nil
}
