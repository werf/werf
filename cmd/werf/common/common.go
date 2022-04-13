package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
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

	SetDockerConfigJsonValue *bool
	Set                      *[]string
	SetString                *[]string
	Values                   *[]string
	SetFile                  *[]string
	SecretValues             *[]string
	IgnoreSecretKey          *bool

	CommonRepoData *RepoData
	StagesStorage  *string

	CommonFinalRepoData *RepoData
	FinalStagesStorage  *string

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

const (
	CleaningCommandsForceOptionDescription = "First remove containers that use werf docker images which are going to be deleted"
	StubRepoAddress                        = "stub/repository"
	StubTag                                = "TAG"
	DefaultBuildParallelTasksLimit         = 5
	DefaultCleanupParallelTasksLimit       = 10
)

func GetLongCommandDescription(text string) string {
	return logboek.FitText(text, types.FitTextOptions{MaxWidth: 100})
}

func SetupSetDockerConfigJsonValue(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SetDockerConfigJsonValue = new(bool)
	cmd.Flags().BoolVarP(cmdData.SetDockerConfigJsonValue, "set-docker-config-json-value", "", GetBoolEnvironmentDefaultFalse("WERF_SET_DOCKER_CONFIG_JSON_VALUE"), "Shortcut to set current docker config into the .Values.dockerconfigjson")
}

func SetupGitWorkTree(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.GitWorkTree = new(string)
	cmd.Flags().StringVarP(cmdData.GitWorkTree, "git-work-tree", "", os.Getenv("WERF_GIT_WORK_TREE"), "Use specified git work tree dir (default $WERF_WORK_TREE or lookup for directory that contains .git in the current or parent directories)")
}

func SetupProjectName(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ProjectName = new(string)
	cmd.Flags().StringVarP(cmdData.ProjectName, "project-name", "N", os.Getenv("WERF_PROJECT_NAME"), "Set a specific project name (default $WERF_PROJECT_NAME)")
}

func SetupDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Dir = new(string)
	cmd.Flags().StringVarP(cmdData.Dir, "dir", "", os.Getenv("WERF_DIR"), "Use specified project directory where project’s werf.yaml and other configuration files should reside (default $WERF_DIR or current working directory)")
}

func SetupConfigPath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ConfigPath = new(string)
	cmd.Flags().StringVarP(cmdData.ConfigPath, "config", "", os.Getenv("WERF_CONFIG"), `Use custom configuration file (default $WERF_CONFIG or werf.yaml in working directory)`)
}

func SetupConfigTemplatesDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ConfigTemplatesDir = new(string)
	cmd.Flags().StringVarP(cmdData.ConfigTemplatesDir, "config-templates-dir", "", os.Getenv("WERF_CONFIG_TEMPLATES_DIR"), `Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf in working directory)`)
}

type SetupTmpDirOptions struct {
	Persistent bool
}

func SetupTmpDir(cmdData *CmdData, cmd *cobra.Command, opts SetupTmpDirOptions) {
	cmdData.TmpDir = new(string)
	getFlags(cmd, opts.Persistent).StringVarP(cmdData.TmpDir, "tmp-dir", "", "", "Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)")
}

func SetupGiterminismOptions(cmdData *CmdData, cmd *cobra.Command) {
	setupLooseGiterminism(cmdData, cmd)
	setupDev(cmdData, cmd)
	setupDevIgnore(cmdData, cmd)
	setupDevBranch(cmdData, cmd)
}

func setupLooseGiterminism(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LooseGiterminism = new(bool)
	cmd.Flags().BoolVarP(cmdData.LooseGiterminism, "loose-giterminism", "", GetBoolEnvironmentDefaultFalse("WERF_LOOSE_GITERMINISM"), "Loose werf giterminism mode restrictions (NOTE: not all restrictions can be removed, more info https://werf.io/documentation/advanced/giterminism.html, default $WERF_LOOSE_GITERMINISM)")
}

func setupDev(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Dev = new(bool)
	cmd.Flags().BoolVarP(cmdData.Dev, "dev", "", GetBoolEnvironmentDefaultFalse("WERF_DEV"), `Enable development mode (default $WERF_DEV).
The mode allows working with project files without doing redundant commits during debugging and development`)
}

func setupDevIgnore(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DevIgnore = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.DevIgnore, "dev-ignore", "", []string{}, `Add rules to ignore tracked and untracked changes in development mode (can specify multiple).
Also, can be specified with $WERF_DEV_IGNORE_* (e.g. $WERF_DEV_IGNORE_TESTS=*_test.go, $WERF_DEV_IGNORE_DOCS=path/to/docs)`)
}

func setupDevBranch(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DevBranch = new(string)

	defaultValue := "_werf-dev"
	envValue := os.Getenv("WERF_DEV_BRANCH")
	if envValue != "" {
		defaultValue = envValue
	}

	cmd.Flags().StringVarP(cmdData.DevBranch, "dev-branch", "", defaultValue, fmt.Sprintf("Set dev git branch name (default $WERF_DEV_BRANCH or %q)", defaultValue))
}

type SetupHomeDirOptions struct {
	Persistent bool
}

func SetupHomeDir(cmdData *CmdData, cmd *cobra.Command, opts SetupHomeDirOptions) {
	cmdData.HomeDir = new(string)
	getFlags(cmd, opts.Persistent).StringVarP(cmdData.HomeDir, "home-dir", "", "", "Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)")
}

func SetupSSHKey(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SSHKeys = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SSHKeys, "ssh-key", "", []string{}, `Use only specific ssh key(s).
Can be specified with $WERF_SSH_KEY_* (e.g. $WERF_SSH_KEY_REPO=~/.ssh/repo_rsa, $WERF_SSH_KEY_NODEJS=~/.ssh/nodejs_rsa).
Defaults to $WERF_SSH_KEY_*, system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see https://werf.io/documentation/reference/toolbox/ssh.html`)
}

func SetupReportPath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ReportPath = new(string)
	cmd.Flags().StringVarP(cmdData.ReportPath, "report-path", "", os.Getenv("WERF_REPORT_PATH"), "Report save path ($WERF_REPORT_PATH by default)")
}

func SetupReportFormat(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ReportFormat = new(string)

	defaultValue := os.Getenv("WERF_REPORT_FORMAT")
	if defaultValue == "" {
		defaultValue = string(build.ReportJSON)
	}

	cmd.Flags().StringVarP(cmdData.ReportFormat, "report-format", "", defaultValue, fmt.Sprintf(`Report format: %[1]s or %[2]s (%[1]s or $WERF_REPORT_FORMAT by default)
%[1]s:
	{
	  "Images": {
		"<WERF_IMAGE_NAME>": {
			"WerfImageName": "<WERF_IMAGE_NAME>",
			"DockerRepo": "<REPO>",
			"DockerTag": "<TAG>"
			"DockerImageName": "<REPO>:<TAG>",
			"DockerImageID": "<SHA256>",
			"DockerImageDigest": "<SHA256>",
		},
		...
	  }
	}
%[2]s:
	WERF_<FORMATTED_WERF_IMAGE_NAME>_DOCKER_IMAGE_NAME=<REPO>:<TAG>
	...
<FORMATTED_WERF_IMAGE_NAME> is werf image name from werf.yaml modified according to the following rules:
- all characters are uppercase (app -> APP);
- charset /- is replaced with _ (DEV/APP-FRONTEND -> DEV_APP_FRONTEND)`, string(build.ReportJSON), string(build.ReportEnvFile)))
}

