package plan

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	nelmcommon "github.com/werf/nelm/pkg/common"
	"github.com/werf/nelm/pkg/featgate"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	DetailedExitCode       bool
	DiffContextLines       int
	ShowInsignificantDiffs bool
	ShowSensitiveDiffs     bool
	ShowVerboseCRDDiffs    bool
	ShowVerboseDiffs       bool
	Tag                    string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "plan",
		Short:                 "Plan bundle apply into Kubernetes",
		Long:                  common.GetLongCommandDescription(`Take latest bundle from the specified container registry using specified version tag or version mask and plan its deployment as a helm chart into Kubernetes cluster.`),
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

			return common.LogRunningTime(func() error { return runPlan(ctx) })
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

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	commonCmdData.SetupDebugTemplates(cmd)

	lo.Must0(common.SetupKubeConnectionFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupChartRepoConnectionFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupValuesFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupSecretValuesFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupResourceValidationFlags(&commonCmdData, cmd))

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)
	common.SetupChartProvenanceKeyring(&commonCmdData, cmd)
	common.SetupChartProvenanceStrategy(&commonCmdData, cmd)
	common.SetupDefaultDeletePropagation(&commonCmdData, cmd)
	common.SetupDeployGraphPath(&commonCmdData, cmd)
	common.SetupExtraRuntimeAnnotations(&commonCmdData, cmd)
	common.SetupExtraRuntimeLabels(&commonCmdData, cmd)
	common.SetupForceAdoption(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd, false)
	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupNoFinalTrackingFlag(&commonCmdData, cmd)
	common.SetupNoInstallCRDs(&commonCmdData, cmd)
	common.SetupNoRemoveManualChanges(&commonCmdData, cmd)
	common.SetupRelease(&commonCmdData, cmd, false)
	common.SetupReleaseInfoAnnotations(&commonCmdData, cmd)
	common.SetupReleaseLabel(&commonCmdData, cmd)
	common.SetupReleaseStorageDriver(&commonCmdData, cmd)
	common.SetupReleaseStorageSQLConnection(&commonCmdData, cmd)
	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupTemplatesAllowDNS(&commonCmdData, cmd)
	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Provide exact tag version or semver-based pattern, werf will install or upgrade to the latest version of the specified bundle ($WERF_TAG or latest by default)")
	var desc string
	if featgate.FeatGateMoreDetailedExitCodeForPlan.Enabled() || featgate.FeatGatePreviewV2.Enabled() {
		desc = "Return exit code 0 if no changes, 1 if error, 2 if resource changes planned, 3 if no resource changes planned, but release still should be installed (default $WERF_EXIT_CODE or false)"
	} else {
		desc = "Return exit code 0 if no changes, 1 if error, 2 if any changes planned (default $WERF_EXIT_CODE or false)"
	}
	cmd.Flags().BoolVarP(&cmdData.DetailedExitCode, "exit-code", "", util.GetBoolEnvironmentDefaultFalse("WERF_EXIT_CODE"), desc)
	cmd.Flags().BoolVarP(&cmdData.ShowInsignificantDiffs, "show-insignificant-diffs", "", util.GetBoolEnvironmentDefaultFalse("WERF_SHOW_INSIGNIFICANT_DIFFS"), "Show insignificant diff lines ($WERF_SHOW_INSIGNIFICANT_DIFFS by default)")
	cmd.Flags().BoolVarP(&cmdData.ShowSensitiveDiffs, "show-sensitive-diffs", "", util.GetBoolEnvironmentDefaultFalse("WERF_SHOW_SENSITIVE_DIFFS"), "Show sensitive diff lines ($WERF_SHOW_SENSITIVE_DIFFS by default)")
	cmd.Flags().BoolVarP(&cmdData.ShowVerboseCRDDiffs, "show-verbose-crd-diffs", "", util.GetBoolEnvironmentDefaultFalse("WERF_SHOW_VERBOSE_CRD_DIFFS"), "Show verbose CRD diff lines ($WERF_SHOW_VERBOSE_CRD_DIFFS by default)")
	// TODO(major): get rid?
	cmd.Flags().BoolVarP(&cmdData.ShowVerboseDiffs, "show-verbose-diffs", "", util.GetBoolEnvironmentDefaultTrue("WERF_SHOW_VERBOSE_DIFFS"), "Show verbose diff lines ($WERF_SHOW_VERBOSE_DIFFS by default)")
	var defaultDiffLines int
	if lines := lo.Must(util.GetIntEnvVar("WERF_DIFF_CONTEXT_LINES")); lines != nil {
		defaultDiffLines = int(*lines)
	} else {
		defaultDiffLines = nelmcommon.DefaultDiffContextLines
	}
	cmd.Flags().IntVarP(&cmdData.DiffContextLines, "diff-context-lines", "", defaultDiffLines, "Show N lines of context around diffs ($WERF_DIFF_CONTEXT_LINES by default)")

	return cmd
}

