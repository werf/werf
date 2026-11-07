package plan

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
	"github.com/werf/nelm/v2/pkg/action"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/engine"
	"github.com/werf/werf/v3/cmd/werf/common"
	"github.com/werf/werf/v3/pkg/deploy"
	"github.com/werf/werf/v3/pkg/deploy/bundles"
	"github.com/werf/werf/v3/pkg/docker"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/pkg/werf/global_warnings"
)

var cmdData struct {
	DetailedExitCode       bool
	DiffContextLines       int
	PlanArtifactPath       string
	ShowPlanArtifactPath   string
	ShowInsignificantDiffs bool
	ShowSensitiveDiffs     bool
	ShowVerboseCRDDiffs    bool
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
	common.SetupPatchesFlags(&commonCmdData, cmd, common.SetupPatchesFlagsOptions{})

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)
	common.SetupChartProvenanceKeyring(&commonCmdData, cmd)
	common.SetupChartProvenanceStrategy(&commonCmdData, cmd)
	common.SetupDefaultDeletePropagation(&commonCmdData, cmd)
	common.SetupDeployGraphPath(&commonCmdData, cmd)
	common.SetupExtraRuntimeAnnotations(&commonCmdData, cmd)
	common.SetupExtraRuntimeLabels(&commonCmdData, cmd)
	common.SetupForceAdoption(&commonCmdData, cmd)
	common.SetupBundleNamespace(&commonCmdData, cmd)
	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupNoFinalTrackingFlag(&commonCmdData, cmd)
	common.SetupNoInstallCRDs(&commonCmdData, cmd)
	common.SetupNoRemoveManualChanges(&commonCmdData, cmd)
	common.SetupBundleRelease(&commonCmdData, cmd)
	common.SetupReleaseInfoAnnotations(&commonCmdData, cmd)
	common.SetupReleaseLabel(&commonCmdData, cmd)
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
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Tag of the bundle to install or upgrade to ($WERF_TAG or latest by default)")
	cmd.Flags().BoolVarP(&cmdData.DetailedExitCode, "exit-code", "", util.GetBoolEnvironmentDefaultFalse("WERF_EXIT_CODE"), "Return exit code 0 if no changes, 1 if error, 2 if resource changes planned, 3 if no resource changes planned, but release still should be installed (default $WERF_EXIT_CODE or false)")
	cmd.Flags().BoolVarP(&cmdData.ShowInsignificantDiffs, "show-insignificant-diffs", "", util.GetBoolEnvironmentDefaultFalse("WERF_SHOW_INSIGNIFICANT_DIFFS"), "Show insignificant diff lines ($WERF_SHOW_INSIGNIFICANT_DIFFS by default)")
	cmd.Flags().BoolVarP(&cmdData.ShowSensitiveDiffs, "show-sensitive-diffs", "", util.GetBoolEnvironmentDefaultFalse("WERF_SHOW_SENSITIVE_DIFFS"), "Show sensitive diff lines ($WERF_SHOW_SENSITIVE_DIFFS by default)")
	cmd.Flags().BoolVarP(&cmdData.ShowVerboseCRDDiffs, "show-verbose-crd-diffs", "", util.GetBoolEnvironmentDefaultFalse("WERF_SHOW_VERBOSE_CRD_DIFFS"), "Show verbose CRD diff lines ($WERF_SHOW_VERBOSE_CRD_DIFFS by default)")
	cmd.Flags().StringVarP(&cmdData.PlanArtifactPath, "save-plan", "", os.Getenv("WERF_SAVE_PLAN_PATH"), "Save the gzip-compressed JSON install plan to the specified file")
	cmd.Flags().StringVarP(&cmdData.ShowPlanArtifactPath, "show-plan", "", os.Getenv("WERF_SHOW_PLAN_PATH"), "Show plan artifact planned changes")
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
	ctx = action.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultReleasePlanInstallLogLevel), action.SetupLoggingOptions{
		ColorMode: *commonCmdData.LogColorMode,
	})

	if cmdData.ShowPlanArtifactPath != "" {
		if err := action.ReleasePlanShow(ctx, action.ReleasePlanShowOptions{
			ResourceDiffOptions: nelmcommon.ResourceDiffOptions{
				DiffContextLines:       cmdData.DiffContextLines,
				ShowVerboseCRDDiffs:    cmdData.ShowVerboseCRDDiffs,
				ShowSensitiveDiffs:     cmdData.ShowSensitiveDiffs,
				ShowInsignificantDiffs: cmdData.ShowInsignificantDiffs,
			},
			PlanArtifactPath: cmdData.ShowPlanArtifactPath,
			SecretKey:        commonCmdData.SecretKey,
			SecretWorkDir:    commonCmdData.SecretWorkDir,
		}); err != nil {
			return fmt.Errorf("release plan show: %w", err)
		}

		return nil
	}

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

	bundleDir := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
	defer os.RemoveAll(bundleDir)

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return fmt.Errorf("construct bundles registry client: %w", err)
	}

	bundlePath, err := bundles.Download(ctx, repoAddress, cmdData.Tag, bundleDir, bundlesRegistryClient, bundles.DownloadOptions{
		ChartRepoConnectionOptions: common.GetBundlesRegistryConnectionOptions(&commonCmdData),
		ProvenanceKeyring:          commonCmdData.ChartProvenanceKeyring,
		ProvenanceStrategy:         commonCmdData.ChartProvenanceStrategy,
	})
	if err != nil {
		return err
	}

	releaseName, releaseNamespace, err := bundles.ResolveRelease(ctx, bundlePath, commonCmdData.Release, commonCmdData.Namespace, commonCmdData.Environment)
	if err != nil {
		return err
	}

	releaseNamespace = cmp.Or(releaseNamespace, common.GetNamespace(&commonCmdData))

	if releaseName == "" {
		releaseName, err = common.GetRequiredRelease(&commonCmdData)
		if err != nil {
			return fmt.Errorf("get release name: %w", err)
		}
	}

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

	releaseLabels, err := common.GetReleaseLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get release labels: %w", err)
	}

	engine.Debug = commonCmdData.DebugTemplates

	ctx = common.SetupDenoContext(ctx)

	if _, err := action.ReleasePlanInstall(ctx, releaseName, releaseNamespace, action.ReleasePlanInstallOptions{
		KubeConnectionOptions:      commonCmdData.KubeConnectionOptions,
		ChartRepoConnectionOptions: commonCmdData.ChartRepoConnectionOptions,
		ValuesOptions:              commonCmdData.ValuesOptions,
		SecretValuesOptions:        commonCmdData.SecretValuesOptions,
		Chart:                      bundlePath,
		ChartProvenanceKeyring:     commonCmdData.ChartProvenanceKeyring,
		ChartProvenanceStrategy:    commonCmdData.ChartProvenanceStrategy,
		ChartRepoSkipUpdate:        commonCmdData.ChartRepoSkipUpdate,
		ErrorIfChangesPlanned:      cmdData.DetailedExitCode,
		InstallGraphPath:           commonCmdData.InstallGraphPath,
		LegacyChartType:            nelmcommon.LegacyChartTypeBundle,
		LegacyExtraValues:          serviceValues,
		LegacyLogRegistryStreamOut: os.Stdout,
		NetworkParallelism:         commonCmdData.NetworkParallelism,
		NoFinalTracking:            commonCmdData.NoFinalTracking,
		PlanArtifactPath:           cmdData.PlanArtifactPath,
		RegistryCredentialsPath:    registryCredentialsPath,
		TemplatesAllowDNS:          commonCmdData.TemplatesAllowDNS,
		ReleaseInstallRuntimeOptions: nelmcommon.ReleaseInstallRuntimeOptions{
			ResourceValidationOptions:   commonCmdData.ResourceValidationOptions,
			DefaultDeletePropagation:    commonCmdData.DefaultDeletePropagation,
			ExtraAnnotations:            extraAnnotations,
			ExtraLabels:                 extraLabels,
			ExtraRuntimeAnnotations:     lo.Assign(commonCmdData.ExtraRuntimeAnnotations, serviceAnnotations),
			ExtraRuntimeLabels:          commonCmdData.ExtraRuntimeLabels,
			ForceAdoption:               commonCmdData.ForceAdoption,
			NoInstallStandaloneCRDs:     commonCmdData.NoInstallStandaloneCRDs,
			NoRemoveManualChanges:       commonCmdData.NoRemoveManualChanges,
			PatchesFiles:                commonCmdData.PatchesFiles,
			DefaultPatchesDisable:       commonCmdData.DefaultPatchesDisable,
			ReleaseInfoAnnotations:      lo.Assign(commonCmdData.ReleaseInfoAnnotations, serviceAnnotations),
			ReleaseLabels:               releaseLabels,
			ReleaseStorageDriver:        commonCmdData.ReleaseStorageDriver,
			ReleaseStorageSQLConnection: commonCmdData.ReleaseStorageSQLConnection,
		},
		ResourceDiffOptions: nelmcommon.ResourceDiffOptions{
			DiffContextLines:       cmdData.DiffContextLines,
			ShowVerboseCRDDiffs:    cmdData.ShowVerboseCRDDiffs,
			ShowSensitiveDiffs:     cmdData.ShowSensitiveDiffs,
			ShowInsignificantDiffs: cmdData.ShowInsignificantDiffs,
		},
		IgnoreBundleJS: commonCmdData.IgnoreBundleJS,
		DenoBinaryPath: commonCmdData.DenoBinaryPath,
	}); err != nil {
		return fmt.Errorf("release plan install: %w", err)
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