func GetReportFormat(cmdData *CmdData) (build.ReportFormat, error) {
	switch format := build.ReportFormat(*cmdData.ReportFormat); format {
	case build.ReportJSON, build.ReportEnvFile:
		return format, nil
	default:
		return "", fmt.Errorf("bad --report-format given %q, expected: \"%s\"", format, strings.Join([]string{string(build.ReportJSON), string(build.ReportEnvFile)}, "\", \""))
	}
}

func SetupWithoutKube(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.WithoutKube = new(bool)
	cmd.Flags().BoolVarP(cmdData.WithoutKube, "without-kube", "", GetBoolEnvironmentDefaultFalse("WERF_WITHOUT_KUBE"), "Do not skip deployed Kubernetes images (default $WERF_WITHOUT_KUBE)")
}

func SetupKeepStagesBuiltWithinLastNHours(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KeepStagesBuiltWithinLastNHours = new(uint64)

	envValue, err := GetUint64EnvVar("WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS")
	if err != nil {
		TerminateWithError(err.Error(), 1)
	}

	var defaultValue uint64
	if envValue != nil {
		defaultValue = *envValue
	} else {
		defaultValue = 2
	}

	cmd.Flags().Uint64VarP(cmdData.KeepStagesBuiltWithinLastNHours, "keep-stages-built-within-last-n-hours", "", defaultValue, "Keep stages that were built within last hours (default $WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS or 2)")
}

func PredefinedValuesByEnvNamePrefix(envNamePrefix string, envNamePrefixesToExcept ...string) []string {
	var result []string

	env := os.Environ()
	sort.Strings(env)

environLoop:
	for _, keyValue := range env {
		parts := strings.SplitN(keyValue, "=", 2)
		if strings.HasPrefix(parts[0], envNamePrefix) {
			for _, exceptEnvNamePrefix := range envNamePrefixesToExcept {
				if strings.HasPrefix(parts[0], exceptEnvNamePrefix) {
					continue environLoop
				}
			}

			result = append(result, parts[1])
		}
	}

	return result
}

func SetupEnvironment(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Environment = new(string)
	cmd.Flags().StringVarP(cmdData.Environment, "env", "", os.Getenv("WERF_ENV"), "Use specified environment (default $WERF_ENV)")
}

func SetupRelease(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Release = new(string)
	cmd.Flags().StringVarP(cmdData.Release, "release", "", os.Getenv("WERF_RELEASE"), "Use specified Helm release name (default [[ project ]]-[[ env ]] template or deploy.helmRelease custom template from werf.yaml or $WERF_RELEASE)")
}

func SetupNamespace(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Namespace = new(string)
	cmd.Flags().StringVarP(cmdData.Namespace, "namespace", "", os.Getenv("WERF_NAMESPACE"), "Use specified Kubernetes namespace (default [[ project ]]-[[ env ]] template or deploy.namespace custom template from werf.yaml or $WERF_NAMESPACE)")
}

func SetupAddAnnotations(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.AddAnnotations = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.AddAnnotations, "add-annotation", "", []string{}, `Add annotation to deploying resources (can specify multiple).
Format: annoName=annoValue.
Also, can be specified with $WERF_ADD_ANNOTATION_* (e.g. $WERF_ADD_ANNOTATION_1=annoName1=annoValue1, $WERF_ADD_ANNOTATION_2=annoName2=annoValue2)`)
}

func SetupAddLabels(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.AddLabels = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.AddLabels, "add-label", "", []string{}, `Add label to deploying resources (can specify multiple).
Format: labelName=labelValue.
Also, can be specified with $WERF_ADD_LABEL_* (e.g. $WERF_ADD_LABEL_1=labelName1=labelValue1, $WERF_ADD_LABEL_2=labelName2=labelValue2)`)
}

func SetupKubeContext(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeContext = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.KubeContext, "kube-context", "", os.Getenv("WERF_KUBE_CONTEXT"), "Kubernetes config context (default $WERF_KUBE_CONTEXT)")
}

func SetupKubeConfig(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeConfig = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.KubeConfig, "kube-config", "", "", "Kubernetes config file path (default $WERF_KUBE_CONFIG, or $WERF_KUBECONFIG, or $KUBECONFIG)")

	cmdData.KubeConfigPathMergeList = new([]string)
	kubeConfigPathMergeListStr := GetFirstExistingKubeConfigEnvVar()
	for _, path := range filepath.SplitList(kubeConfigPathMergeListStr) {
		*cmdData.KubeConfigPathMergeList = append(*cmdData.KubeConfigPathMergeList, path)
	}
}

func GetFirstExistingKubeConfigEnvVar() string {
	return GetFirstExistingEnvVarAsString("WERF_KUBE_CONFIG", "WERF_KUBECONFIG", "KUBECONFIG")
}

func SetupKubeConfigBase64(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeConfigBase64 = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.KubeConfigBase64, "kube-config-base64", "", GetFirstExistingKubeConfigBase64EnvVar(), "Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)")
}

func GetFirstExistingKubeConfigBase64EnvVar() string {
	return GetFirstExistingEnvVarAsString("WERF_KUBE_CONFIG_BASE64", "WERF_KUBECONFIG_BASE64", "KUBECONFIG_BASE64")
}

func GetFirstExistingEnvVarAsString(envNames ...string) string {
	for _, envName := range envNames {
		if v := os.Getenv(envName); v != "" {
			return v
		}
	}

	return ""
}

func SetupCommonRepoData(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.CommonRepoData = &RepoData{IsCommon: true}

	SetupImplementationForRepoData(cmdData.CommonRepoData, cmd, "repo-implementation", []string{"WERF_REPO_IMPLEMENTATION"}) // legacy
	SetupContainerRegistryForRepoData(cmdData.CommonRepoData, cmd, "repo-container-registry", []string{"WERF_REPO_CONTAINER_REGISTRY"})
	SetupDockerHubUsernameForRepoData(cmdData.CommonRepoData, cmd, "repo-docker-hub-username", []string{"WERF_REPO_DOCKER_HUB_USERNAME"})
	SetupDockerHubPasswordForRepoData(cmdData.CommonRepoData, cmd, "repo-docker-hub-password", []string{"WERF_REPO_DOCKER_HUB_PASSWORD"})
	SetupDockerHubTokenForRepoData(cmdData.CommonRepoData, cmd, "repo-docker-hub-token", []string{"WERF_REPO_DOCKER_HUB_TOKEN"})
	SetupGithubTokenForRepoData(cmdData.CommonRepoData, cmd, "repo-github-token", []string{"WERF_REPO_GITHUB_TOKEN"})
	SetupHarborUsernameForRepoData(cmdData.CommonRepoData, cmd, "repo-harbor-username", []string{"WERF_REPO_HARBOR_USERNAME"})
	SetupHarborPasswordForRepoData(cmdData.CommonRepoData, cmd, "repo-harbor-password", []string{"WERF_REPO_HARBOR_PASSWORD"})
	SetupQuayTokenForRepoData(cmdData.CommonRepoData, cmd, "repo-quay-token", []string{"WERF_REPO_QUAY_TOKEN"})
}

