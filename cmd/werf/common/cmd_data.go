package common

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/nelm/pkg/common"
	"github.com/werf/werf/v2/pkg/util/option"
)

type CmdData struct {
	common.KubeConnectionOptions
	common.ChartRepoConnectionOptions
	common.ValuesOptions
	common.SecretValuesOptions
	common.TrackingOptions

	GitWorkTree              *string
	ProjectName              *string
	Dir                      *string
	ConfigPath               *string
	GiterminismConfigRelPath *string
	ConfigTemplatesDir       *string
	ConfigRenderPath         *string
	TmpDir                   *string
	HomeDir                  *string
	SSHKeys                  *[]string

	Environment string

	SetDockerConfigJsonValue *bool

	FinalImagesOnly *bool
	WithoutImages   *bool
	Repo            *RepoData
	FinalRepo       *RepoData

	SecondaryStagesStorage *[]string
	CacheStagesStorage     *[]string

	CheckBuiltImages       *bool
	LegacyCheckBuiltImages *bool // TODO(v3): remove
	RequireBuiltImages     *bool
	StubTags               *bool

	AddCustomTag *[]string
	UseCustomTag *string

	Synchronization    *string
	Parallel           *bool
	ParallelTasksLimit *int64

	DockerConfig                    *string
	InsecureRegistry                *bool
	SkipTlsVerifyRegistry           *bool
	DryRun                          *bool
	keepStagesBuiltWithinLastNHours *uint64
	WithoutKube                     *bool
	ContainerRegistryMirror         *[]string

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

	SaveBuildReport *bool
	BuildReportPath *string

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

	AnnotateLayersWithDmvVerityRootHash *bool

	SignManifest      *bool
	SignKey           *string
	SignCert          *string
	SignIntermediates *string

	VerifyManifest *bool
	VerifyRoots    *[]string

	BSignELFFiles               *bool
	SignELFFiles                *bool
	ELFPGPPrivateKeyBase64      *string
	ELFPGPPrivateKeyFingerprint *string
	ELFPGPPrivateKeyPassphrase  *string

	VerifyELFFiles      *bool
	VerifyBSignELFFiles *bool
	ImageRef            *[]string

	SkipImageSpecStage *bool
	IncludesLsFilter   *string

	CreateIncludesLockFile bool
	AllowIncludesUpdate    bool

	ChartProvenanceKeyring           string
	ChartProvenanceStrategy          string
	ChartRepoSkipUpdate              bool
	DebugTemplates                   bool
	DefaultDeletePropagation         string
	DeployReportPath                 string
	ExtraAPIVersions                 []string
	ExtraAnnotations                 []string
	ExtraLabels                      []string
	ExtraRuntimeAnnotations          map[string]string
	ExtraRuntimeLabels               map[string]string
	ForceAdoption                    bool
	HelmCompatibleChart              bool
	InstallGraphPath                 string
	KubeVersion                      string
	LegacyKubeConfigPath             string
	LegacyKubeConfigPathsMergeList   []string
	LegacyProgressTablePrintInterval int
	LegacyTrackTimeout               int
	Namespace                        string
	NetworkParallelism               int
	NoInstallStandaloneCRDs          bool
	NoRemoveManualChanges            bool
	NoShowNotes                      bool
	Release                          string
	ReleaseHistoryLimit              int
	ReleaseInfoAnnotations           map[string]string
	ReleaseLabels                    []string
	ReleaseStorageDriver             string
	ReleaseStorageSQLConnection      string
	RenameChart                      string
	RollbackGraphPath                string
	RollbackReportPath               string
	SaveDeployReport                 bool
	SaveRollbackReport               bool
	SaveUninstallReport              bool
	ShowSubchartNotes                bool
	TemplatesAllowDNS                bool
	UninstallGraphPath               string
	UninstallReportPath              string
	UseDeployReport                  bool
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
	return option.PtrValueOrDefault(cmdData.Platform, []string{})
}

func (cmdData *CmdData) SetupSkipDependenciesRepoRefresh(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.ChartRepoSkipUpdate, "skip-dependencies-repo-refresh", "L", util.GetBoolEnvironmentDefaultFalse("WERF_SKIP_DEPENDENCIES_REPO_REFRESH"), `Do not refresh helm chart repositories locally cached index`)
}

