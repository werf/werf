package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/build/stage"
	cleanup "github.com/werf/werf/pkg/cleaning"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

type CmdData struct {
	ProjectName        *string
	Dir                *string
	ConfigPath         *string
	ConfigTemplatesDir *string
	TmpDir             *string
	HomeDir            *string
	SSHKeys            *[]string

	TagCustom            *[]string
	TagGitBranch         *string
	TagGitTag            *string
	TagGitCommit         *string
	TagByStagesSignature *bool

	HelmChartDir                     *string
	Environment                      *string
	Release                          *string
	Namespace                        *string
	AddAnnotations                   *[]string
	AddLabels                        *[]string
	KubeContext                      *string
	KubeConfig                       *string
	HelmReleaseStorageNamespace      *string
	HelmReleaseStorageType           *string
	StatusProgressPeriodSeconds      *int64
	HooksStatusProgressPeriodSeconds *int64
	ReleasesHistoryMax               *int

	Set             *[]string
	SetString       *[]string
	Values          *[]string
	SecretValues    *[]string
	IgnoreSecretKey *bool

	CommonRepoData *RepoData

	StagesStorage         *string
	StagesStorageRepoData *RepoData

	ImagesRepo     *string
	ImagesRepoMode *string
	ImagesRepoData *RepoData

	Synchronization *string

	DockerConfig          *string
	InsecureRegistry      *bool
	SkipTlsVerifyRegistry *bool
	DryRun                *bool

	GitTagStrategyLimit               *int64
	GitTagStrategyExpiryDays          *int64
	GitCommitStrategyLimit            *int64
	GitCommitStrategyExpiryDays       *int64
	StagesSignatureStrategyLimit      *int64
	StagesSignatureStrategyExpiryDays *int64

	WithoutKube *bool

	StagesToIntrospect *[]string

	LogDebug         *bool
	LogPretty        *bool
	LogVerbose       *bool
	LogQuiet         *bool
	LogColorMode     *string
	LogProjectDir    *bool
	LogTerminalWidth *int64

	ThreeWayMergeMode *string

	PublishReportPath   *string
	PublishReportFormat *string

	VirtualMerge           *bool
	VirtualMergeFromCommit *string
	VirtualMergeIntoCommit *string
}

const (
	CleaningCommandsForceOptionDescription = "First remove containers that use werf docker images which are going to be deleted"
	StubImagesRepoAddress                  = "stub/repository"
)

func GetLongCommandDescription(text string) string {
	return logboek.FitText(text, logboek.FitTextOptions{MaxWidth: 100})
}

func SetupProjectName(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ProjectName = new(string)
	cmd.Flags().StringVarP(cmdData.ProjectName, "project-name", "N", os.Getenv("WERF_PROJECT_NAME"), "Use custom project name (default $WERF_PROJECT_NAME)")
}

func SetupDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Dir = new(string)
	cmd.Flags().StringVarP(cmdData.Dir, "dir", "", os.Getenv("WERF_DIR"), "Use custom working directory (default $WERF_DIR or current directory)")
}

func SetupConfigPath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ConfigPath = new(string)
	cmd.Flags().StringVarP(cmdData.ConfigPath, "config", "", os.Getenv("WERF_CONFIG"), `Use custom configuration file (default $WERF_CONFIG or werf.yaml in working directory)`)
}

func SetupConfigTemplatesDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ConfigTemplatesDir = new(string)
	cmd.Flags().StringVarP(cmdData.ConfigTemplatesDir, "config-templates-dir", "", os.Getenv("WERF_CONFIG_TEMPLATES_DIR"), `Change to the custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf in working directory)`)
}

func SetupTmpDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.TmpDir = new(string)
	cmd.Flags().StringVarP(cmdData.TmpDir, "tmp-dir", "", "", "Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)")
}

func SetupHomeDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HomeDir = new(string)
	cmd.Flags().StringVarP(cmdData.HomeDir, "home-dir", "", "", "Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)")
}

func SetupSSHKey(cmdData *CmdData, cmd *cobra.Command) {
	sshKeys := predefinedValuesByEnvNamePrefix("WERF_SSH_KEY")

	cmdData.SSHKeys = &sshKeys
	cmd.Flags().StringArrayVarP(cmdData.SSHKeys, "ssh-key", "", sshKeys, `Use only specific ssh key(s).
Can be specified with $WERF_SSH_KEY* (e.g. $WERF_SSH_KEY_REPO=~/.ssh/repo_rsa", $WERF_SSH_KEY_NODEJS=~/.ssh/nodejs_rsa").
Defaults to $WERF_SSH_KEY*, system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see https://werf.io/documentation/reference/toolbox/ssh.html`)
}

func SetupPublishReportPath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.PublishReportPath = new(string)
	cmd.Flags().StringVarP(cmdData.PublishReportPath, "publish-report-path", "", os.Getenv("WERF_PUBLISH_REPORT_PATH"), "Publish report contains image info: full docker repo, tag, ID — for each published image ($WERF_PUBLISH_REPORT_PATH by default)")
}

func SetupPublishReportFormat(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.PublishReportFormat = new(string)
	cmd.Flags().StringVarP(cmdData.PublishReportFormat, "publish-report-format", "", "json", "Publish report format (only json available for now, $WERF_PUBLISH_REPORT_FORMAT by default)")
}

func GetPublishReportFormat(cmdData *CmdData) (build.PublishReportFormat, error) {
	switch format := build.PublishReportFormat(*cmdData.PublishReportFormat); format {
	case build.PublishReportJSON:
		return format, nil
	default:
		return "", fmt.Errorf("bad --publish-report-format given %q, expected: \"json\"", format)
	}
}