func SetupCommonFinalRepoData(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.CommonFinalRepoData = &RepoData{}

	cmdData.CommonFinalRepoData.Implementation = new(string) // legacy
	SetupContainerRegistryForRepoData(cmdData.CommonFinalRepoData, cmd, "final-repo-container-registry", []string{"WERF_FINAL_REPO_CONTAINER_REGISTRY"})
	SetupDockerHubUsernameForRepoData(cmdData.CommonFinalRepoData, cmd, "final-repo-docker-hub-username", []string{"WERF_FINAL_REPO_DOCKER_HUB_USERNAME"})
	SetupDockerHubPasswordForRepoData(cmdData.CommonFinalRepoData, cmd, "final-repo-docker-hub-password", []string{"WERF_FINAL_REPO_DOCKER_HUB_PASSWORD"})
	SetupDockerHubTokenForRepoData(cmdData.CommonFinalRepoData, cmd, "final-repo-docker-hub-token", []string{"WERF_FINAL_REPO_DOCKER_HUB_TOKEN"})
	SetupGithubTokenForRepoData(cmdData.CommonFinalRepoData, cmd, "final-repo-github-token", []string{"WERF_FINAL_REPO_GITHUB_TOKEN"})
	SetupHarborUsernameForRepoData(cmdData.CommonFinalRepoData, cmd, "final-repo-harbor-username", []string{"WERF_FINAL_REPO_HARBOR_USERNAME"})
	SetupHarborPasswordForRepoData(cmdData.CommonFinalRepoData, cmd, "final-repo-harbor-password", []string{"WERF_FINAL_REPO_HARBOR_PASSWORD"})
	SetupQuayTokenForRepoData(cmdData.CommonFinalRepoData, cmd, "final-repo-quay-token", []string{"WERF_FINAL_REPO_QUAY_TOKEN"})
}

func SetupSecondaryStagesStorageOptions(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SecondaryStagesStorage = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SecondaryStagesStorage, "secondary-repo", "", []string{}, `Specify one or multiple secondary read-only repos with images that will be used as a cache.
Also, can be specified with $WERF_SECONDARY_REPO_* (e.g. $WERF_SECONDARY_REPO_1=..., $WERF_SECONDARY_REPO_2=...)`)
}

func SetupAddCustomTag(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.AddCustomTag = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.AddCustomTag, "add-custom-tag", "", []string{}, `Set tag alias for the content-based tag.
The alias may contain the following shortcuts:
- %image%, %image_slug% or %image_safe_slug% to use the image name (necessary if there is more than one image in the werf config);
- %image_content_based_tag% to use a content-based tag.
For cleaning custom tags and associated content-based tag are treated as one.
Also can be defined with $WERF_ADD_CUSTOM_TAG_* (e.g. $WERF_ADD_CUSTOM_TAG_1="%image%-tag1", $WERF_ADD_CUSTOM_TAG_2="%image%-tag2")`)
}

func SetupUseCustomTag(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.UseCustomTag = new(string)
	cmd.Flags().StringVarP(cmdData.UseCustomTag, "use-custom-tag", "", os.Getenv("WERF_USE_CUSTOM_TAG"), `Use a tag alias in helm templates instead of an image content-based tag (NOT RECOMMENDED).
The alias may contain the following shortcuts:
- %image%, %image_slug% or %image_safe_slug% to use the image name (necessary if there is more than one image in the werf config);
- %image_content_based_tag% to use a content-based tag.
For cleaning custom tags and associated content-based tag are treated as one.
Also, can be defined with $WERF_USE_CUSTOM_TAG (e.g. $WERF_USE_CUSTOM_TAG="%image%-tag")`)
}

func SetupCacheStagesStorageOptions(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.CacheStagesStorage = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.CacheStagesStorage, "cache-repo", "", []string{}, `Specify one or multiple cache repos with images that will be used as a cache. Cache will be populated when pushing newly built images into the primary repo and when pulling existing images from the primary repo. Cache repo will be used to pull images and to get manifests before making requests to the primary repo.
Also, can be specified with $WERF_CACHE_REPO_* (e.g. $WERF_CACHE_REPO_1=..., $WERF_CACHE_REPO_2=...)`)
}

func SetupStagesStorageOptions(cmdData *CmdData, cmd *cobra.Command) {
	SetupInsecureRegistry(cmdData, cmd)
	SetupSkipTlsVerifyRegistry(cmdData, cmd)
	SetupCommonRepoData(cmdData, cmd)
	setupStagesStorage(cmdData, cmd)
}

func SetupFinalStagesStorageOptions(cmdData *CmdData, cmd *cobra.Command) {
	SetupCommonFinalRepoData(cmdData, cmd)
	setupFinalStagesStorage(cmdData, cmd)
}

func setupStagesStorage(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.StagesStorage, "repo", "", os.Getenv("WERF_REPO"), fmt.Sprintf("Docker Repo to store stages (default $WERF_REPO)"))
}

func setupFinalStagesStorage(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.FinalStagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.FinalStagesStorage, "final-repo", "", os.Getenv("WERF_FINAL_REPO"), fmt.Sprintf("Docker Repo to store only those stages which are going to be used by the Kubernetes cluster, in other word final images (default $WERF_FINAL_REPO)"))
}

func SetupStatusProgressPeriod(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StatusProgressPeriodSeconds = new(int64)
	SetupStatusProgressPeriodP(cmdData.StatusProgressPeriodSeconds, cmd)
}

func SetupStatusProgressPeriodP(destination *int64, cmd *cobra.Command) {
	cmd.PersistentFlags().Int64VarP(
		destination,
		"status-progress-period",
		"",
		*statusProgressPeriodDefaultValue(),
		"Status progress period in seconds. Set -1 to stop showing status progress. Defaults to $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds",
	)
}

func SetupReleasesHistoryMax(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ReleasesHistoryMax = new(int)

	defaultValueP, err := GetIntEnvVar("WERF_RELEASES_HISTORY_MAX")
	if err != nil {
		TerminateWithError(fmt.Sprintf("bad WERF_RELEASES_HISTORY_MAX value: %s", err), 1)
	}

	var defaultValue int
	if defaultValueP != nil {
		defaultValue = int(*defaultValueP)
	}

	cmd.Flags().IntVarP(
		cmdData.ReleasesHistoryMax,
		"releases-history-max",
		"",
		defaultValue,
		"Max releases to keep in release storage. Can be set by environment variable $WERF_RELEASES_HISTORY_MAX. By default werf keeps all releases.",
	)
}

func statusProgressPeriodDefaultValue() *int64 {
	defaultValue := int64(5)

	v, err := GetIntEnvVar("WERF_STATUS_PROGRESS_PERIOD_SECONDS")
	if err != nil {
		TerminateWithError(err.Error(), 1)
	}

	if v == nil {
		v, err = GetIntEnvVar("WERF_STATUS_PROGRESS_PERIOD")
		if err != nil {
			TerminateWithError(err.Error(), 1)
		}

		if v == nil {
			return &defaultValue
		} else {
			return v
		}
	} else {
		return v
	}
}

func SetupHooksStatusProgressPeriod(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HooksStatusProgressPeriodSeconds = new(int64)
	SetupHooksStatusProgressPeriodP(cmdData.HooksStatusProgressPeriodSeconds, cmd)
}

func SetupHooksStatusProgressPeriodP(destination *int64, cmd *cobra.Command) {
	cmd.PersistentFlags().Int64VarP(
		destination,
		"hooks-status-progress-period",
		"",
		*hooksStatusProgressPeriodDefaultValue(),
		"Hooks status progress period in seconds. Set 0 to stop showing hooks status progress. Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value",
	)
}

func hooksStatusProgressPeriodDefaultValue() *int64 {
	defaultValue := statusProgressPeriodDefaultValue()

	v, err := GetIntEnvVar("WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS")
	if err != nil {
		TerminateWithError(err.Error(), 1)
	}

	if v == nil {
		v, err = GetIntEnvVar("WERF_HOOKS_STATUS_PROGRESS_PERIOD")
		if err != nil {
			TerminateWithError(err.Error(), 1)
		}

		if v == nil {
			return defaultValue
		} else {
			return v
		}
	} else {
		return v
	}
}