func (cmdData *CmdData) SetupHelmCompatibleChart(cmd *cobra.Command, defaultEnabled bool) {
	var defaultVal bool
	if defaultEnabled {
		defaultVal = util.GetBoolEnvironmentDefaultTrue("WERF_HELM_COMPATIBLE_CHART")
	} else {
		defaultVal = util.GetBoolEnvironmentDefaultFalse("WERF_HELM_COMPATIBLE_CHART")
	}

	cmd.Flags().BoolVarP(&cmdData.HelmCompatibleChart, "helm-compatible-chart", "C", defaultVal, fmt.Sprintf(`Set chart name in the Chart.yaml of the published chart to the last path component of container registry repo (for REGISTRY/PATH/TO/REPO address chart name will be REPO, more info https://helm.sh/docs/topics/registries/#oci-feature-deprecation-and-behavior-changes-with-v370). In helm compatibility mode chart is fully conforming with the helm OCI registry requirements. Default %v or $WERF_HELM_COMPATIBLE_CHART.`, defaultEnabled))
}

func (cmdData *CmdData) SetupRenameChart(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.RenameChart, "rename-chart", "", os.Getenv("WERF_RENAME_CHART"), `Force setting of chart name in the Chart.yaml of the published chart to the specified value (can be set by the $WERF_RENAME_CHART, no rename by default, could not be used together with the '--helm-compatible-chart' option).`)
}

// TODO: remove this legacy logic in v3.
func (cmdData *CmdData) SetupSkipImageSpecStage(cmd *cobra.Command) {
	cmdData.SkipImageSpecStage = new(bool)
	cmd.Flags().BoolVarP(cmdData.SkipImageSpecStage, "skip-image-spec-stage", "", util.GetBoolEnvironmentDefaultFalse("WERF_SKIP_IMAGE_SPEC_STAGE"), `Force skipping "imageSpec" build stage (default $WERF_SKIP_IMAGE_SPEC_STAGE or false)`)
	cmd.Flags().MarkHidden("skip-image-spec-stage")
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
	cmd.Flags().BoolVarP(
		&cmdData.DebugTemplates,
		"debug-templates",
		"",
		util.GetBoolEnvironmentDefaultFalse("WERF_DEBUG_TEMPLATES"),
		`Enable debug mode for Go templates (default $WERF_DEBUG_TEMPLATES or false)`,
	)
}

func (cmdData *CmdData) ProcessFlags() error {
	if err := cmdData.validateFlags(); err != nil {
		return fmt.Errorf("validate flags: %w", err)
	}

	if err := cmdData.mapLegacyFlags(); err != nil {
		return fmt.Errorf("map legacy flags: %w", err)
	}

	if err := cmdData.processFlags(); err != nil {
		return err
	}

	return nil
}

func (cmdData *CmdData) validateFlags() error {
	return nil
}

func (cmdData *CmdData) mapLegacyFlags() error {
	cmdData.KubeConnectionOptions.KubeConfigPaths = append([]string{cmdData.LegacyKubeConfigPath}, cmdData.LegacyKubeConfigPathsMergeList...)
	cmdData.TrackingOptions.NoProgressTablePrint = cmdData.LegacyProgressTablePrintInterval == -1
	cmdData.TrackingOptions.ProgressTablePrintInterval = time.Duration(cmdData.LegacyProgressTablePrintInterval) * time.Second
	cmdData.TrackingOptions.TrackCreationTimeout = time.Duration(cmdData.LegacyTrackTimeout) * time.Second
	cmdData.TrackDeletionTimeout = time.Duration(cmdData.LegacyTrackTimeout) * time.Second
	cmdData.TrackReadinessTimeout = time.Duration(cmdData.LegacyTrackTimeout) * time.Second

	return nil
}