func SetupImagesCleanupPolicies(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.GitTagStrategyLimit = new(int64)
	cmdData.GitTagStrategyExpiryDays = new(int64)
	cmdData.GitCommitStrategyLimit = new(int64)
	cmdData.GitCommitStrategyExpiryDays = new(int64)
	cmdData.StagesSignatureStrategyLimit = new(int64)
	cmdData.StagesSignatureStrategyExpiryDays = new(int64)

	cmd.Flags().Int64VarP(cmdData.GitTagStrategyLimit, "git-tag-strategy-limit", "", -1, "Keep max number of images published with the git-tag tagging strategy in the images repo. No limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_TAG_STRATEGY_LIMIT")
	cmd.Flags().Int64VarP(cmdData.GitTagStrategyExpiryDays, "git-tag-strategy-expiry-days", "", -1, "Keep images published with the git-tag tagging strategy in the images repo for the specified maximum days since image published. Republished image will be kept specified maximum days since new publication date. No days limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS")
	cmd.Flags().Int64VarP(cmdData.GitCommitStrategyLimit, "git-commit-strategy-limit", "", -1, "Keep max number of images published with the git-commit tagging strategy in the images repo. No limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_COMMIT_STRATEGY_LIMIT")
	cmd.Flags().Int64VarP(cmdData.GitCommitStrategyExpiryDays, "git-commit-strategy-expiry-days", "", -1, "Keep images published with the git-commit tagging strategy in the images repo for the specified maximum days since image published. Republished image will be kept specified maximum days since new publication date. No days limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS")
	cmd.Flags().Int64VarP(cmdData.StagesSignatureStrategyLimit, "stages-signature-strategy-limit", "", -1, "Keep max number of images published with the stages-signature tagging strategy in the images repo. No limit by default, -1 disables the limit. Value can be specified by the $WERF_STAGES_SIGNATURE_STRATEGY_LIMIT")
	cmd.Flags().Int64VarP(cmdData.StagesSignatureStrategyExpiryDays, "stages-signature-strategy-expiry-days", "", -1, "Keep images published with the stages-signature tagging strategy in the images repo for the specified maximum days since image published. Republished image will be kept specified maximum days since new publication date. No days limit by default, -1 disables the limit. Value can be specified by the $WERF_STAGES_SIGNATURE_STRATEGY_EXPIRY_DAYS")

	_ = cmd.Flags().MarkHidden("stages-signature-strategy-limit")
	_ = cmd.Flags().MarkHidden("stages-signature-strategy-expiry-days")
}

func SetupWithoutKube(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.WithoutKube = new(bool)
	cmd.Flags().BoolVarP(cmdData.WithoutKube, "without-kube", "", GetBoolEnvironmentDefaultFalse("WERF_WITHOUT_KUBE"), "Do not skip deployed Kubernetes images (default $WERF_KUBE_CONTEXT)")
}

func SetupTag(cmdData *CmdData, cmd *cobra.Command) {
	tagCustom := predefinedValuesByEnvNamePrefix("WERF_TAG_CUSTOM")

	cmdData.TagCustom = &tagCustom
	cmdData.TagGitBranch = new(string)
	cmdData.TagGitTag = new(string)
	cmdData.TagGitCommit = new(string)
	cmdData.TagByStagesSignature = new(bool)

	cmd.Flags().StringArrayVarP(cmdData.TagCustom, "tag-custom", "", tagCustom, "Use custom tagging strategy and tag by the specified arbitrary tags.\nOption can be used multiple times to produce multiple images with the specified tags.\nAlso can be specified in $WERF_TAG_CUSTOM* (e.g. $WERF_TAG_CUSTOM_TAG1=tag1, $WERF_TAG_CUSTOM_TAG2=tag2)")
	cmd.Flags().StringVarP(cmdData.TagGitBranch, "tag-git-branch", "", os.Getenv("WERF_TAG_GIT_BRANCH"), "Use git-branch tagging strategy and tag by the specified git branch (option can be enabled by specifying git branch in the $WERF_TAG_GIT_BRANCH)")
	cmd.Flags().StringVarP(cmdData.TagGitTag, "tag-git-tag", "", os.Getenv("WERF_TAG_GIT_TAG"), "Use git-tag tagging strategy and tag by the specified git tag (option can be enabled by specifying git tag in the $WERF_TAG_GIT_TAG)")
	cmd.Flags().StringVarP(cmdData.TagGitCommit, "tag-git-commit", "", os.Getenv("WERF_TAG_GIT_COMMIT"), "Use git-commit tagging strategy and tag by the specified git commit hash (option can be enabled by specifying git commit hash in the $WERF_TAG_GIT_COMMIT)")
	cmd.Flags().BoolVarP(cmdData.TagByStagesSignature, "tag-by-stages-signature", "", GetBoolEnvironmentDefaultFalse("WERF_TAG_BY_STAGES_SIGNATURE"), "Use stages-signature tagging strategy and tag each image by the corresponding signature of last image stage (option can be enabled by specifying $WERF_TAG_BY_STAGES_SIGNATURE=true)")
}

