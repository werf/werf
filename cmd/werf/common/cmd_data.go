package common

import (
	"github.com/spf13/cobra"

	"github.com/werf/werf/pkg/util"
)

type CmdData struct {
	GitWorkTree        *string
	ProjectName        *string
	Dir                *string
	ConfigPath         *string
	ConfigTemplatesDir *string
	TmpDir             *string
	HomeDir            *string
	SSHKeys            *[]string

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

	WithoutImages *bool
	Repo          *RepoData
	FinalRepo     *RepoData

	SecondaryStagesStorage *[]string
	CacheStagesStorage     *[]string

	SkipBuild *bool
	StubTags  *bool

	AddCustomTag *[]string
	UseCustomTag *string

	Synchronization    *string
	Parallel           *bool
	ParallelTasksLimit *int64

	DockerConfig                    *string
	InsecureRegistry                *bool
	SkipTlsVerifyRegistry           *bool
	InsecureHelmDependencies        *bool
	DryRun                          *bool
	KeepStagesBuiltWithinLastNHours *uint64
	WithoutKube                     *bool

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
	LogVerbose       *bool
	LogQuiet         *bool
	LogColorMode     *string
	LogProjectDir    *bool
	LogTerminalWidth *int64

	ReportPath   *string
	ReportFormat *string

	VirtualMerge *bool

	ScanContextNamespaceOnly *bool

	// Host storage cleanup options
	DisableAutoHostCleanup                *bool
	DockerServerStoragePath               *string
	AllowedDockerStorageVolumeUsage       *uint
	AllowedDockerStorageVolumeUsageMargin *uint
	AllowedLocalCacheVolumeUsage          *uint
	AllowedLocalCacheVolumeUsageMargin    *uint

	Platform *string
}

func (cmdData *CmdData) SetupWithoutImages(cmd *cobra.Command) {
	cmdData.WithoutImages = new(bool)
	cmd.Flags().BoolVarP(cmdData.WithoutImages, "without-images", "", util.GetBoolEnvironmentDefaultFalse("WERF_WITHOUT_IMAGES"), "Disable building of images defined in the werf.yaml (if any) and usage of such images in the .helm/templates ($WERF_WITHOUT_IMAGES or false by default — e.g. enable all images defined in the werf.yaml by default)")
}

func (cmdData *CmdData) SetupDisableDefaultValues(cmd *cobra.Command) {
	cmdData.DisableDefaultValues = new(bool)
	cmd.Flags().BoolVarP(cmdData.DisableDefaultValues, "disable-default-values", "", util.GetBoolEnvironmentDefaultFalse("WERF_DISABLE_DEFAULT_VALUES"), `Do not use values from the default .helm/values.yaml file (default $WERF_DISABLE_DEFAULT_VALUES or false)`)
}

func (cmdData *CmdData) SetupDisableDefaultSecretValues(cmd *cobra.Command) {
	cmdData.DisableDefaultSecretValues = new(bool)
	cmd.Flags().BoolVarP(cmdData.DisableDefaultSecretValues, "disable-default-secret-values", "", util.GetBoolEnvironmentDefaultFalse("WERF_DISABLE_DEFAULT_SECRET_VALUES"), `Do not use secret values from the default .helm/secret-values.yaml file (default $WERF_DISABLE_DEFAULT_SECRET_VALUES or false)`)
}

func (cmdData *CmdData) SetupSkipDependenciesRepoRefresh(cmd *cobra.Command) {
	cmdData.SkipDependenciesRepoRefresh = new(bool)
	cmd.Flags().BoolVarP(cmdData.SkipDependenciesRepoRefresh, "skip-dependencies-repo-refresh", "L", util.GetBoolEnvironmentDefaultFalse("WERF_SKIP_DEPENDENCIES_REPO_REFRESH"), `Do not refresh helm chart repositories locally cached index`)
}