func SetupInsecureHelmDependencies(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.InsecureHelmDependencies = new(bool)
	cmd.Flags().BoolVarP(cmdData.InsecureHelmDependencies, "insecure-helm-dependencies", "", GetBoolEnvironmentDefaultFalse("WERF_INSECURE_HELM_DEPENDENCIES"), "Allow insecure oci registries to be used in the .helm/Chart.yaml dependencies configuration (default $WERF_INSECURE_HELM_DEPENDENCIES)")
}

func SetupInsecureRegistry(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.InsecureRegistry != nil {
		return
	}

	cmdData.InsecureRegistry = new(bool)
	cmd.Flags().BoolVarP(cmdData.InsecureRegistry, "insecure-registry", "", GetBoolEnvironmentDefaultFalse("WERF_INSECURE_REGISTRY"), "Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)")
}

func SetupSkipTlsVerifyRegistry(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.SkipTlsVerifyRegistry != nil {
		return
	}

	cmdData.SkipTlsVerifyRegistry = new(bool)
	cmd.Flags().BoolVarP(cmdData.SkipTlsVerifyRegistry, "skip-tls-verify-registry", "", GetBoolEnvironmentDefaultFalse("WERF_SKIP_TLS_VERIFY_REGISTRY"), "Skip TLS certificate validation when accessing a registry (default $WERF_SKIP_TLS_VERIFY_REGISTRY)")
}

func SetupDryRun(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DryRun = new(bool)
	cmd.Flags().BoolVarP(cmdData.DryRun, "dry-run", "", GetBoolEnvironmentDefaultFalse("WERF_DRY_RUN"), "Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)")
}

func SetupDockerConfig(cmdData *CmdData, cmd *cobra.Command, extraDesc string) {
	defaultValue := os.Getenv("WERF_DOCKER_CONFIG")
	if defaultValue == "" {
		defaultValue = os.Getenv("DOCKER_CONFIG")
	}

	cmdData.DockerConfig = new(string)

	desc := "Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or ~/.docker (in the order of priority)"

	if extraDesc != "" {
		desc += "\n"
		desc += extraDesc
	}

	cmd.Flags().StringVarP(cmdData.DockerConfig, "docker-config", "", defaultValue, desc)
}

func SetupLogOptions(cmdData *CmdData, cmd *cobra.Command) {
	setupLogDebug(cmdData, cmd)
	setupLogVerbose(cmdData, cmd)
	setupLogQuiet(cmdData, cmd, false)
	setupLogColor(cmdData, cmd)
	setupLogPretty(cmdData, cmd)
	setupTerminalWidth(cmdData, cmd)
}

func SetupLogOptionsDefaultQuiet(cmdData *CmdData, cmd *cobra.Command) {
	setupLogDebug(cmdData, cmd)
	setupLogVerbose(cmdData, cmd)
	setupLogQuiet(cmdData, cmd, true)
	setupLogColor(cmdData, cmd)
	setupLogPretty(cmdData, cmd)
	setupTerminalWidth(cmdData, cmd)
}

func setupLogDebug(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogDebug = new(bool)

	defaultValue := false
	for _, envName := range []string{
		"WERF_LOG_DEBUG",
		"WERF_DEBUG",
	} {
		if os.Getenv(envName) != "" {
			defaultValue = GetBoolEnvironmentDefaultFalse(envName)
			break
		}
	}

	for alias, env := range map[string]string{
		"log-debug": "WERF_LOG_DEBUG",
		"debug":     "WERF_DEBUG",
	} {
		cmd.PersistentFlags().BoolVarP(
			cmdData.LogDebug,
			alias,
			"",
			defaultValue,
			fmt.Sprintf("Enable debug (default $%s).", env),
		)
	}

	if err := cmd.PersistentFlags().MarkHidden("debug"); err != nil {
		panic(err)
	}
}

func setupLogColor(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogColorMode = new(string)

	logColorEnvironmentValue := os.Getenv("WERF_LOG_COLOR_MODE")

	defaultValue := "auto"
	if logColorEnvironmentValue != "" {
		defaultValue = logColorEnvironmentValue
	}

	cmd.PersistentFlags().StringVarP(cmdData.LogColorMode, "log-color-mode", "", defaultValue, `Set log color mode.
Supported on, off and auto (based on the stdout’s file descriptor referring to a terminal) modes.
Default $WERF_LOG_COLOR_MODE or auto mode.`)
}

func setupLogQuiet(cmdData *CmdData, cmd *cobra.Command, isDefaultQuiet bool) {
	cmdData.LogQuiet = new(bool)

	defaultValue := isDefaultQuiet

	for _, envName := range []string{
		"WERF_LOG_QUIET",
		"WERF_QUIET",
	} {
		if os.Getenv(envName) != "" {
			if defaultValue {
				defaultValue = GetBoolEnvironmentDefaultTrue(envName)
			} else {
				defaultValue = GetBoolEnvironmentDefaultFalse(envName)
			}

			break
		}
	}

	for alias, env := range map[string]string{
		"log-quiet": "WERF_LOG_QUIET",
		"quiet":     "WERF_QUIET",
	} {
		cmd.PersistentFlags().BoolVarP(
			cmdData.LogQuiet,
			alias,
			"",
			defaultValue,
			fmt.Sprintf(`Disable explanatory output (default $%s).`, env),
		)
	}

	if err := cmd.PersistentFlags().MarkHidden("quiet"); err != nil {
		panic(err)
	}
}

func setupLogVerbose(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogVerbose = new(bool)

	var defaultValue bool
	for _, envName := range []string{
		"WERF_LOG_VERBOSE",
		"WERF_VERBOSE",
	} {
		if os.Getenv(envName) != "" {
			defaultValue = GetBoolEnvironmentDefaultFalse(envName)
			break
		}
	}

	for alias, env := range map[string]string{
		"log-verbose": "WERF_LOG_VERBOSE",
		"verbose":     "WERF_VERBOSE",
	} {
		cmd.PersistentFlags().BoolVarP(
			cmdData.LogVerbose,
			alias,
			"",
			defaultValue,
			fmt.Sprintf(`Enable verbose output (default $%s).`, env),
		)
	}

	if err := cmd.PersistentFlags().MarkHidden("verbose"); err != nil {
		panic(err)
	}
}

func setupLogPretty(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogPretty = new(bool)

	var defaultValue bool
	if os.Getenv("WERF_LOG_PRETTY") != "" {
		defaultValue = GetBoolEnvironmentDefaultFalse("WERF_LOG_PRETTY")
	} else {
		defaultValue = true
	}

	cmd.PersistentFlags().BoolVarP(cmdData.LogPretty, "log-pretty", "", defaultValue, `Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or true).`)
}

func setupTerminalWidth(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogTerminalWidth = new(int64)
	cmd.PersistentFlags().Int64VarP(cmdData.LogTerminalWidth, "log-terminal-width", "", -1, fmt.Sprintf(`Set log terminal width.
Defaults to:
* $WERF_LOG_TERMINAL_WIDTH
* interactive terminal width or %d`, 140))
}

func SetupSet(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Set = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.Set, "set", "", []string{}, `Set helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2).
Also, can be defined with $WERF_SET_* (e.g. $WERF_SET_1=key1=val1, $WERF_SET_2=key2=val2)`)
}

func SetupSetString(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SetString = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SetString, "set-string", "", []string{}, `Set STRING helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2).
Also, can be defined with $WERF_SET_STRING_* (e.g. $WERF_SET_STRING_1=key1=val1, $WERF_SET_STRING_2=key2=val2)`)
}

func SetupValues(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Values = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.Values, "values", "", []string{}, `Specify helm values in a YAML file or a URL (can specify multiple).
Also, can be defined with $WERF_VALUES_* (e.g. $WERF_VALUES_ENV=.helm/values_test.yaml, $WERF_VALUES_DB=.helm/values_db.yaml)`)
}