func predefinedValuesByEnvNamePrefix(envNamePrefix string, envNamePrefixesToExcept ...string) []string {
	var result []string

environLoop:
	for _, keyValue := range os.Environ() {
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

func SetupHelmChartDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HelmChartDir = new(string)
	cmd.Flags().StringVarP(cmdData.HelmChartDir, "helm-chart-dir", "", os.Getenv("WERF_HELM_CHART_DIR"), "Use custom helm chart dir (default $WERF_HELM_CHART_DIR or .helm in working directory)")
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
	addAnnotations := predefinedValuesByEnvNamePrefix("WERF_ADD_ANNOTATION")

	cmdData.AddAnnotations = &addAnnotations
	cmd.Flags().StringArrayVarP(cmdData.AddAnnotations, "add-annotation", "", addAnnotations, `Add annotation to deploying resources (can specify multiple).
Format: annoName=annoValue.
Also, can be specified with $WERF_ADD_ANNOTATION* (e.g. $WERF_ADD_ANNOTATION_1=annoName1=annoValue1", $WERF_ADD_ANNOTATION_2=annoName2=annoValue2")`)
}

func SetupAddLabels(cmdData *CmdData, cmd *cobra.Command) {
	addLabels := predefinedValuesByEnvNamePrefix("WERF_ADD_LABEL")

	cmdData.AddLabels = &addLabels
	cmd.Flags().StringArrayVarP(cmdData.AddLabels, "add-label", "", addLabels, `Add label to deploying resources (can specify multiple).
Format: labelName=labelValue.
Also, can be specified with $WERF_ADD_LABEL* (e.g. $WERF_ADD_LABEL_1=labelName1=labelValue1", $WERF_ADD_LABEL_2=labelName2=labelValue2")`)
}

func SetupKubeContext(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeContext = new(string)
	cmd.Flags().StringVarP(cmdData.KubeContext, "kube-context", "", os.Getenv("WERF_KUBE_CONTEXT"), "Kubernetes config context (default $WERF_KUBE_CONTEXT)")
}

func SetupKubeConfig(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeConfig = new(string)
	cmd.Flags().StringVarP(cmdData.KubeConfig, "kube-config", "", os.Getenv("WERF_KUBE_CONFIG"), "Kubernetes config file path (default $WERF_KUBE_CONFIG)")
}

func SetupHelmReleaseStorageNamespace(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HelmReleaseStorageNamespace = new(string)

	defaultValues := []string{
		os.Getenv("WERF_HELM_RELEASE_STORAGE_NAMESPACE"),
		os.Getenv("TILLER_NAMESPACE"),
		helm.DefaultReleaseStorageNamespace,
	}

	var defaultValue string
	for _, value := range defaultValues {
		if value != "" {
			defaultValue = value
			break
		}
	}

	cmd.Flags().StringVarP(cmdData.HelmReleaseStorageNamespace, "helm-release-storage-namespace", "", defaultValue, fmt.Sprintf("Helm release storage namespace (same as --tiller-namespace for regular helm, default $WERF_HELM_RELEASE_STORAGE_NAMESPACE, $TILLER_NAMESPACE or '%s')", helm.DefaultReleaseStorageNamespace))
}

func SetupHelmReleaseStorageType(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HelmReleaseStorageType = new(string)

	defaultValue := os.Getenv("WERF_HELM_RELEASE_STORAGE_TYPE")
	if defaultValue == "" {
		defaultValue = helm.ConfigMapStorage
	}

	cmd.Flags().StringVarP(cmdData.HelmReleaseStorageType, "helm-release-storage-type", "", defaultValue, fmt.Sprintf("helm storage driver to use. One of '%[1]s' or '%[2]s' (default $WERF_HELM_RELEASE_STORAGE_TYPE or '%[1]s')", helm.ConfigMapStorage, helm.SecretStorage))
}

func SetupCommonRepoData(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.CommonRepoData = &RepoData{IsCommon: true}

	SetupImplementationForRepoData(cmdData.CommonRepoData, cmd, "repo-implementation", []string{"WERF_REPO_IMPLEMENTATION"})
	SetupDockerHubUsernameForRepoData(cmdData.CommonRepoData, cmd, "repo-docker-hub-username", []string{"WERF_REPO_DOCKER_HUB_USERNAME"})
	SetupDockerHubPasswordForRepoData(cmdData.CommonRepoData, cmd, "repo-docker-hub-password", []string{"WERF_REPO_DOCKER_HUB_PASSWORD"})
	SetupDockerHubTokenForRepoData(cmdData.CommonRepoData, cmd, "repo-docker-hub-token", []string{"WERF_REPO_DOCKER_HUB_TOKEN"})
	SetupGithubTokenForRepoData(cmdData.CommonRepoData, cmd, "repo-github-token", []string{"WERF_REPO_GITHUB_TOKEN"})
	SetupHarborUsernameForRepoData(cmdData.CommonRepoData, cmd, "repo-harbor-username", []string{"WERF_REPO_HARBOR_USERNAME"})
	SetupHarborPasswordForRepoData(cmdData.CommonRepoData, cmd, "repo-harbor-password", []string{"WERF_REPO_HARBOR_PASSWORD"})
	SetupQuayTokenForRepoData(cmdData.CommonRepoData, cmd, "repo-quay-token", []string{"WERF_REPO_QUAY_TOKEN"})
}

func SetupStagesStorageOptions(cmdData *CmdData, cmd *cobra.Command) {
	// TODO: add the following options for images repo
	SetupInsecureRegistry(cmdData, cmd)
	SetupSkipTlsVerifyRegistry(cmdData, cmd)

	if cmdData.CommonRepoData == nil {
		SetupCommonRepoData(cmdData, cmd)
	}

	setupStagesStorage(cmdData, cmd)

	cmdData.StagesStorageRepoData = &RepoData{DesignationStorageName: "stages storage"}

	SetupImplementationForRepoData(cmdData.StagesStorageRepoData, cmd, "stages-storage-repo-implementation", []string{"WERF_STAGES_STORAGE_REPO_IMPLEMENTATION", "WERF_REPO_IMPLEMENTATION"})
	SetupDockerHubUsernameForRepoData(cmdData.StagesStorageRepoData, cmd, "stages-storage-repo-docker-hub-username", []string{"WERF_STAGES_STORAGE_REPO_DOCKER_HUB_USERNAME", "WERF_REPO_DOCKER_HUB_USERNAME"})
	SetupDockerHubPasswordForRepoData(cmdData.StagesStorageRepoData, cmd, "stages-storage-repo-docker-hub-password", []string{"WERF_STAGES_STORAGE_REPO_DOCKER_HUB_PASSWORD", "WERF_REPO_DOCKER_HUB_PASSWORD"})
	SetupDockerHubTokenForRepoData(cmdData.StagesStorageRepoData, cmd, "stages-storage-repo-docker-hub-token", []string{"WERF_STAGES_STORAGE_REPO_DOCKER_HUB_TOKEN", "WERF_REPO_DOCKER_HUB_TOKEN"})
	SetupGithubTokenForRepoData(cmdData.StagesStorageRepoData, cmd, "stages-storage-repo-github-token", []string{"WERF_STAGES_STORAGE_REPO_GITHUB_TOKEN", "WERF_REPO_GITHUB_TOKEN"})
	SetupHarborUsernameForRepoData(cmdData.StagesStorageRepoData, cmd, "stages-storage-repo-harbor-username", []string{"WERF_STAGES_STORAGE_REPO_HARBOR_USERNAME", "WERF_REPO_HARBOR_USERNAME"})
	SetupHarborPasswordForRepoData(cmdData.StagesStorageRepoData, cmd, "stages-storage-repo-harbor-password", []string{"WERF_STAGES_STORAGE_REPO_HARBOR_PASSWORD", "WERF_REPO_HARBOR_PASSWORD"})
	SetupQuayTokenForRepoData(cmdData.StagesStorageRepoData, cmd, "stages-storage-repo-quay-token", []string{"WERF_STAGES_STORAGE_REPO_QUAY_TOKEN", "WERF_REPO_QUAY_TOKEN"})
}

func setupStagesStorage(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.StagesStorage, "stages-storage", "s", os.Getenv("WERF_STAGES_STORAGE"), fmt.Sprintf("Docker Repo to store stages or %[1]s for non-distributed build (only %[1]s is supported for now; default $WERF_STAGES_STORAGE environment).\nMore info about stages: https://werf.io/documentation/reference/stages_and_images.html", storage.LocalStorageAddress))
}

