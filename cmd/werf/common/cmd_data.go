package common

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
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
	RuntimeJSONSets            *[]string
	IgnoreSecretKey            *bool
	DisableDefaultValues       *bool
	DisableDefaultSecretValues *bool
	HelmCompatibleChart        *bool
	RenameChart                *string

	FinalImagesOnly *bool
	WithoutImages   *bool
	Repo            *RepoData
	FinalRepo       *RepoData

	SecondaryStagesStorage *[]string
	CacheStagesStorage     *[]string

	RequireBuiltImages *bool
	StubTags           *bool

	AddCustomTag *[]string
	UseCustomTag *string

	Synchronization    *string
	Parallel           *bool
	ParallelTasksLimit *int64
	NetworkParallelism *int
	KubeQpsLimit       *int
	KubeBurstLimit     *int

	DockerConfig                    *string
	InsecureRegistry                *bool
	SkipTlsVerifyRegistry           *bool
	SkipTlsVerifyKube               *bool
	SkipTlsVerifyHelmDependencies   *bool
	KubeApiServer                   *string
	KubeCaPath                      *string
	KubeTlsServer                   *string
	KubeToken                       *string
	InsecureHelmDependencies        *bool
	DryRun                          *bool
	keepStagesBuiltWithinLastNHours *uint64
	WithoutKube                     *bool
	KubeVersion                     *string
	ContainerRegistryMirror         *[]string
	ForceAdoption                   *bool

	LooseGiterminism *bool
	Dev              *bool
	DevIgnore        *[]string
	DevBranch        *string

	IntrospectBeforeError *bool
	IntrospectAfterError  *bool
	StagesToIntrospect    *[]string

	Follow *bool

	// Logging options
	LogDebug         *bool
	LogPretty        *bool
	LogTime          *bool
	LogTimeFormat    *string
	LogVerbose       *bool
	LogQuiet         *bool
	LogColorMode     *string
	LogProjectDir    *bool
	LogTerminalWidth *int64
	NoPodLogs        *bool

	DebugTemplates *bool

	SaveBuildReport *bool
	BuildReportPath *string

	SaveDeployReport *bool
	UseDeployReport  *bool
	DeployReportPath *string

	SaveUninstallReport *bool
	UninstallReportPath *string

	DeployGraphPath    *string
	RollbackGraphPath  *string
	UninstallGraphPath *string

	RenderSubchartNotes   *bool
	NoInstallCRDs         *bool
	ReleaseLabels         *[]string
	NoRemoveManualChanges *bool
	NoFinalTracking       *bool

	VirtualMerge *bool

	ScanContextNamespaceOnly *bool

	// Host storage cleanup options
	DisableAutoHostCleanup                 *bool
	BackendStoragePath                     *string
	AllowedBackendStorageVolumeUsage       *uint
	AllowedBackendStorageVolumeUsageMargin *uint
	AllowedLocalCacheVolumeUsage           *uint
	AllowedLocalCacheVolumeUsageMargin     *uint

	Platform *[]string

	SkipImageSpecStage *bool
	IncludesLsFilter   *string

	CreateIncludesLockFile bool
	AllowIncludesUpdate    bool

	ReleaseStorageSQLConnection *string
}

func (cmdData *CmdData) SetupFinalImagesOnly(cmd *cobra.Command, defaultEnabled bool) {
	cmdData.FinalImagesOnly = new(bool)

	var defaultVal bool
	if defaultEnabled {
		defaultVal = util.GetBoolEnvironmentDefaultTrue("WERF_FINAL_IMAGES_ONLY")
	} else {
		defaultVal = util.GetBoolEnvironmentDefaultFalse("WERF_FINAL_IMAGES_ONLY")
	}

	cmd.Flags().BoolVarP(cmdData.FinalImagesOnly, "final-images-only", "", defaultVal, fmt.Sprintf("Process final images only ($WERF_FINAL_IMAGES_ONLY or %v by default)", defaultEnabled))
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

func (cmdData *CmdData) SetupSkipImageSpecStage(cmd *cobra.Command) {
	cmdData.SkipImageSpecStage = new(bool)
	cmd.Flags().BoolVarP(cmdData.SkipImageSpecStage, "skip-image-spec-stage", "", util.GetBoolEnvironmentDefaultFalse("WERF_SKIP_IMAGE_SPEC_STAGE"), `Force skipping "imageSpec" build stage (default $WERF_SKIP_IMAGE_SPEC_STAGE or false)`)
}

func (cmdData *CmdData) SetupCreateIncludesLockFile() {
	cmdData.CreateIncludesLockFile = true
}

func (cmdData *CmdData) SetupAllowIncludesUpdate(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.AllowIncludesUpdate, "allow-includes-update", "", util.GetBoolEnvironmentDefaultFalse("WERF_ALLOW_INCLUDES_UPDATE"), `Allow use includes latest versions (default $WERF_ALLOW_INCLUDES_UPDATE or false)`)
}

func (cmdData *CmdData) SetupIncludesLsFilter(cmd *cobra.Command) {
	cmdData.IncludesLsFilter = new(string)
	cmd.Flags().StringVar(cmdData.IncludesLsFilter, "filter", os.Getenv("WERF_INCLUDES_LIST_FILTER"), "Filter by source, e.g. --filter=source=local,remoteRepo (default $WERF_INCLUDES_LIST_FILTER or all sources).")
}

func (cmdData *CmdData) SetupDebugTemplates(cmd *cobra.Command) {
	if cmdData.DebugTemplates == nil {
		cmdData.DebugTemplates = new(bool)
	}
	cmd.Flags().BoolVarP(
		cmdData.DebugTemplates,
		"debug-templates",
		"",
		util.GetBoolEnvironmentDefaultFalse("WERF_DEBUG_TEMPLATES"),
		`Enable debug mode for Go templates (default $WERF_DEBUG_TEMPLATES or false)`,
	)
}
