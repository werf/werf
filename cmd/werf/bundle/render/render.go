package render

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/nelm/v2/pkg/action"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/engine"
	"github.com/werf/werf/v3/cmd/werf/common"
	"github.com/werf/werf/v3/pkg/deploy"
	"github.com/werf/werf/v3/pkg/deploy/bundles"
	"github.com/werf/werf/v3/pkg/docker"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/tmp_manager"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/pkg/werf/global_warnings"
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
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptionsDefaultQuiet(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	commonCmdData.SetupDebugTemplates(cmd)

	lo.Must0(common.SetupKubeConnectionFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupChartRepoConnectionFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupValuesFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupSecretValuesFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupNoValuesSchemaValidationFlag(&commonCmdData, cmd))
	common.SetupPatchesFlags(&commonCmdData, cmd, common.SetupPatchesFlagsOptions{})

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)
	common.SetupChartProvenanceKeyring(&commonCmdData, cmd)
	common.SetupChartProvenanceStrategy(&commonCmdData, cmd)
	common.SetupExtraAPIVersions(&commonCmdData, cmd)
	common.SetupKubeVersion(&commonCmdData, cmd)
	common.SetupBundleNamespace(&commonCmdData, cmd)
	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupBundleRelease(&commonCmdData, cmd)
	common.SetupReleaseStorageDriver(&commonCmdData, cmd)
	common.SetupReleaseStorageSQLConnection(&commonCmdData, cmd)
	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupTemplatesAllowDNS(&commonCmdData, cmd)
	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)
	common.SetupTSOptions(&commonCmdData, cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag,
		"Tag of the bundle to render ($WERF_TAG or latest by default)")
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

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	ctx = action.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultChartRenderLogLevel), action.SetupLoggingOptions{
		ColorMode:      *commonCmdData.LogColorMode,
		LogIsParseable: true,
	})

	var bundlePath string
	if isLocalBundle {
		if commonCmdData.ChartProvenanceStrategy == bundles.ProvenanceStrategyVerify {
			return fmt.Errorf("--provenance-strategy=%s can't verify the signature of the unpacked bundle in --bundle-dir: render the bundle from the container registry instead", bundles.ProvenanceStrategyVerify)
		}

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

		bundleDir := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
		defer os.RemoveAll(bundleDir)

		bundlePath, err = bundles.Download(ctx, repoAddress, cmdData.Tag, bundleDir, bundlesRegistryClient, bundles.DownloadOptions{
			ChartRepoConnectionOptions: common.GetBundlesRegistryConnectionOptions(&commonCmdData),
			ProvenanceKeyring:          commonCmdData.ChartProvenanceKeyring,
			ProvenanceStrategy:         commonCmdData.ChartProvenanceStrategy,
		})
		if err != nil {
			return err
		}
	}

	releaseName, releaseNamespace, err := bundles.ResolveRelease(ctx, bundlePath, commonCmdData.Release, commonCmdData.Namespace, commonCmdData.Environment)
	if err != nil {
		return err
	}

	releaseNamespace = cmp.Or(releaseNamespace, common.GetNamespace(&commonCmdData))
	releaseName = cmp.Or(releaseName, common.GetOptionalRelease(&commonCmdData))
	registryCredentialsPath := docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig)

	serviceValues, err := deploy.GetBundleServiceValues(ctx, deploy.ServiceValuesOptions{
		Env:                      commonCmdData.Environment,
		Namespace:                releaseNamespace,
		SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
		DockerConfigPath:         filepath.Dir(registryCredentialsPath),
	})
	if err != nil {
		return fmt.Errorf("get service values: %w", err)
	}

	serviceAnnotations, extraAnnotations, extraLabels, err := getAnnotationsAndLabels()
	if err != nil {
		return fmt.Errorf("get annotations and labels: %w", err)
	}

	engine.Debug = commonCmdData.DebugTemplates

	if _, err := action.ChartRender(ctx, action.ChartRenderOptions{
		KubeConnectionOptions:       commonCmdData.KubeConnectionOptions,
		ChartRepoConnectionOptions:  commonCmdData.ChartRepoConnectionOptions,
		ValuesOptions:               commonCmdData.ValuesOptions,
		SecretValuesOptions:         commonCmdData.SecretValuesOptions,
		Chart:                       bundlePath,
		ChartProvenanceKeyring:      commonCmdData.ChartProvenanceKeyring,
		ChartProvenanceStrategy:     commonCmdData.ChartProvenanceStrategy,
		ChartRepoSkipUpdate:         commonCmdData.ChartRepoSkipUpdate,
		DefaultPatchesDisable:       commonCmdData.DefaultPatchesDisable,
		ExtraAPIVersions:            commonCmdData.ExtraAPIVersions,
		ExtraAnnotations:            extraAnnotations,
		ExtraLabels:                 extraLabels,
		ExtraRuntimeAnnotations:     serviceAnnotations,
		LegacyChartType:             nelmcommon.LegacyChartTypeBundle,
		LegacyExtraValues:           serviceValues,
		LegacyLogRegistryStreamOut:  os.Stdout,
		LocalKubeVersion:            commonCmdData.KubeVersion,
		NetworkParallelism:          commonCmdData.NetworkParallelism,
		NoValuesSchemaValidation:    commonCmdData.NoValuesSchemaValidation,
		OutputFilePath:              cmdData.RenderOutput,
		PatchesFiles:                commonCmdData.PatchesFiles,
		RegistryCredentialsPath:     registryCredentialsPath,
		ReleaseName:                 releaseName,
		ReleaseNamespace:            releaseNamespace,
		ReleaseStorageDriver:        commonCmdData.ReleaseStorageDriver,
		ReleaseStorageSQLConnection: commonCmdData.ReleaseStorageSQLConnection,
		Remote:                      cmdData.Validate,
		ShowOnlyFiles:               append(util.PredefinedValuesByEnvNamePrefix("WERF_SHOW_ONLY"), cmdData.ShowOnly...),
		ShowStandaloneCRDs:          cmdData.IncludeCRDs,
		TemplatesAllowDNS:           commonCmdData.TemplatesAllowDNS,
		IgnoreBundleJS:              commonCmdData.IgnoreBundleJS,
		DenoBinaryPath:              commonCmdData.DenoBinaryPath,
	}); err != nil {
		return fmt.Errorf("chart render: %w", err)
	}

	return nil
}

// getAnnotationsAndLabels returns werf's service annotations and the user's extra annotations and
// labels to deploy the bundle with, in addition to the ones recorded in the bundle, which Nelm
// applies.
func getAnnotationsAndLabels() (map[string]string, map[string]string, map[string]string, error) {
	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get user extra annotations: %w", err)
	}

	serviceAnnotations := map[string]string{}
	extraAnnotations := map[string]string{}
	for key, value := range userExtraAnnotations {
		if strings.HasPrefix(key, "project.werf.io/") ||
			strings.Contains(key, "ci.werf.io/") ||
			key == "werf.io/release-channel" {
			serviceAnnotations[key] = value
		} else {
			extraAnnotations[key] = value
		}
	}

	serviceAnnotations["werf.io/version"] = werf.Version
	if commonCmdData.Environment != "" {
		serviceAnnotations["project.werf.io/env"] = commonCmdData.Environment
	}

	extraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get user extra labels: %w", err)
	}

	return serviceAnnotations, extraAnnotations, extraLabels, nil
}