func SetupSynchronization(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Synchronization = new(string)

	defaultValue := os.Getenv("WERF_SYNCHRONIZATION")

	cmd.Flags().StringVarP(cmdData.Synchronization, "synchronization", "S", defaultValue, fmt.Sprintf("Address of synchronizer for multiple werf processes to work with a single stages storage (default :local if --stages-storage=:local or %s if non-local stages-storage specified or $WERF_SYNCHRONIZATION if set). The same address should be specified for all werf processes that work with a single stages storage. :local address allows execution of werf processes from a single host only.", storage.DefaultKubernetesStorageAddress))
}

func SetupStatusProgressPeriod(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StatusProgressPeriodSeconds = new(int64)
	cmd.Flags().Int64VarP(
		cmdData.StatusProgressPeriodSeconds,
		"status-progress-period",
		"",
		*statusProgressPeriodDefaultValue(),
		"Status progress period in seconds. Set -1 to stop showing status progress. Defaults to $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds",
	)
}

func SetupReleasesHistoryMax(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ReleasesHistoryMax = new(int)

	defaultValueP, err := getIntEnvVar("WERF_RELEASES_HISTORY_MAX")
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

	v, err := getIntEnvVar("WERF_STATUS_PROGRESS_PERIOD_SECONDS")
	if err != nil {
		TerminateWithError(err.Error(), 1)
	}

	if v == nil {
		return &defaultValue
	} else {
		return v
	}
}

func SetupHooksStatusProgressPeriod(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HooksStatusProgressPeriodSeconds = new(int64)
	cmd.Flags().Int64VarP(
		cmdData.HooksStatusProgressPeriodSeconds,
		"hooks-status-progress-period",
		"",
		*hooksStatusProgressPeriodDefaultValue(),
		"Hooks status progress period in seconds. Set 0 to stop showing hooks status progress. Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value",
	)
}

func hooksStatusProgressPeriodDefaultValue() *int64 {
	defaultValue := statusProgressPeriodDefaultValue()

	v, err := getIntEnvVar("WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS")
	if err != nil {
		TerminateWithError(err.Error(), 1)
	}

	if v == nil {
		return defaultValue
	} else {
		return v
	}
}

func SetupImagesRepoOptions(cmdData *CmdData, cmd *cobra.Command) {
	// TODO: add the following options for images repo
	SetupInsecureRegistry(cmdData, cmd)
	SetupSkipTlsVerifyRegistry(cmdData, cmd)

	if cmdData.CommonRepoData == nil {
		SetupCommonRepoData(cmdData, cmd)
	}

	setupImagesRepo(cmdData, cmd)
	setupImagesRepoMode(cmdData, cmd)

	cmdData.ImagesRepoData = &RepoData{DesignationStorageName: "images repo"}

	SetupImplementationForRepoData(cmdData.ImagesRepoData, cmd, "images-repo-implementation", []string{"WERF_IMAGES_REPO_IMPLEMENTATION", "WERF_REPO_IMPLEMENTATION"})
	SetupDockerHubUsernameForRepoData(cmdData.ImagesRepoData, cmd, "images-repo-docker-hub-username", []string{"WERF_IMAGES_REPO_DOCKER_HUB_USERNAME", "WERF_REPO_DOCKER_HUB_USERNAME"})
	SetupDockerHubPasswordForRepoData(cmdData.ImagesRepoData, cmd, "images-repo-docker-hub-password", []string{"WERF_IMAGES_REPO_DOCKER_HUB_PASSWORD", "WERF_REPO_DOCKER_HUB_PASSWORD"})
	SetupDockerHubTokenForRepoData(cmdData.ImagesRepoData, cmd, "images-repo-docker-hub-token", []string{"WERF_IMAGES_REPO_DOCKER_HUB_TOKEN", "WERF_REPO_DOCKER_HUB_TOKEN"})
	SetupGithubTokenForRepoData(cmdData.ImagesRepoData, cmd, "images-repo-github-token", []string{"WERF_IMAGES_REPO_GITHUB_TOKEN", "WERF_REPO_GITHUB_TOKEN"})
	SetupHarborUsernameForRepoData(cmdData.ImagesRepoData, cmd, "images-repo-harbor-username", []string{"WERF_IMAGES_REPO_HARBOR_USERNAME", "WERF_REPO_HARBOR_USERNAME"})
	SetupHarborPasswordForRepoData(cmdData.ImagesRepoData, cmd, "images-repo-harbor-password", []string{"WERF_IMAGES_REPO_HARBOR_PASSWORD", "WERF_REPO_HARBOR_PASSWORD"})
	SetupQuayTokenForRepoData(cmdData.ImagesRepoData, cmd, "images-repo-quay-token", []string{"WERF_IMAGES_REPO_QUAY_TOKEN", "WERF_REPO_QUAY_TOKEN"})
}

func setupImagesRepo(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ImagesRepo = new(string)
	cmd.Flags().StringVarP(cmdData.ImagesRepo, "images-repo", "i", os.Getenv("WERF_IMAGES_REPO"), "Docker Repo to store images (default $WERF_IMAGES_REPO)")
}