func SetupSetFile(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SetFile = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SetFile, "set-file", "", []string{}, `Set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2).
Also, can be defined with $WERF_SET_FILE_* (e.g. $WERF_SET_FILE_1=key1=path1, $WERF_SET_FILE_2=key2=val2)`)
}

func SetupSecretValues(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SecretValues = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SecretValues, "secret-values", "", []string{}, `Specify helm secret values in a YAML file (can specify multiple).
Also, can be defined with $WERF_SECRET_VALUES_* (e.g. $WERF_SECRET_VALUES_ENV=.helm/secret_values_test.yaml, $WERF_SECRET_VALUES_DB=.helm/secret_values_db.yaml)`)
}

func SetupIgnoreSecretKey(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.IgnoreSecretKey = new(bool)
	cmd.Flags().BoolVarP(cmdData.IgnoreSecretKey, "ignore-secret-key", "", GetBoolEnvironmentDefaultFalse("WERF_IGNORE_SECRET_KEY"), "Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)")
}

func SetupParallelOptions(cmdData *CmdData, cmd *cobra.Command, defaultValue int64) {
	SetupParallel(cmdData, cmd)
	SetupParallelTasksLimit(cmdData, cmd, defaultValue)
}

func SetupParallel(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Parallel = new(bool)
	cmd.Flags().BoolVarP(cmdData.Parallel, "parallel", "p", GetBoolEnvironmentDefaultTrue("WERF_PARALLEL"), "Run in parallel (default $WERF_PARALLEL or true)")
}

func SetupParallelTasksLimit(cmdData *CmdData, cmd *cobra.Command, defaultValue int64) {
	cmdData.ParallelTasksLimit = new(int64)
	cmd.Flags().Int64VarP(cmdData.ParallelTasksLimit, "parallel-tasks-limit", "", defaultValue, "Parallel tasks limit, set -1 to remove the limitation (default $WERF_PARALLEL_TASKS_LIMIT or 5)")
}

func SetupLogProjectDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogProjectDir = new(bool)
	cmd.Flags().BoolVarP(cmdData.LogProjectDir, "log-project-dir", "", GetBoolEnvironmentDefaultFalse("WERF_LOG_PROJECT_DIR"), `Print current project directory path (default $WERF_LOG_PROJECT_DIR)`)
}

func SetupIntrospectAfterError(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.IntrospectAfterError = new(bool)
	cmd.Flags().BoolVarP(cmdData.IntrospectAfterError, "introspect-error", "", false, "Introspect failed stage in the state, right after running failed assembly instruction")
}

func SetupIntrospectBeforeError(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.IntrospectBeforeError = new(bool)
	cmd.Flags().BoolVarP(cmdData.IntrospectBeforeError, "introspect-before-error", "", false, "Introspect failed stage in the clean state, before running all assembly instructions of the stage")
}

func SetupIntrospectStage(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StagesToIntrospect = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.StagesToIntrospect, "introspect-stage", "", []string{}, `Introspect a specific stage. The option can be used multiple times to introspect several stages.

There are the following formats to use:
* specify IMAGE_NAME/STAGE_NAME to introspect stage STAGE_NAME of either image or artifact IMAGE_NAME
* specify STAGE_NAME or */STAGE_NAME for the introspection of all existing stages with name STAGE_NAME

IMAGE_NAME is the name of an image or artifact described in werf.yaml, the nameless image specified with ~.
STAGE_NAME should be one of the following: `+strings.Join(allStagesNames(), ", "))
}

func SetupSkipBuild(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SkipBuild = new(bool)
	cmd.Flags().BoolVarP(cmdData.SkipBuild, "skip-build", "Z", GetBoolEnvironmentDefaultFalse("WERF_SKIP_BUILD"), "Disable building of docker images, cached images in the repo should exist in the repo if werf.yaml contains at least one image description (default $WERF_SKIP_BUILD)")
}

func SetupStubTags(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StubTags = new(bool)
	cmd.Flags().BoolVarP(cmdData.StubTags, "stub-tags", "", GetBoolEnvironmentDefaultFalse("WERF_STUB_TAGS"), "Use stubs instead of real tags (default $WERF_STUB_TAGS)")
}

func SetupFollow(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Follow = new(bool)
	cmd.Flags().BoolVarP(cmdData.Follow, "follow", "", GetBoolEnvironmentDefaultFalse("WERF_FOLLOW"), `Enable follow mode (default $WERF_FOLLOW).
The mode allows restarting the command on a new commit.
In development mode (--dev), werf restarts the command on any changes (including untracked files) in the git repository worktree`)
}

func allStagesNames() []string {
	var stageNames []string
	for _, stageName := range stage.AllStages {
		stageNames = append(stageNames, string(stageName))
	}

	return stageNames
}

func GetBoolEnvironmentDefaultFalse(environmentName string) bool {
	switch os.Getenv(environmentName) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

func GetBoolEnvironmentDefaultTrue(environmentName string) bool {
	switch os.Getenv(environmentName) {
	case "0", "false", "no":
		return false
	default:
		return true
	}
}

func getInt64EnvVar(varName string) (*int64, error) {
	if v := os.Getenv(varName); v != "" {
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value %q: %w", varName, v, err)
		}

		res := new(int64)
		*res = vInt

		return res, nil
	}

	return nil, nil
}

func GetIntEnvVarStrict(varName string) *int64 {
	valP, err := GetIntEnvVar(varName)
	if err != nil {
		TerminateWithError(fmt.Sprintf("bad %s value: %s", varName, err), 1)
	}
	return valP
}

func GetIntEnvVar(varName string) (*int64, error) {
	if v := os.Getenv(varName); v != "" {
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value %q: %w", varName, v, err)
		}

		res := new(int64)
		*res = vInt

		return res, nil
	}

	return nil, nil
}

func GetUint64EnvVarStrict(varName string) *uint64 {
	valP, err := GetUint64EnvVar(varName)
	if err != nil {
		TerminateWithError(fmt.Sprintf("bad %s value: %s", varName, err), 1)
	}
	return valP
}

func GetUint64EnvVar(varName string) (*uint64, error) {
	if v := os.Getenv(varName); v != "" {
		vUint, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value %q: %w", varName, v, err)
		}

		res := new(uint64)
		*res = vUint

		return res, nil
	}

	return nil, nil
}

func GetParallelTasksLimit(cmdData *CmdData) (int64, error) {
	v, err := getInt64EnvVar("WERF_PARALLEL_TASKS_LIMIT")
	if err != nil {
		return 0, err
	}
	if v == nil {
		v = cmdData.ParallelTasksLimit
	}
	if *v <= 0 {
		return -1, nil
	} else {
		return *v, nil
	}
}

func GetStagesStorageAddress(cmdData *CmdData) (string, error) {
	if *cmdData.StagesStorage == "" || *cmdData.StagesStorage == storage.LocalStorageAddress {
		return "", fmt.Errorf("--repo=ADDRESS param required")
	}

	return *cmdData.StagesStorage, nil
}

func GetOptionalStagesStorageAddress(cmdData *CmdData) string {
	if *cmdData.StagesStorage == "" {
		return storage.LocalStorageAddress
	}

	return *cmdData.StagesStorage
}

func GetLocalStagesStorage(containerBackend container_backend.ContainerBackend) (storage.StagesStorage, error) {
	return storage.NewStagesStorage(
		storage.LocalStorageAddress,
		containerBackend,
		storage.StagesStorageOptions{},
	)
}