func (cmdData *CmdData) processFlags() error {
	if cmdData.DeployReportPath == "" {
		cmdData.DeployReportPath = DefaultDeployReportPathJSON
	}

	switch ext := filepath.Ext(cmdData.DeployReportPath); ext {
	case ".json":
	case "":
		cmdData.DeployReportPath += ".json"
	default:
		return fmt.Errorf("invalid --deploy-report-path %q: extension must be either .json or unspecified", cmdData.DeployReportPath)
	}

	if cmdData.RollbackReportPath == "" {
		cmdData.RollbackReportPath = DefaultRollbackReportPathJSON
	}

	switch ext := filepath.Ext(cmdData.RollbackReportPath); ext {
	case ".json":
	case "":
		cmdData.RollbackReportPath += ".json"
	default:
		return fmt.Errorf("invalid --rollback-report-path %q: extension must be either .json or unspecified", cmdData.RollbackReportPath)
	}

	if cmdData.UninstallReportPath == "" {
		cmdData.UninstallReportPath = DefaultUninstallReportPathJSON
	}

	switch ext := filepath.Ext(cmdData.UninstallReportPath); ext {
	case ".json":
	case "":
		cmdData.UninstallReportPath += ".json"
	default:
		return fmt.Errorf("invalid --uninstall-report-path %q: extension must be either .json or unspecified", cmdData.UninstallReportPath)
	}

	if cmdData.InstallGraphPath != "" {
		switch ext := filepath.Ext(cmdData.InstallGraphPath); ext {
		case ".dot":
		case "":
			cmdData.InstallGraphPath += ".dot"
		default:
			return fmt.Errorf("invalid --deploy-graph-path %q: extension must be either .dot or unspecified", cmdData.InstallGraphPath)
		}
	}

	if cmdData.RollbackGraphPath != "" {
		switch ext := filepath.Ext(cmdData.RollbackGraphPath); ext {
		case ".dot":
		case "":
			cmdData.RollbackGraphPath += ".dot"
		default:
			return fmt.Errorf("invalid --rollback-graph-path %q: extension must be either .dot or unspecified", cmdData.RollbackGraphPath)
		}
	}

	if cmdData.UninstallGraphPath != "" {
		switch ext := filepath.Ext(cmdData.UninstallGraphPath); ext {
		case ".dot":
		case "":
			cmdData.UninstallGraphPath += ".dot"
		default:
			return fmt.Errorf("invalid --uninstall-graph-path %q: extension must be either .dot or unspecified", cmdData.UninstallGraphPath)
		}
	}

	cmdData.KubeImpersonateGroups = append(util.PredefinedValuesByEnvNamePrefix("WERF_KUBE_IMPERSONATE_GROUP_"), cmdData.KubeImpersonateGroups...)
	cmdData.ValuesSet = append(util.PredefinedValuesByEnvNamePrefix("WERF_SET_", "WERF_SET_STRING_", "WERF_SET_FILE_", "WERF_SET_DOCKER_CONFIG_JSON_VALUE"), cmdData.ValuesSet...)
	cmdData.ValuesSetString = append(util.PredefinedValuesByEnvNamePrefix("WERF_SET_STRING_"), cmdData.ValuesSetString...)
	cmdData.ValuesSetFile = append(util.PredefinedValuesByEnvNamePrefix("WERF_SET_FILE_"), cmdData.ValuesSetFile...)
	cmdData.RuntimeSetJSON = append(util.PredefinedValuesByEnvNamePrefix("WERF_SET_RUNTIME_JSON_"), cmdData.RuntimeSetJSON...)
	cmdData.ValuesSetJSON = append(util.PredefinedValuesByEnvNamePrefix("WERF_SET_JSON_"), cmdData.ValuesSetJSON...)
	cmdData.ValuesSetLiteral = append(util.PredefinedValuesByEnvNamePrefix("WERF_SET_LITERAL_"), cmdData.ValuesSetLiteral...)
	cmdData.ValuesFiles = append(util.PredefinedValuesByEnvNamePrefix("WERF_VALUES_"), cmdData.ValuesFiles...)
	cmdData.SecretValuesFiles = append(util.PredefinedValuesByEnvNamePrefix("WERF_SECRET_VALUES_"), cmdData.SecretValuesFiles...)
	cmdData.ExtraAPIVersions = append(util.PredefinedValuesByEnvNamePrefix("WERF_EXTRA_APIVERSIONS_"), cmdData.ExtraAPIVersions...)

	return nil
}