func setupImagesRepoMode(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ImagesRepoMode = new(string)

	defaultValue := os.Getenv("WERF_IMAGES_REPO_MODE")
	if defaultValue == "" {
		defaultValue = "auto"
	}

	cmd.Flags().StringVarP(cmdData.ImagesRepoMode, "images-repo-mode", "", defaultValue, fmt.Sprintf(`Define how to store in images repo: %s or %s.
Default $WERF_IMAGES_REPO_MODE or auto mode`, docker_registry.MultirepoRepoMode, docker_registry.MonorepoRepoMode))
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
	setupLogQuiet(cmdData, cmd)
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
		cmd.Flags().BoolVarP(
			cmdData.LogDebug,
			alias,
			"",
			defaultValue,
			fmt.Sprintf("Enable debug (default $%s).", env),
		)
	}

	if err := cmd.Flags().MarkHidden("debug"); err != nil {
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

	cmd.Flags().StringVarP(cmdData.LogColorMode, "log-color-mode", "", defaultValue, `Set log color mode.
Supported on, off and auto (based on the stdout’s file descriptor referring to a terminal) modes.
Default $WERF_LOG_COLOR_MODE or auto mode.`)
}

func setupLogQuiet(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogQuiet = new(bool)

	var defaultValue bool
	for _, envName := range []string{
		"WERF_LOG_QUIET",
		"WERF_QUIET",
	} {
		if os.Getenv(envName) != "" {
			defaultValue = GetBoolEnvironmentDefaultFalse(envName)
			break
		}
	}

	for alias, env := range map[string]string{
		"log-quiet": "WERF_LOG_QUIET",
		"quiet":     "WERF_QUIET",
	} {
		cmd.Flags().BoolVarP(
			cmdData.LogQuiet,
			alias,
			"",
			defaultValue,
			fmt.Sprintf(`Disable explanatory output (default $%s).`, env),
		)
	}

	if err := cmd.Flags().MarkHidden("quiet"); err != nil {
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
		cmd.Flags().BoolVarP(
			cmdData.LogVerbose,
			alias,
			"",
			defaultValue,
			fmt.Sprintf(`Enable verbose output (default $%s).`, env),
		)
	}

	if err := cmd.Flags().MarkHidden("verbose"); err != nil {
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

	cmd.Flags().BoolVarP(cmdData.LogPretty, "log-pretty", "", defaultValue, `Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or true).`)
}

func setupTerminalWidth(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogTerminalWidth = new(int64)
	cmd.Flags().Int64VarP(cmdData.LogTerminalWidth, "log-terminal-width", "", -1, fmt.Sprintf(`Set log terminal width.
Defaults to:
* $WERF_LOG_TERMINAL_WIDTH
* interactive terminal width or %d`, logboek.DefaultWidth))
}

func SetupSet(cmdData *CmdData, cmd *cobra.Command) {
	set := predefinedValuesByEnvNamePrefix("WERF_SET", "WERF_SET_STRING")

	cmdData.Set = &set
	cmd.Flags().StringArrayVarP(cmdData.Set, "set", "", set, `Set helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2).
Also, can be defined with $WERF_SET* (e.g. $WERF_SET_1=key1=val1, $WERF_SET_2=key2=val2)`)
}

func SetupSetString(cmdData *CmdData, cmd *cobra.Command) {
	setString := predefinedValuesByEnvNamePrefix("WERF_SET_STRING")

	cmdData.SetString = &setString
	cmd.Flags().StringArrayVarP(cmdData.SetString, "set-string", "", setString, `Set STRING helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2).
Also, can be defined with $WERF_SET_STRING* (e.g. $WERF_SET_STRING_1=key1=val1, $WERF_SET_STRING_2=key2=val2)`)
}

func SetupValues(cmdData *CmdData, cmd *cobra.Command) {
	values := predefinedValuesByEnvNamePrefix("WERF_VALUES")

	cmdData.Values = &values
	cmd.Flags().StringArrayVarP(cmdData.Values, "values", "", values, `Specify helm values in a YAML file or a URL (can specify multiple).
Also, can be defined with $WERF_VALUES* (e.g. $WERF_VALUES_ENV=.helm/values_test.yaml, $WERF_VALUES_DB=.helm/values_db.yaml)`)
}

func SetupSecretValues(cmdData *CmdData, cmd *cobra.Command) {
	secretValues := predefinedValuesByEnvNamePrefix("WERF_SECRET_VALUES")

	cmdData.SecretValues = &secretValues
	cmd.Flags().StringArrayVarP(cmdData.SecretValues, "secret-values", "", secretValues, `Specify helm secret values in a YAML file (can specify multiple).
Also, can be defined with $WERF_SECRET_VALUES* (e.g. $WERF_SECRET_VALUES_ENV=.helm/secret_values_test.yaml, $WERF_SECRET_VALUES=.helm/secret_values_db.yaml)`)
}

func SetupIgnoreSecretKey(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.IgnoreSecretKey = new(bool)
	cmd.Flags().BoolVarP(cmdData.IgnoreSecretKey, "ignore-secret-key", "", GetBoolEnvironmentDefaultFalse("WERF_IGNORE_SECRET_KEY"), "Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)")
}

func SetupLogProjectDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogProjectDir = new(bool)
	cmd.Flags().BoolVarP(cmdData.LogProjectDir, "log-project-dir", "", GetBoolEnvironmentDefaultFalse("WERF_LOG_PROJECT_DIR"), `Print current project directory path (default $WERF_LOG_PROJECT_DIR)`)
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

func allStagesNames() []string {
	var stageNames []string
	for _, stageName := range stage.AllStages {
		stageNames = append(stageNames, string(stageName))
	}

	return stageNames
}

func SetupThreeWayMergeMode(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ThreeWayMergeMode = new(string)

	modeEnvironmentValue := os.Getenv("WERF_THREE_WAY_MERGE_MODE")

	defaultValue := ""
	if modeEnvironmentValue != "" {
		defaultValue = modeEnvironmentValue
	}

	cmd.Flags().StringVarP(cmdData.ThreeWayMergeMode, "three-way-merge-mode", "", defaultValue, `Set three way merge mode for release.
Supported 'enabled', 'disabled' and 'onlyNewReleases', see docs for more info https://werf.io/documentation/reference/deploy_process/experimental_three_way_merge.html`)
}

func GetThreeWayMergeMode(threeWayMergeModeParam string) (helm.ThreeWayMergeModeType, error) {
	switch threeWayMergeModeParam {
	case "enabled", "disabled", "onlyNewReleases", "":
		return helm.ThreeWayMergeModeType(threeWayMergeModeParam), nil
	}

	return "", fmt.Errorf("bad three-way-merge-mode '%s': enabled, disabled or  onlyNewReleases modes can be specified", threeWayMergeModeParam)
}

func GetBoolEnvironmentDefaultFalse(environmentName string) bool {
	switch os.Getenv(environmentName) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

func ConvertInt32Value(v string) (int32, error) {
	res, err := ConvertIntValue(v, 32)
	if err != nil {
		return 0, err
	}
	return int32(res), nil
}

func ConvertIntValue(v string, bitSize int) (int64, error) {
	vInt, err := strconv.ParseInt(v, 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf("bad integer value '%s': %s", v, err)
	}
	return vInt, nil
}

func getInt64EnvVar(varName string) (*int64, error) {
	if v := os.Getenv(varName); v != "" {
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value '%s': %s", varName, v, err)
		}

		res := new(int64)
		*res = vInt

		return res, nil
	}

	return nil, nil
}

func getIntEnvVar(varName string) (*int64, error) {
	if v := os.Getenv(varName); v != "" {
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value '%s': %s", varName, v, err)
		}

		res := new(int64)
		*res = vInt

		return res, nil
	}

	return nil, nil
}