func GetStagesStorage(stagesStorageAddress string, containerBackend container_backend.ContainerBackend, cmdData *CmdData) (storage.StagesStorage, error) {
	if _, match := containerBackend.(*container_backend.BuildahBackend); match {
		if stagesStorageAddress == "" || stagesStorageAddress == storage.LocalStorageAddress {
			return nil, fmt.Errorf(`"--repo" should be specified and not equal ":local" for Buildah container backend`)
		}
	}

	if err := ValidateRepoContainerRegistry(cmdData.CommonRepoData.GetContainerRegistry()); err != nil {
		return nil, err
	}

	return storage.NewStagesStorage(
		stagesStorageAddress,
		containerBackend,
		storage.StagesStorageOptions{
			RepoStagesStorageOptions: storage.RepoStagesStorageOptions{
				ContainerRegistry: cmdData.CommonRepoData.GetContainerRegistry(),
				DockerRegistryOptions: docker_registry.DockerRegistryOptions{
					InsecureRegistry:      *cmdData.InsecureRegistry,
					SkipTlsVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
					DockerHubUsername:     *cmdData.CommonRepoData.DockerHubUsername,
					DockerHubPassword:     *cmdData.CommonRepoData.DockerHubPassword,
					DockerHubToken:        *cmdData.CommonRepoData.DockerHubToken,
					GitHubToken:           *cmdData.CommonRepoData.GitHubToken,
					HarborUsername:        *cmdData.CommonRepoData.HarborUsername,
					HarborPassword:        *cmdData.CommonRepoData.HarborPassword,
					QuayToken:             *cmdData.CommonRepoData.QuayToken,
				},
			},
		},
	)
}

func GetOptionalFinalStagesStorage(containerBackend container_backend.ContainerBackend, cmdData *CmdData) (storage.StagesStorage, error) {
	finalRepoAddress := *cmdData.FinalStagesStorage
	if finalRepoAddress == "" {
		return nil, nil
	}

	if err := ValidateRepoContainerRegistry(cmdData.CommonFinalRepoData.GetContainerRegistry()); err != nil {
		return nil, err
	}

	return storage.NewStagesStorage(
		finalRepoAddress,
		containerBackend,
		storage.StagesStorageOptions{
			RepoStagesStorageOptions: storage.RepoStagesStorageOptions{
				ContainerRegistry: cmdData.CommonFinalRepoData.GetContainerRegistry(),
				DockerRegistryOptions: docker_registry.DockerRegistryOptions{
					InsecureRegistry:      *cmdData.InsecureRegistry,
					SkipTlsVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
					DockerHubUsername:     *cmdData.CommonFinalRepoData.DockerHubUsername,
					DockerHubPassword:     *cmdData.CommonFinalRepoData.DockerHubPassword,
					DockerHubToken:        *cmdData.CommonFinalRepoData.DockerHubToken,
					GitHubToken:           *cmdData.CommonFinalRepoData.GitHubToken,
					HarborUsername:        *cmdData.CommonFinalRepoData.HarborUsername,
					HarborPassword:        *cmdData.CommonFinalRepoData.HarborPassword,
					QuayToken:             *cmdData.CommonFinalRepoData.QuayToken,
				},
			},
		},
	)
}

func GetCacheStagesStorageList(containerBackend container_backend.ContainerBackend, cmdData *CmdData) ([]storage.StagesStorage, error) {
	var res []storage.StagesStorage

	for _, address := range GetCacheStagesStorage(cmdData) {
		repoStagesStorage, err := storage.NewStagesStorage(address, containerBackend, storage.StagesStorageOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to create cache stages storage at %s: %w", address, err)
		}
		res = append(res, repoStagesStorage)
	}

	return res, nil
}

func GetSecondaryStagesStorageList(stagesStorage storage.StagesStorage, containerBackend container_backend.ContainerBackend, cmdData *CmdData) ([]storage.StagesStorage, error) {
	var res []storage.StagesStorage

	if _, matched := containerBackend.(*container_backend.DockerServerBackend); matched {
		if stagesStorage.Address() != storage.LocalStorageAddress {
			localStagesStorage, err := storage.NewStagesStorage(storage.LocalStorageAddress, containerBackend, storage.StagesStorageOptions{})
			if err != nil {
				return nil, fmt.Errorf("unable to create local secondary stages storage: %w", err)
			}
			res = append(res, localStagesStorage)
		}
	}

	for _, address := range GetSecondaryStagesStorage(cmdData) {
		repoStagesStorage, err := storage.NewStagesStorage(address, containerBackend, storage.StagesStorageOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to create secondary stages storage at %s: %w", address, err)
		}
		res = append(res, repoStagesStorage)
	}

	return res, nil
}

func GetOptionalWerfConfig(ctx context.Context, cmdData *CmdData, giterminismManager giterminism_manager.Interface, opts config.WerfConfigOptions) (string, *config.WerfConfig, error) {
	customWerfConfigRelPath, err := GetCustomWerfConfigRelPath(giterminismManager, cmdData)
	if err != nil {
		return "", nil, err
	}

	exist, err := giterminismManager.FileReader().IsConfigExistAnywhere(ctx, customWerfConfigRelPath)
	if err != nil {
		return "", nil, err
	}

	if exist {
		customWerfConfigTemplatesDirRelPath, err := GetCustomWerfConfigTemplatesDirRelPath(giterminismManager, cmdData)
		if err != nil {
			return "", nil, err
		}

		configPath, c, err := config.GetWerfConfig(ctx, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, giterminismManager, opts)
		if err != nil {
			return "", nil, err
		}

		return configPath, c, nil
	}

	return "", nil, nil
}

func GetRequiredWerfConfig(ctx context.Context, cmdData *CmdData, giterminismManager giterminism_manager.Interface, opts config.WerfConfigOptions) (string, *config.WerfConfig, error) {
	customWerfConfigRelPath, err := GetCustomWerfConfigRelPath(giterminismManager, cmdData)
	if err != nil {
		return "", nil, err
	}

	customWerfConfigTemplatesDirRelPath, err := GetCustomWerfConfigTemplatesDirRelPath(giterminismManager, cmdData)
	if err != nil {
		return "", nil, err
	}

	configPath, c, err := config.GetWerfConfig(ctx, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, giterminismManager, opts)
	if err != nil {
		return "", nil, err
	}

	return configPath, c, nil
}

func GetCustomWerfConfigRelPath(giterminismManager giterminism_manager.Interface, cmdData *CmdData) (string, error) {
	customConfigPath := *cmdData.ConfigPath
	if customConfigPath == "" {
		return "", nil
	}

	customConfigPath = util.GetAbsoluteFilepath(customConfigPath)
	if !util.IsSubpathOfBasePath(giterminismManager.LocalGitRepo().GetWorkTreeDir(), customConfigPath) {
		return "", fmt.Errorf("the werf config %q must be in the project git work tree %q", customConfigPath, giterminismManager.LocalGitRepo().GetWorkTreeDir())
	}

	return util.GetRelativeToBaseFilepath(giterminismManager.ProjectDir(), customConfigPath), nil
}

func GetCustomWerfConfigTemplatesDirRelPath(giterminismManager giterminism_manager.Interface, cmdData *CmdData) (string, error) {
	customConfigTemplatesDirPath := *cmdData.ConfigTemplatesDir
	if customConfigTemplatesDirPath == "" {
		return "", nil
	}

	customConfigTemplatesDirPath = util.GetAbsoluteFilepath(customConfigTemplatesDirPath)
	if !util.IsSubpathOfBasePath(giterminismManager.LocalGitRepo().GetWorkTreeDir(), customConfigTemplatesDirPath) {
		return "", fmt.Errorf("the werf configuration templates directory %q must be in the project git work tree %q", customConfigTemplatesDirPath, giterminismManager.LocalGitRepo().GetWorkTreeDir())
	}

	return util.GetRelativeToBaseFilepath(giterminismManager.ProjectDir(), customConfigTemplatesDirPath), nil
}

