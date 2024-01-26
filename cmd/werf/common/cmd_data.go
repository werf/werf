package common

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/werf/pkg/util"
)

type CmdData struct {
	GitWorkTree              *string
	ProjectName              *string
	Dir                      *string
	ConfigPath               *string
	GiterminismConfigRelPath *string
	ConfigTemplatesDir       *string
	TmpDir                   *string
	HomeDir                  *string
	SSHKeys                  *[]string

	SkipDependenciesRepoRefresh *bool

	HelmChartDir                     *string
	Environment                      *string
	Release                          *string
	Namespace                        *string
	AddAnnotations                   *[]string
	AddLabels                        *[]string
	KubeContext                      *string
	KubeConfig                       *string
	KubeConfigBase64                 *string
	KubeConfigPathMergeList          *[]string
	StatusProgressPeriodSeconds      *int64
	HooksStatusProgressPeriodSeconds *int64
	ReleasesHistoryMax               *int

	SetDockerConfigJsonValue   *bool
	Set                        *[]string
	SetString                  *[]string
	Values                     *[]string
	SetFile                    *[]string
	SecretValues               *[]string
	IgnoreSecretKey            *bool
	DisableDefaultValues       *bool
	DisableDefaultSecretValues *bool
	HelmCompatibleChart        *bool
	RenameChart                *string

	WithoutImages *bool
	Repo          *RepoData
	FinalRepo     *RepoData

	SecondaryStagesStorage *[]string
	CacheStagesStorage     *[]string

	SkipBuild          *bool
	RequireBuiltImages *bool
	StubTags           *bool

	AddCustomTag *[]string
	UseCustomTag *string

	Synchronization    *string
	Parallel           *bool
	ParallelTasksLimit *int64
	NetworkParallelism *int

	DockerConfig                    *string
	InsecureRegistry                *bool
	SkipTlsVerifyRegistry           *bool
	InsecureHelmDependencies        *bool
	DryRun                          *bool
	KeepStagesBuiltWithinLastNHours *uint64
	WithoutKube                     *bool
	KubeVersion                     *string

	LooseGiterminism *bool
	Dev              *bool
	DevIgnore        *[]string
	DevBranch        *string

	IntrospectBeforeError *bool
	IntrospectAfterError  *bool
	StagesToIntrospect    *[]string

	Follow *bool

	LogDebug         *bool
	LogPretty        *bool
	LogTime          *bool
	LogTimeFormat    *string
	LogVerbose       *bool
	LogQuiet         *bool
	LogColorMode     *string
	LogProjectDir    *bool
	LogTerminalWidth *int64

	DeprecatedReportPath   *string
	DeprecatedReportFormat *string

	SaveBuildReport *bool
	BuildReportPath *string

	SaveDeployReport *bool
	UseDeployReport  *bool
	DeployReportPath *string

	DeployGraphPath   *string
	RollbackGraphPath *string

	VirtualMerge *bool

	ScanContextNamespaceOnly *bool

	// Host storage cleanup options
	DisableAutoHostCleanup                *bool
	DockerServerStoragePath               *string
	AllowedDockerStorageVolumeUsage       *uint
	AllowedDockerStorageVolumeUsageMargin *uint
	AllowedLocalCacheVolumeUsage          *uint
	AllowedLocalCacheVolumeUsageMargin    *uint

	Platform *[]string
}

func (cmdData *CmdData) SetupWithoutImages(cmd *cobra.Command) {
	cmdData.WithoutImages = new(bool)
	cmd.Flags().BoolVarP(cmdData.WithoutImages, "without-images", "", util.GetBoolEnvironmentDefaultFalse("WERF_WITHOUT_IMAGES"), "Disable building of images defined in the werf.yaml (if any) and usage of such images in the .helm/templates ($WERF_WITHOUT_IMAGES or false by default — e.g. enable all images defined in the werf.yaml by default)")
}