func GetGitTagStrategyLimit(cmdData *CmdData) (int64, error) {
	v, err := getInt64EnvVar("WERF_GIT_TAG_STRATEGY_LIMIT")
	if err != nil {
		return 0, err
	}
	if v != nil {
		return *v, nil
	}
	return *cmdData.GitTagStrategyLimit, nil
}

func GetGitTagStrategyExpiryDays(cmdData *CmdData) (int64, error) {
	v, err := getInt64EnvVar("WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS")
	if err != nil {
		return 0, err
	}
	if v != nil {
		return *v, nil
	}
	return *cmdData.GitTagStrategyExpiryDays, nil
}

func GetGitCommitStrategyLimit(cmdData *CmdData) (int64, error) {
	v, err := getInt64EnvVar("WERF_GIT_COMMIT_STRATEGY_LIMIT")
	if err != nil {
		return 0, err
	}
	if v != nil {
		return *v, nil
	}
	return *cmdData.GitCommitStrategyLimit, nil
}

func GetGitCommitStrategyExpiryDays(cmdData *CmdData) (int64, error) {
	v, err := getInt64EnvVar("WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS")
	if err != nil {
		return 0, err
	}
	if v != nil {
		return *v, nil
	}
	return *cmdData.GitCommitStrategyExpiryDays, nil
}

func GetStagesSignatureStrategyLimit(cmdData *CmdData) (int64, error) {
	v, err := getInt64EnvVar("WERF_STAGES_SIGNATURE_STRATEGY_LIMIT")
	if err != nil {
		return 0, err
	}
	if v != nil {
		return *v, nil
	}
	return *cmdData.StagesSignatureStrategyLimit, nil
}

func GetStagesSignatureStrategyExpiryDays(cmdData *CmdData) (int64, error) {
	v, err := getInt64EnvVar("WERF_STAGES_SIGNATURE_STRATEGY_EXPIRY_DAYS")
	if err != nil {
		return 0, err
	}
	if v != nil {
		return *v, nil
	}
	return *cmdData.StagesSignatureStrategyExpiryDays, nil
}

func GetImagesCleanupPolicies(cmdData *CmdData) (cleanup.ImagesCleanupPolicies, error) {
	tagLimit, err := GetGitTagStrategyLimit(cmdData)
	if err != nil {
		return cleanup.ImagesCleanupPolicies{}, err
	}

	tagDays, err := GetGitTagStrategyExpiryDays(cmdData)
	if err != nil {
		return cleanup.ImagesCleanupPolicies{}, err
	}

	commitLimit, err := GetGitCommitStrategyLimit(cmdData)
	if err != nil {
		return cleanup.ImagesCleanupPolicies{}, err
	}

	commitDays, err := GetGitCommitStrategyExpiryDays(cmdData)
	if err != nil {
		return cleanup.ImagesCleanupPolicies{}, err
	}

	stagesSignatureLimit, err := GetStagesSignatureStrategyLimit(cmdData)
	if err != nil {
		return cleanup.ImagesCleanupPolicies{}, err
	}

	stagesSignatureDays, err := GetStagesSignatureStrategyExpiryDays(cmdData)
	if err != nil {
		return cleanup.ImagesCleanupPolicies{}, err
	}

	res := cleanup.ImagesCleanupPolicies{}

	if tagLimit >= 0 {
		res.GitTagStrategyHasLimit = true
		res.GitTagStrategyLimit = tagLimit
	}
	if tagDays >= 0 {
		res.GitTagStrategyHasExpiryPeriod = true
		res.GitTagStrategyExpiryPeriod = time.Hour * 24 * time.Duration(tagDays)
	}
	if commitLimit >= 0 {
		res.GitCommitStrategyHasLimit = true
		res.GitCommitStrategyLimit = commitLimit
	}
	if commitDays >= 0 {
		res.GitCommitStrategyHasExpiryPeriod = true
		res.GitCommitStrategyExpiryPeriod = time.Hour * 24 * time.Duration(commitDays)
	}
	if stagesSignatureLimit >= 0 {
		res.StagesSignatureStrategyHasLimit = true
		res.StagesSignatureStrategyLimit = stagesSignatureLimit
	}
	if stagesSignatureDays >= 0 {
		res.StagesSignatureStrategyHasExpiryPeriod = true
		res.StagesSignatureStrategyExpiryPeriod = time.Hour * 24 * time.Duration(stagesSignatureDays)
	}

	return res, nil
}

func GetStagesStorageAddress(cmdData *CmdData) (string, error) {
	if *cmdData.StagesStorage == "" {
		return "", fmt.Errorf("--stages-storage=ADDRESS param required")
	}
	return *cmdData.StagesStorage, nil
}

func GetOptionalStagesStorageAddress(cmdData *CmdData) string {
	return *cmdData.StagesStorage
}

func GetImagesRepoWithOptionalStubRepoAddress(projectName string, cmdData *CmdData) (storage.ImagesRepo, error) {
	return getImagesRepo(projectName, cmdData, true)
}