func runPlan(ctx context.Context) error {
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

	bundlePath := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
	defer os.RemoveAll(bundlePath)

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return fmt.Errorf("construct bundles registry client: %w", err)
	}

	registryCredentialsPath := docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig)

	serviceValues, err := helpers.GetBundleServiceValues(ctx, helpers.ServiceValuesOptions{
		Env:                      commonCmdData.Environment,
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
			DefaultSecretValuesDisable: commonCmdData.DefaultSecretValuesDisable,
			DefaultValuesDisable:       commonCmdData.DefaultValuesDisable,
			ExtraValues:                serviceValues,
			SecretKeyIgnore:            commonCmdData.SecretKeyIgnore,
			SecretValuesFiles:          commonCmdData.SecretValuesFiles,
			SecretWorkDir:              secretWorkDir,
		},
	}); err != nil {
		return fmt.Errorf("pull bundle: %w", err)
	}

	serviceAnnotations, extraAnnotations, extraLabels, err := getAnnotationsAndLabels(bundlePath)
	if err != nil {
		return fmt.Errorf("get annotations and labels: %w", err)
	}

	extraRuntimeAnnotations := lo.Assign(commonCmdData.ExtraRuntimeAnnotations, serviceAnnotations)
	releaseInfoAnnotations := lo.Assign(commonCmdData.ReleaseInfoAnnotations, serviceAnnotations)

	releaseLabels, err := common.GetReleaseLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get release labels: %w", err)
	}

	ctx = log.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultReleasePlanInstallLogLevel), log.SetupLoggingOptions{
		ColorMode: *commonCmdData.LogColorMode,
	})
	engine.Debug = commonCmdData.DebugTemplates

	if err := action.ReleasePlanInstall(ctx, releaseName, releaseNamespace, action.ReleasePlanInstallOptions{
		KubeConnectionOptions:      commonCmdData.KubeConnectionOptions,
		ChartRepoConnectionOptions: commonCmdData.ChartRepoConnectionOptions,
		ValuesOptions:              commonCmdData.ValuesOptions,
		SecretValuesOptions:        commonCmdData.SecretValuesOptions,
		ChartDirPath:               bundlePath,
		ChartProvenanceKeyring:     commonCmdData.ChartProvenanceKeyring,
		ChartProvenanceStrategy:    commonCmdData.ChartProvenanceStrategy,
		ChartRepoSkipUpdate:        commonCmdData.ChartRepoSkipUpdate,
		ErrorIfChangesPlanned:      cmdData.DetailedExitCode,
		InstallGraphPath:           commonCmdData.InstallGraphPath,
		LegacyChartType:            helmopts.ChartTypeBundle,
		LegacyExtraValues:          serviceValues,
		LegacyLogRegistryStreamOut: os.Stdout,
		NetworkParallelism:         commonCmdData.NetworkParallelism,
		NoFinalTracking:            commonCmdData.NoFinalTracking,
		RegistryCredentialsPath:    registryCredentialsPath,
		TemplatesAllowDNS:          commonCmdData.TemplatesAllowDNS,
		ReleaseInstallRuntimeOptions: nelmcommon.ReleaseInstallRuntimeOptions{
			ResourceValidationOptions:   commonCmdData.ResourceValidationOptions,
			DefaultDeletePropagation:    commonCmdData.DefaultDeletePropagation,
			ExtraAnnotations:            extraAnnotations,
			ExtraLabels:                 extraLabels,
			ExtraRuntimeAnnotations:     extraRuntimeAnnotations,
			ExtraRuntimeLabels:          commonCmdData.ExtraRuntimeLabels,
			ForceAdoption:               commonCmdData.ForceAdoption,
			NoInstallStandaloneCRDs:     commonCmdData.NoInstallStandaloneCRDs,
			NoRemoveManualChanges:       commonCmdData.NoRemoveManualChanges,
			ReleaseInfoAnnotations:      releaseInfoAnnotations,
			ReleaseLabels:               releaseLabels,
			ReleaseStorageDriver:        commonCmdData.ReleaseStorageDriver,
			ReleaseStorageSQLConnection: commonCmdData.ReleaseStorageSQLConnection,
		},
		ResourceDiffOptions: nelmcommon.ResourceDiffOptions{
			DiffContextLines:       cmdData.DiffContextLines,
			ShowVerboseCRDDiffs:    cmdData.ShowVerboseCRDDiffs,
			ShowVerboseDiffs:       cmdData.ShowVerboseDiffs,
			ShowSensitiveDiffs:     cmdData.ShowSensitiveDiffs,
			ShowInsignificantDiffs: cmdData.ShowInsignificantDiffs,
		},
	}); err != nil {
		return fmt.Errorf("release plan install: %w", err)
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
	if commonCmdData.Environment != "" {
		serviceAnnotations["project.werf.io/env"] = commonCmdData.Environment
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