func (cmdData *CmdData) SetupDisableDefaultValues(cmd *cobra.Command) {
	cmdData.DisableDefaultValues = new(bool)
	cmd.Flags().BoolVarP(cmdData.DisableDefaultValues, "disable-default-values", "", util.GetBoolEnvironmentDefaultFalse("WERF_DISABLE_DEFAULT_VALUES"), `Do not use values from the default .helm/values.yaml file (default $WERF_DISABLE_DEFAULT_VALUES or false)`)
}

func (cmdData *CmdData) SetupPlatform(cmd *cobra.Command) {
	cmdData.Platform = new([]string)

	var defaultValue []string
	for _, envName := range []string{
		"WERF_PLATFORM",
		"DOCKER_DEFAULT_PLATFORM",
	} {
		if v := os.Getenv(envName); v != "" {
			defaultValue = []string{v}
			break
		}
	}

	cmd.Flags().StringArrayVarP(cmdData.Platform, "platform", "", defaultValue, "Enable platform emulation when building images with werf, format: OS/ARCH[/VARIANT] ($WERF_PLATFORM or $DOCKER_DEFAULT_PLATFORM by default)")
}

func (cmdData *CmdData) GetPlatform() []string {
	return *cmdData.Platform
}

func (cmdData *CmdData) SetupDisableDefaultSecretValues(cmd *cobra.Command) {
	cmdData.DisableDefaultSecretValues = new(bool)
	cmd.Flags().BoolVarP(cmdData.DisableDefaultSecretValues, "disable-default-secret-values", "", util.GetBoolEnvironmentDefaultFalse("WERF_DISABLE_DEFAULT_SECRET_VALUES"), `Do not use secret values from the default .helm/secret-values.yaml file (default $WERF_DISABLE_DEFAULT_SECRET_VALUES or false)`)
}

func (cmdData *CmdData) SetupSkipDependenciesRepoRefresh(cmd *cobra.Command) {
	cmdData.SkipDependenciesRepoRefresh = new(bool)
	cmd.Flags().BoolVarP(cmdData.SkipDependenciesRepoRefresh, "skip-dependencies-repo-refresh", "L", util.GetBoolEnvironmentDefaultFalse("WERF_SKIP_DEPENDENCIES_REPO_REFRESH"), `Do not refresh helm chart repositories locally cached index`)
}

func (cmdData *CmdData) SetupHelmCompatibleChart(cmd *cobra.Command, defaultEnabled bool) {
	cmdData.HelmCompatibleChart = new(bool)

	var defaultVal bool
	if defaultEnabled {
		defaultVal = util.GetBoolEnvironmentDefaultTrue("WERF_HELM_COMPATIBLE_CHART")
	} else {
		defaultVal = util.GetBoolEnvironmentDefaultFalse("WERF_HELM_COMPATIBLE_CHART")
	}

	cmd.Flags().BoolVarP(cmdData.HelmCompatibleChart, "helm-compatible-chart", "C", defaultVal, fmt.Sprintf(`Set chart name in the Chart.yaml of the published chart to the last path component of container registry repo (for REGISTRY/PATH/TO/REPO address chart name will be REPO, more info https://helm.sh/docs/topics/registries/#oci-feature-deprecation-and-behavior-changes-with-v370). In helm compatibility mode chart is fully conforming with the helm OCI registry requirements. Default %v or $WERF_HELM_COMPATIBLE_CHART.`, defaultEnabled))
}

func (cmdData *CmdData) SetupRenameChart(cmd *cobra.Command) {
	cmdData.RenameChart = new(string)
	cmd.Flags().StringVarP(cmdData.RenameChart, "rename-chart", "", os.Getenv("WERF_RENAME_CHART"), `Force setting of chart name in the Chart.yaml of the published chart to the specified value (can be set by the $WERF_RENAME_CHART, no rename by default, could not be used together with the '--helm-compatible-chart' option).`)
}