func GetImagesRepo(projectName string, cmdData *CmdData) (storage.ImagesRepo, error) {
	return getImagesRepo(projectName, cmdData, false)
}

func getImagesRepo(projectName string, cmdData *CmdData, optionalStubRepoAddress bool) (storage.ImagesRepo, error) {
	var imagesRepoAddress string
	var err error
	if optionalStubRepoAddress {
		imagesRepoAddress, err = getImagesRepoAddressOrStub(projectName, cmdData)
		if err != nil {
			return nil, err
		}
	} else {
		imagesRepoAddress, err = getImagesRepoAddress(projectName, cmdData)
		if err != nil {
			return nil, err
		}
	}

	imagesRepoMode, err := getImagesRepoMode(cmdData)
	if err != nil {
		return nil, err
	}

	repoData := MergeRepoData(cmdData.ImagesRepoData, cmdData.CommonRepoData)

	if err := ValidateRepoImplementation(*repoData.Implementation); err != nil {
		return nil, err
	}

	return storage.NewImagesRepo(
		projectName,
		imagesRepoAddress,
		imagesRepoMode,
		storage.ImagesRepoOptions{
			DockerImagesRepoOptions: storage.DockerImagesRepoOptions{
				Implementation: *repoData.Implementation,
				DockerRegistryOptions: docker_registry.DockerRegistryOptions{
					InsecureRegistry:      *cmdData.InsecureRegistry,
					SkipTlsVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
					DockerHubUsername:     *repoData.DockerHubUsername,
					DockerHubPassword:     *repoData.DockerHubPassword,
					GitHubToken:           *repoData.GitHubToken,
					HarborUsername:        *repoData.HarborUsername,
					HarborPassword:        *repoData.HarborPassword,
					QuayToken:             *repoData.QuayToken,
				},
			},
		},
	)
}

func GetStagesStorage(containerRuntime container_runtime.ContainerRuntime, cmdData *CmdData) (storage.StagesStorage, error) {
	stagesStorageAddress, err := GetStagesStorageAddress(cmdData)
	if err != nil {
		return nil, err
	}

	repoData := MergeRepoData(cmdData.StagesStorageRepoData, cmdData.CommonRepoData)

	if err := ValidateRepoImplementation(*repoData.Implementation); err != nil {
		return nil, err
	}

	return storage.NewStagesStorage(
		stagesStorageAddress,
		containerRuntime,
		storage.StagesStorageOptions{
			RepoStagesStorageOptions: storage.RepoStagesStorageOptions{
				Implementation: *repoData.Implementation,
				DockerRegistryOptions: docker_registry.DockerRegistryOptions{
					InsecureRegistry:      *cmdData.InsecureRegistry,
					SkipTlsVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
					DockerHubUsername:     *repoData.DockerHubUsername,
					DockerHubPassword:     *repoData.DockerHubPassword,
					DockerHubToken:        *repoData.DockerHubToken,
					GitHubToken:           *repoData.GitHubToken,
					HarborUsername:        *repoData.HarborUsername,
					HarborPassword:        *repoData.HarborPassword,
					QuayToken:             *repoData.QuayToken,
				},
			},
		},
	)
}

func GetSynchronization(cmdData *CmdData, stagesStorageAddress string) (string, error) {
	if *cmdData.Synchronization == "" {
		if stagesStorageAddress == storage.LocalStorageAddress {
			return storage.LocalStorageAddress, nil
		} else {
			return storage.DefaultKubernetesStorageAddress, nil
		}
	} else {
		if *cmdData.Synchronization == storage.LocalStorageAddress {
			return *cmdData.Synchronization, nil
		} else if strings.HasPrefix(*cmdData.Synchronization, "kubernetes://") {
			return *cmdData.Synchronization, nil
		} else {
			return "", fmt.Errorf("only --synchronization=%s or --synchronization=kubernetes://NAMESPACE is supported, got %q", storage.LocalStorageAddress, *cmdData.Synchronization)
		}
	}
}

func getImagesRepoAddressOrStub(projectName string, cmdData *CmdData) (string, error) {
	imagesRepoAddress, err := GetOptionalImagesRepoAddress(projectName, cmdData)
	if err != nil {
		return "", err
	}

	if imagesRepoAddress == "" {
		return StubImagesRepoAddress, nil
	}

	return imagesRepoAddress, nil
}

func getImagesRepoAddress(projectName string, cmdData *CmdData) (string, error) {
	if *cmdData.ImagesRepo == "" {
		return "", fmt.Errorf("--images-repo REPO param required")
	}
	return GetOptionalImagesRepoAddress(projectName, cmdData)
}

func GetOptionalImagesRepoAddress(projectName string, cmdData *CmdData) (string, error) {
	repoOption := *cmdData.ImagesRepo

	if repoOption == ":minikube" {
		repoOption = fmt.Sprintf("werf-registry.kube-system.svc.cluster.local:5000/%s", projectName)
	}

	return repoOption, nil
}

func getImagesRepoMode(cmdData *CmdData) (string, error) {
	switch *cmdData.ImagesRepoMode {
	case docker_registry.MultirepoRepoMode, docker_registry.MonorepoRepoMode, "auto":
		return *cmdData.ImagesRepoMode, nil
	default:
		return "", fmt.Errorf("bad --images-repo-mode '%s': only %s, %s or auto supported", *cmdData.ImagesRepoMode, docker_registry.MultirepoRepoMode, docker_registry.MonorepoRepoMode)
	}
}

func GetOptionalWerfConfig(projectDir string, cmdData *CmdData, logRenderedFilePath bool) (*config.WerfConfig, error) {
	werfConfigPath, err := GetWerfConfigPath(projectDir, cmdData, false)
	if err != nil {
		return nil, err
	}

	if werfConfigPath != "" {
		werfConfigTemplatesDir := GetWerfConfigTemplatesDir(projectDir, cmdData)
		return config.GetWerfConfig(werfConfigPath, werfConfigTemplatesDir, logRenderedFilePath)
	}

	return nil, nil
}