func GetWerfConfigOptions(cmdData *CmdData, logRenderedFilePath bool) config.WerfConfigOptions {
	return config.WerfConfigOptions{
		LogRenderedFilePath: logRenderedFilePath,
		Env:                 *cmdData.Environment,
	}
}

func GetGiterminismManager(ctx context.Context, cmdData *CmdData) (giterminism_manager.Interface, error) {
	workingDir := GetWorkingDir(cmdData)

	gitWorkTree, err := GetGitWorkTree(ctx, cmdData, workingDir)
	if err != nil {
		return nil, err
	}

	isWorkingDirInsideGitWorkTree := util.IsSubpathOfBasePath(gitWorkTree, workingDir)
	areWorkingDirAndGitWorkTreeTheSame := gitWorkTree == workingDir
	if !(isWorkingDirInsideGitWorkTree || areWorkingDirAndGitWorkTreeTheSame) {
		return nil, fmt.Errorf("werf requires project dir — the current working directory or directory specified with --dir option (or WERF_DIR env var) — to be located inside the git work tree: %q is located outside of the git work tree %q", gitWorkTree, workingDir)
	}

	var openLocalRepoOptions git_repo.OpenLocalRepoOptions
	if *cmdData.Dev {
		openLocalRepoOptions.WithServiceHeadCommit = true
		openLocalRepoOptions.ServiceBranchOptions.Name = *cmdData.DevBranch
		openLocalRepoOptions.ServiceBranchOptions.GlobExcludeList = GetDevIgnore(cmdData)
	}

	localGitRepo, err := git_repo.OpenLocalRepo(GetContext(), "own", gitWorkTree, openLocalRepoOptions)
	if err != nil {
		return nil, err
	}

	headCommit, err := localGitRepo.HeadCommitHash(GetContext())
	if err != nil {
		return nil, err
	}

	return giterminism_manager.NewManager(GetContext(), workingDir, localGitRepo, headCommit, giterminism_manager.NewManagerOptions{
		LooseGiterminism: *cmdData.LooseGiterminism,
		Dev:              *cmdData.Dev,
	})
}

func GetGitWorkTree(ctx context.Context, cmdData *CmdData, workingDir string) (string, error) {
	if *cmdData.GitWorkTree != "" {
		workTree := *cmdData.GitWorkTree

		if isValid, err := true_git.IsValidWorkTree(ctx, workTree); err != nil {
			return "", err
		} else if isValid {
			return util.GetAbsoluteFilepath(workTree), nil
		}

		return "", fmt.Errorf("werf requires a git work tree for the project to exist: not a valid git work tree %q specified", workTree)
	}

	if found, workTree, err := true_git.UpwardLookupAndVerifyWorkTree(ctx, workingDir); err != nil {
		return "", err
	} else if found {
		return util.GetAbsoluteFilepath(workTree), nil
	}

	return "", fmt.Errorf("werf requires a git work tree for the project to exist: unable to find a valid .git in the current directory %q or parent directories, you may also specify git work tree explicitly with --git-work-tree option (or WERF_GIT_WORK_TREE env var)", util.GetAbsoluteFilepath("."))
}

func GetWorkingDir(cmdData *CmdData) string {
	var workingDir string
	if *cmdData.Dir != "" {
		workingDir = *cmdData.Dir
	} else {
		workingDir = "."
	}
	return util.GetAbsoluteFilepath(workingDir)
}

func GetHelmChartDir(werfConfigPath string, werfConfig *config.WerfConfig, giterminismManager giterminism_manager.Interface) (string, error) {
	var helmChartDir string
	if werfConfig.Meta.Deploy.HelmChartDir != nil && *werfConfig.Meta.Deploy.HelmChartDir != "" {
		helmChartDir = *werfConfig.Meta.Deploy.HelmChartDir
	} else {
		helmChartDir = filepath.Join(filepath.Dir(werfConfigPath), ".helm")
	}

	absHelmChartDir := filepath.Join(giterminismManager.ProjectDir(), helmChartDir)
	if !util.IsSubpathOfBasePath(giterminismManager.LocalGitRepo().GetWorkTreeDir(), absHelmChartDir) {
		return "", fmt.Errorf("the chart directory %s must be in the project git work tree %s", absHelmChartDir, giterminismManager.LocalGitRepo().GetWorkTreeDir())
	}

	return helmChartDir, nil
}

func GetNamespace(cmdData *CmdData) string {
	if *cmdData.Namespace == "" {
		return "default"
	}
	return *cmdData.Namespace
}

func GetDevIgnore(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_DEV_IGNORE_"), *cmdData.DevIgnore...)
}

func GetSSHKey(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_SSH_KEY_"), *cmdData.SSHKeys...)
}

func GetAddLabels(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_ADD_LABEL_"), *cmdData.AddLabels...)
}

func GetAddAnnotations(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_ADD_ANNOTATION_"), *cmdData.AddAnnotations...)
}

func GetCacheStagesStorage(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_CACHE_REPO_"), *cmdData.CacheStagesStorage...)
}

func GetSecondaryStagesStorage(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_SECONDARY_REPO_"), *cmdData.SecondaryStagesStorage...)
}

func getAddCustomTag(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_ADD_CUSTOM_TAG_"), *cmdData.AddCustomTag...)
}

func GetSet(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_SET_", "WERF_SET_STRING_", "WERF_SET_FILE_", "WERF_SET_DOCKER_CONFIG_JSON_VALUE"), *cmdData.Set...)
}

func GetSetString(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_SET_STRING_"), *cmdData.SetString...)
}

func GetSetFile(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_SET_FILE_"), *cmdData.SetFile...)
}

func GetValues(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_VALUES_"), *cmdData.Values...)
}

func GetSecretValues(cmdData *CmdData) []string {
	return append(PredefinedValuesByEnvNamePrefix("WERF_SECRET_VALUES_"), *cmdData.SecretValues...)
}

func GetRequiredRelease(cmdData *CmdData) (string, error) {
	if *cmdData.Release == "" {
		return "", fmt.Errorf("--release=RELEASE param required")
	}
	return *cmdData.Release, nil
}

func GetOptionalRelease(cmdData *CmdData) string {
	if *cmdData.Release == "" {
		return "werf-stub"
	}
	return *cmdData.Release
}

func GetIntrospectOptions(cmdData *CmdData, werfConfig *config.WerfConfig) (build.IntrospectOptions, error) {
	isStageExist := func(sName string) bool {
		for _, stageName := range allStagesNames() {
			if sName == stageName {
				return true
			}
		}

		return false
	}

	introspectOptions := build.IntrospectOptions{}
	for _, imageAndStage := range *cmdData.StagesToIntrospect {
		var imageName, stageName string

		parts := strings.SplitN(imageAndStage, "/", 2)
		if len(parts) == 1 {
			imageName = "*"
			stageName = parts[0]
		} else {
			if parts[0] != "~" {
				imageName = parts[0]
			}

			stageName = parts[1]
		}

		if imageName != "*" && !werfConfig.HasImageOrArtifact(imageName) {
			return introspectOptions, fmt.Errorf("specified image %s (%s) is not defined in werf.yaml", imageName, imageAndStage)
		}

		if !isStageExist(stageName) {
			return introspectOptions, fmt.Errorf("specified stage name %s (%s) is not exist", stageName, imageAndStage)
		}

		introspectTarget := build.IntrospectTarget{ImageName: imageName, StageName: stageName}
		introspectOptions.Targets = append(introspectOptions.Targets, introspectTarget)
	}

	return introspectOptions, nil
}