func GetRequiredWerfConfig(projectDir string, cmdData *CmdData, logRenderedFilePath bool) (*config.WerfConfig, error) {
	werfConfigPath, err := GetWerfConfigPath(projectDir, cmdData, true)
	if err != nil {
		return nil, err
	}

	werfConfigTemplatesDir := GetWerfConfigTemplatesDir(projectDir, cmdData)

	return config.GetWerfConfig(werfConfigPath, werfConfigTemplatesDir, logRenderedFilePath)
}

func GetWerfConfigPath(projectDir string, cmdData *CmdData, required bool) (string, error) {
	var configPathToCheck []string

	customConfigPath := *cmdData.ConfigPath
	if customConfigPath != "" {
		configPathToCheck = append(configPathToCheck, customConfigPath)
	} else {
		for _, werfDefaultConfigName := range []string{"werf.yml", "werf.yaml"} {
			configPathToCheck = append(configPathToCheck, filepath.Join(projectDir, werfDefaultConfigName))
		}
	}

	for _, werfConfigPath := range configPathToCheck {
		exist, err := util.FileExists(werfConfigPath)
		if err != nil {
			return "", err
		}

		if exist {
			return werfConfigPath, nil
		}
	}

	if required {
		if customConfigPath != "" {
			return "", fmt.Errorf("configration file %s not found", customConfigPath)
		} else {
			return "", fmt.Errorf("configration file werf.yaml not found")
		}
	}

	return "", nil
}

func GetWerfConfigTemplatesDir(projectDir string, cmdData *CmdData) string {
	customConfigTemplatesDir := *cmdData.ConfigTemplatesDir
	if customConfigTemplatesDir != "" {
		return customConfigTemplatesDir
	} else {
		return filepath.Join(projectDir, ".werf")
	}
}

func GetProjectDir(cmdData *CmdData) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if *cmdData.Dir != "" {
		if filepath.IsAbs(*cmdData.Dir) {
			return *cmdData.Dir, nil
		} else {
			return filepath.Clean(filepath.Join(currentDir, *cmdData.Dir)), nil
		}
	}

	return currentDir, nil
}

func GetHelmChartDir(projectDir string, cmdData *CmdData) (string, error) {
	var helmChartDir string

	customHelmChartDir := *cmdData.HelmChartDir
	if customHelmChartDir != "" {
		helmChartDir = customHelmChartDir
	} else {
		helmChartDir = filepath.Join(projectDir, ".helm")
	}

	exist, err := util.FileExists(helmChartDir)
	if err != nil {
		return "", err
	}

	if !exist {
		return "", fmt.Errorf("helm chart dir %s not found", helmChartDir)
	}

	return helmChartDir, nil
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

	if *cmdData.LogQuiet {
		logging.EnableLogQuiet()
	} else if *cmdData.LogDebug {
		logging.EnableLogDebug()
	} else if *cmdData.LogVerbose {
		logging.EnableLogVerbose()
	}

	if !*cmdData.LogPretty {
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
		logging.EnableLogColor()
	case "off":
		logging.DisableLogColor()
	default:
		return fmt.Errorf("bad log color mode '%s': on, off and auto modes are supported", logColorMode)
	}

	return nil
}

func ProcessLogTerminalWidth(cmdData *CmdData) error {
	value := *cmdData.LogTerminalWidth

	if value != -1 {
		if value < 0 {
			return fmt.Errorf("--log-terminal-width parameter (%d) can not be negative", value)
		}

		logging.SetWidth(int(value))
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

		logging.SetWidth(int(*pInt64))
	}

	return nil
}

func DockerRegistryInit(cmdData *CmdData) error {
	return docker_registry.Init(*cmdData.InsecureRegistry, *cmdData.SkipTlsVerifyRegistry)
}

func ValidateRepoImplementation(implementation string) error {
	supportedValues := docker_registry.ImplementationList()
	supportedValues = append(supportedValues, "auto", "")

	for _, supportedImplementation := range supportedValues {
		if supportedImplementation == implementation {
			return nil
		}
	}

	return fmt.Errorf("specified docker registry implementation '%s' is not supported", implementation)
}

func ValidateMinimumNArgs(minArgs int, args []string, cmd *cobra.Command) error {
	if len(args) < minArgs {
		PrintHelp(cmd)
		return fmt.Errorf("requires at least %d arg(s), received %d", minArgs, len(args))
	}

	return nil
}

func ValidateMaximumNArgs(maxArgs int, args []string, cmd *cobra.Command) error {
	if len(args) > maxArgs {
		PrintHelp(cmd)
		return fmt.Errorf("accepts at most %d arg(s), received %d", maxArgs, len(args))
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

	logboek.Default.LogFHighlight("Running time %0.2f seconds\n", time.Now().Sub(t).Seconds())

	return err
}

func LogVersion() {
	logboek.LogF("Version: %s\n", werf.Version)
}

func TerminateWithError(errMsg string, exitCode int) {
	msg := fmt.Sprintf("Error: %s", errMsg)
	msg = strings.TrimSuffix(msg, "\n")

	logboek.LogErrorLn(msg)
	os.Exit(exitCode)
}

func SetupVirtualMerge(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.VirtualMerge = new(bool)
	cmd.Flags().BoolVarP(cmdData.VirtualMerge, "virtual-merge", "", GetBoolEnvironmentDefaultFalse("WERF_VIRTUAL_MERGE"), "Enable virtual/ephemeral merge commit mode when building current application state ($WERF_VIRTUAL_MERGE by default)")
}

func SetupVirtualMergeFromCommit(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.VirtualMergeFromCommit = new(string)
	cmd.Flags().StringVarP(cmdData.VirtualMergeFromCommit, "virtual-merge-from-commit", "", os.Getenv("WERF_VIRTUAL_MERGE_FROM_COMMIT"), "Commit hash for virtual/ephemeral merge commit with new changes introduced in the pull request ($WERF_VIRTUAL_MERGE_FROM_COMMIT by default)")
}

func SetupVirtualMergeIntoCommit(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.VirtualMergeIntoCommit = new(string)
	cmd.Flags().StringVarP(cmdData.VirtualMergeIntoCommit, "virtual-merge-into-commit", "", os.Getenv("WERF_VIRTUAL_MERGE_INTO_COMMIT"), "Commit hash for virtual/ephemeral merge commit which is base for changes introduced in the pull request ($WERF_VIRTUAL_MERGE_INTO_COMMIT by default)")
}