func LogKubeContext(kubeContext string) {
	if kubeContext != "" {
		logboek.LogF("Using kube context: %s\n", kubeContext)
	}
}

func ProcessLogProjectDir(cmdData *CmdData, projectDir string) {
	if *cmdData.LogProjectDir {
		logboek.LogF("Using project dir: %s\n", projectDir)
	}
}

func ProcessLogOptions(cmdData *CmdData) error {
	if err := ProcessLogColorMode(cmdData); err != nil {
		return err
	}

	switch {
	case *cmdData.LogDebug:
		logboek.SetAcceptedLevel(level.Debug)
		logboek.Streams().EnablePrefixWithTime()
		logboek.Streams().SetPrefixStyle(style.Details())
	case *cmdData.LogVerbose:
		logboek.SetAcceptedLevel(level.Info)
	case *cmdData.LogQuiet:
		logboek.SetAcceptedLevel(level.Error)
	}

	if !*cmdData.LogPretty {
		logboek.Streams().DisablePrettyLog()
		logging.DisablePrettyLog()
	}

	if err := ProcessLogTerminalWidth(cmdData); err != nil {
		return err
	}

	return nil
}

func ProcessLogColorMode(cmdData *CmdData) error {
	logColorMode := *cmdData.LogColorMode

	switch logColorMode {
	case "auto":
	case "on":
		logboek.Streams().EnableStyle()
	case "off":
		logboek.Streams().DisableStyle()
	default:
		return fmt.Errorf("bad log color mode %q: on, off and auto modes are supported", logColorMode)
	}

	return nil
}

func ProcessLogTerminalWidth(cmdData *CmdData) error {
	value := *cmdData.LogTerminalWidth

	if value != -1 {
		if value < 0 {
			return fmt.Errorf("--log-terminal-width parameter (%d) can not be negative", value)
		}

		logboek.Streams().SetWidth(int(value))
	} else {
		pInt64, err := getInt64EnvVar("WERF_LOG_TERMINAL_WIDTH")
		if err != nil {
			return err
		}

		if pInt64 == nil {
			return nil
		}

		if *pInt64 < 0 {
			return fmt.Errorf("WERF_LOG_TERMINAL_WIDTH value (%s) can not be negative", os.Getenv("WERF_LOG_TERMINAL_WIDTH"))
		}

		logboek.Streams().SetWidth(int(*pInt64))
	}

	return nil
}

func DockerRegistryInit(ctx context.Context, cmdData *CmdData) error {
	return docker_registry.Init(ctx, *cmdData.InsecureRegistry, *cmdData.SkipTlsVerifyRegistry)
}

func ValidateRepoContainerRegistry(containerRegistry string) error {
	supportedValues := docker_registry.ImplementationList()
	supportedValues = append(supportedValues, "auto", "")

	for _, supportedContainerRegistry := range supportedValues {
		if supportedContainerRegistry == containerRegistry {
			return nil
		}
	}

	return fmt.Errorf("specified container registry %q is not supported", containerRegistry)
}

func ValidateMinimumNArgs(minArgs int, args []string, cmd *cobra.Command) error {
	if len(args) < minArgs {
		PrintHelp(cmd)
		return fmt.Errorf("requires at least %d arg(s), received %d", minArgs, len(args))
	}

	return nil
}

func ValidateArgumentCount(expectedCount int, args []string, cmd *cobra.Command) error {
	if len(args) != expectedCount {
		PrintHelp(cmd)
		return fmt.Errorf("requires %d position argument(s), received %d", expectedCount, len(args))
	}

	return nil
}

func PrintHelp(cmd *cobra.Command) {
	_ = cmd.Help()
	logboek.LogOptionalLn()
}

func LogRunningTime(f func() error) error {
	t := time.Now()
	err := f()

	logboek.Default().LogFHighlight("Running time %0.2f seconds\n", time.Since(t).Seconds())

	return err
}

func LogVersion() {
	logboek.LogF("Version: %s\n", werf.Version)
}

func TerminateWithError(errMsg string, exitCode int) {
	msg := fmt.Sprintf("Error: %s", errMsg)
	msg = strings.TrimSuffix(msg, "\n")

	logboek.Streams().DisableLineWrapping()
	logboek.Error().LogLn(msg)
	os.Exit(exitCode)
}

func SetupVirtualMerge(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.VirtualMerge = new(bool)
	cmd.Flags().BoolVarP(cmdData.VirtualMerge, "virtual-merge", "", GetBoolEnvironmentDefaultFalse("WERF_VIRTUAL_MERGE"), "Enable virtual/ephemeral merge commit mode when building current application state ($WERF_VIRTUAL_MERGE by default)")
}

func SetupPlatform(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Platform = new(string)

	var defaultValue string

	for _, envName := range []string{
		"WERF_PLATFORM",
		"DOCKER_DEFAULT_PLATFORM",
	} {
		if v := os.Getenv(envName); v != "" {
			defaultValue = v
			break
		}
	}

	cmd.Flags().StringVarP(cmdData.Platform, "platform", "", defaultValue, "Enable platform emulation when building images with werf. The only supported option for now is linux/amd64.")
}

func GetContext() context.Context {
	return logboek.NewContext(context.Background(), logboek.DefaultLogger())
}

func WithContext(allowBackgroundMode bool, f func(ctx context.Context) error) error {
	var ctx context.Context

	if allowBackgroundMode && IsBackgroundModeEnabled() {
		out, err := os.OpenFile(GetBackgroundOutputFile(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			return fmt.Errorf("unable to open background output file %q: %w", GetBackgroundOutputFile(), err)
		}
		defer out.Close()

		ctx = logboek.NewContext(context.Background(), logboek.NewLogger(out, out))

		if err := f(ctx); err != nil {
			if err := os.WriteFile(GetLastBackgroundErrorFile(), []byte(err.Error()+"\n"), 0o644); err != nil {
				logboek.Context(ctx).Warn().LogF("ERROR: unable to write %q: %s\n", GetLastBackgroundErrorFile(), err)
			}
			return err
		}

		return nil
	} else {
		ctx = logboek.NewContext(context.Background(), logboek.DefaultLogger())

		if allowBackgroundMode {
			if backgroundErr, err := GetAndRemoveLastBackgroundError(); err != nil {
				return fmt.Errorf("unable to get last background error: %w", err)
			} else if backgroundErr != nil {
				global_warnings.GlobalWarningLn(ctx, fmt.Sprintf("last background error: %s", backgroundErr))
			}
		}

		return f(ctx)
	}
}

func GetAndRemoveLastBackgroundError() (error, error) {
	data, err := os.ReadFile(GetLastBackgroundErrorFile())
	if err != nil {
		return nil, nil
	}

	if err := os.RemoveAll(GetLastBackgroundErrorFile()); err != nil {
		return nil, fmt.Errorf("unable to remove %q: %w", GetLastBackgroundErrorFile(), err)
	}

	if len(data) != 0 {
		return fmt.Errorf("%s", string(data)), nil
	}

	return nil, nil
}

func IsBackgroundModeEnabled() bool {
	return os.Getenv("_WERF_BACKGROUND_MODE_ENABLED") == "1"
}

func GetBackgroundOutputFile() string {
	return filepath.Join(werf.GetServiceDir(), "background_output.log")
}

func GetLastBackgroundErrorFile() string {
	return filepath.Join(werf.GetServiceDir(), "last_background_error")
}

func getFlags(cmd *cobra.Command, persistent bool) *pflag.FlagSet {
	if persistent {
		return cmd.PersistentFlags()
	}

	return cmd.Flags()
}
