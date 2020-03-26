package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/build/stage"
	cleanup "github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/storage"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

type CmdData struct {
	ProjectName *string
	Dir         *string
	TmpDir      *string
	HomeDir     *string
	SSHKeys     *[]string

	TagCustom            *[]string
	TagGitBranch         *string
	TagGitTag            *string
	TagGitCommit         *string
	TagByStagesSignature *bool

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

	StagesStorage                      *string
	StagesStorageRepoImplementation    *string
	StagesStorageRepoDockerHubToken    *string
	StagesStorageRepoDockerHubUsername *string
	StagesStorageRepoDockerHubPassword *string
	StagesStorageRepoGitHubToken       *string
	StagesStorageRepoHarborUsername    *string
	StagesStorageRepoHarborPassword    *string

	Synchronization *string

	ImagesRepo                  *string
	ImagesRepoMode              *string
	ImagesRepoImplementation    *string
	ImagesRepoDockerHubToken    *string
	ImagesRepoDockerHubUsername *string
	ImagesRepoDockerHubPassword *string
	ImagesRepoGitHubToken       *string
	ImagesRepoHarborUsername    *string
	ImagesRepoHarborPassword    *string

	RepoImplementation    *string
	RepoDockerHubToken    *string
	RepoDockerHubUsername *string
	RepoDockerHubPassword *string
	RepoGitHubToken       *string
	RepoHarborUsername    *string
	RepoHarborPassword    *string

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
}

const (
	CleaningCommandsForceOptionDescription = "Remove containers that are based on deleting werf docker images"

	StubImagesRepoAddress = "stub/repository"
)

func GetLongCommandDescription(text string) string {
	return logboek.FitText(text, logboek.FitTextOptions{MaxWidth: 100})
}

func SetupProjectName(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ProjectName = new(string)
	cmd.Flags().StringVarP(cmdData.ProjectName, "project-name", "N", os.Getenv("WERF_PROJECT_NAME"), "Use specified project name (default $WERF_PROJECT_NAME)")
}

func SetupDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Dir = new(string)
	cmd.Flags().StringVarP(cmdData.Dir, "dir", "", "", "Change to the specified directory to find werf.yaml config")
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
	cmdData.SSHKeys = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SSHKeys, "ssh-key", "", []string{}, "Use only specific ssh keys (Defaults to system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see https://werf.io/documentation/reference/toolbox/ssh.html).\nOption can be specified multiple times to use multiple keys")
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
	var tagCustom []string
	for _, keyValue := range os.Environ() {
		parts := strings.SplitN(keyValue, "=", 2)
		if strings.HasPrefix(parts[0], "WERF_TAG_CUSTOM") {
			tagCustom = append(tagCustom, parts[1])
		}
	}

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

func SetupEnvironment(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Environment = new(string)
	cmd.Flags().StringVarP(cmdData.Environment, "env", "", os.Getenv("WERF_ENV"), "Use specified environment (default $WERF_ENV)")
}

func SetupRelease(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Release = new(string)
	cmd.Flags().StringVarP(cmdData.Release, "release", "", "", "Use specified Helm release name (default [[ project ]]-[[ env ]] template or deploy.helmRelease custom template from werf.yaml)")
}

func SetupNamespace(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Namespace = new(string)
	cmd.Flags().StringVarP(cmdData.Namespace, "namespace", "", "", "Use specified Kubernetes namespace (default [[ project ]]-[[ env ]] template or deploy.namespace custom template from werf.yaml)")
}

func SetupAddAnnotations(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.AddAnnotations = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.AddAnnotations, "add-annotation", "", []string{}, `Add annotation to deploying resources (can specify multiple).
Format: annoName=annoValue.
Also can be specified in $WERF_ADD_ANNOTATION* (e.g. $WERF_ADD_ANNOTATION_1=annoName1=annoValue1", $WERF_ADD_ANNOTATION_2=annoName2=annoValue2")`)
}

func SetupAddLabels(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.AddLabels = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.AddLabels, "add-label", "", []string{}, `Add label to deploying resources (can specify multiple).
Format: labelName=labelValue.
Also can be specified in $WERF_ADD_LABEL* (e.g. $WERF_ADD_LABEL_1=labelName1=labelValue1", $WERF_ADD_LABEL_2=labelName2=labelValue2")`)
}

func SetupKubeContext(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeContext = new(string)
	cmd.Flags().StringVarP(cmdData.KubeContext, "kube-context", "", os.Getenv("WERF_KUBE_CONTEXT"), "Kubernetes config context (default $WERF_KUBE_CONTEXT)")
}

func SetupKubeConfig(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeConfig = new(string)
	cmd.Flags().StringVarP(cmdData.KubeConfig, "kube-config", "", "", "Kubernetes config file path")
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

func setupRepoImplementation(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.RepoImplementation != nil {
		return
	}

	usage := fmt.Sprintf(`Choose default repo implementation for images repo and stages storage repo.
The following docker registry implementations are supported: %s.
Default %s or auto mode (detect implementation by a registry).`,
		strings.Join(docker_registry.ImplementationList(), ", "),
		"$WERF_REPO_IMPLEMENTATION",
	)

	cmdData.RepoImplementation = new(string)
	cmd.Flags().StringVarP(
		cmdData.RepoImplementation,
		"repo-implementation",
		"",
		os.Getenv("WERF_REPO_IMPLEMENTATION"),
		usage,
	)
}

func setupRepoDockerHubToken(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.RepoDockerHubToken != nil {
		return
	}

	cmdData.RepoDockerHubToken = new(string)
	cmd.Flags().StringVarP(
		cmdData.RepoDockerHubToken,
		"repo-docker-hub-token",
		"",
		os.Getenv("WERF_REPO_DOCKER_HUB_TOKEN"),
		"Default Docker Hub token for stages storage repo and images repo implementations (default $WERF_REPO_DOCKER_HUB_TOKEN).",
	)
}

func setupRepoDockerHubUsername(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.RepoDockerHubUsername != nil {
		return
	}

	cmdData.RepoDockerHubUsername = new(string)
	cmd.Flags().StringVarP(
		cmdData.RepoDockerHubUsername,
		"repo-docker-hub-username",
		"",
		os.Getenv("WERF_REPO_DOCKER_HUB_USERNAME"),
		"Default Docker Hub username for stages storage repo and images repo implementations (default $WERF_REPO_DOCKER_HUB_USERNAME).",
	)
}

func setupRepoDockerHubPassword(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.RepoDockerHubPassword != nil {
		return
	}

	cmdData.RepoDockerHubPassword = new(string)
	cmd.Flags().StringVarP(
		cmdData.RepoDockerHubPassword,
		"repo-docker-hub-password",
		"",
		os.Getenv("WERF_REPO_DOCKER_HUB_PASSWORD"),
		"Default Docker Hub password for stages storage repo and images repo implementations (default $WERF_REPO_DOCKER_HUB_PASSWORD).",
	)
}

func setupRepoGitHubToken(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.RepoGitHubToken != nil {
		return
	}

	cmdData.RepoGitHubToken = new(string)
	cmd.Flags().StringVarP(
		cmdData.RepoGitHubToken,
		"repo-github-token",
		"",
		os.Getenv("WERF_REPO_GITHUB_TOKEN"),
		fmt.Sprintf("Default GitHub token for stages storage repo and images repo implementations (default $WERF_REPO_GITHUB_TOKEN)."),
	)
}

func setupRepoHarborUsername(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.RepoHarborUsername != nil {
		return
	}

	cmdData.RepoHarborUsername = new(string)
	cmd.Flags().StringVarP(
		cmdData.RepoHarborUsername,
		"repo-harbor-username",
		"",
		os.Getenv("WERF_REPO_HARBOR_USERNAME"),
		"Default harbor username for stages storage repo and images repo implementations (default $WERF_REPO_HARBOR_USERNAME).",
	)

	_ = cmd.Flags().MarkHidden("repo-harbor-username")
}

func setupRepoHarborPassword(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.RepoHarborPassword != nil {
		return
	}

	cmdData.RepoHarborPassword = new(string)
	cmd.Flags().StringVarP(
		cmdData.RepoHarborPassword,
		"repo-harbor-password",
		"",
		os.Getenv("WERF_REPO_HARBOR_PASSWORD"),
		"Default harbor password for stages storage repo and images repo implementations (default $WERF_REPO_HARBOR_PASSWORD).",
	)

	_ = cmd.Flags().MarkHidden("repo-harbor-password")
}

func SetupStagesStorageOptions(cmdData *CmdData, cmd *cobra.Command) {
	setupStagesStorage(cmdData, cmd)

	setupRepoImplementation(cmdData, cmd)
	setupStagesStorageRepoImplementation(cmdData, cmd)

	setupRepoDockerHubToken(cmdData, cmd)
	setupRepoDockerHubUsername(cmdData, cmd)
	setupRepoDockerHubPassword(cmdData, cmd)
	setupRepoGitHubToken(cmdData, cmd)
	setupRepoHarborUsername(cmdData, cmd)
	setupRepoHarborPassword(cmdData, cmd)
	setupStagesStorageRepoDockerHubToken(cmdData, cmd)
	setupStagesStorageRepoDockerHubUsername(cmdData, cmd)
	setupStagesStorageRepoDockerHubPassword(cmdData, cmd)
	setupStagesStorageRepoGitHubToken(cmdData, cmd)
	setupStagesStorageRepoHarborUsername(cmdData, cmd)
	setupStagesStorageRepoHarborPassword(cmdData, cmd)

	// TODO: add the following options for stages storage
	SetupInsecureRegistry(cmdData, cmd)
	SetupSkipTlsVerifyRegistry(cmdData, cmd)
}

func setupStagesStorageRepoImplementation(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_STAGES_STORAGE_REPO_IMPLEMENTATION"),
		os.Getenv("WERF_REPO_IMPLEMENTATION"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf(`Choose stages storage repo implementation.
The following  docker registry implementations are supported: %s.
Default $WERF_STAGES_STORAGE_REPO_IMPLEMENTATION, $WERF_REPO_IMPLEMENTATION or auto mode (detect implementation by a registry).`,
		strings.Join(docker_registry.ImplementationList(), ", "),
	)

	cmdData.StagesStorageRepoImplementation = new(string)
	cmd.Flags().StringVarP(
		cmdData.StagesStorageRepoImplementation,
		"stages-storage-repo-implementation",
		"",
		defaultValue,
		usage,
	)
}

func setupStagesStorageRepoDockerHubToken(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_STAGES_STORAGE_REPO_DOCKER_HUB_TOKEN"),
		os.Getenv("WERF_REPO_DOCKER_HUB_TOKEN"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("Docker Hub token for stages storage repo implementation (default $WERF_STAGES_STORAGE_REPO_DOCKER_HUB_TOKEN or $WERF_REPO_DOCKER_HUB_TOKEN).")

	cmdData.StagesStorageRepoDockerHubUsername = new(string)
	cmd.Flags().StringVarP(
		cmdData.StagesStorageRepoDockerHubUsername,
		"stages-storage-repo-docker-hub-token",
		"",
		defaultValue,
		usage,
	)
}

func setupStagesStorageRepoDockerHubUsername(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_STAGES_STORAGE_REPO_DOCKER_HUB_USERNAME"),
		os.Getenv("WERF_REPO_DOCKER_HUB_USERNAME"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("Docker Hub username for stages storage repo implementation (default $WERF_STAGES_STORAGE_REPO_DOCKER_HUB_USERNAME or $WERF_REPO_DOCKER_HUB_USERNAME).")

	cmdData.StagesStorageRepoDockerHubUsername = new(string)
	cmd.Flags().StringVarP(
		cmdData.StagesStorageRepoDockerHubUsername,
		"stages-storage-repo-docker-hub-username",
		"",
		defaultValue,
		usage,
	)
}

func setupStagesStorageRepoDockerHubPassword(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_STAGES_STORAGE_REPO_DOCKER_HUB_PASSWORD"),
		os.Getenv("WERF_REPO_DOCKER_HUB_PASSWORD"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("Docker Hub password for stages storage repo implementation (default $WERF_STAGES_STORAGE_REPO_DOCKER_HUB_PASSWORD or $WERF_REPO_DOCKER_HUB_PASSWORD).")

	cmdData.StagesStorageRepoDockerHubPassword = new(string)
	cmd.Flags().StringVarP(
		cmdData.StagesStorageRepoDockerHubPassword,
		"stages-storage-repo-docker-hub-password",
		"",
		defaultValue,
		usage,
	)
}

func setupStagesStorageRepoGitHubToken(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_STAGES_STORAGE_REPO_GITHUB_TOKEN"),
		os.Getenv("WERF_REPO_GITHUB_TOKEN"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("GitHub token for stages storage repo implementation (default $WERF_STAGES_STORAGE_REPO_GITHUB_TOKEN or $WERF_REPO_GITHUB_TOKEN).")

	cmdData.StagesStorageRepoGitHubToken = new(string)
	cmd.Flags().StringVarP(
		cmdData.StagesStorageRepoGitHubToken,
		"stages-storage-repo-github-token",
		"",
		defaultValue,
		usage,
	)
}

func setupStagesStorageRepoHarborUsername(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_STAGES_STORAGE_REPO_HARBOR_USERNAME"),
		os.Getenv("WERF_REPO_HARBOR_USERNAME"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("Harbor username for stages storage repo implementation (default $WERF_STAGES_STORAGE_REPO_HARBOR_USERNAME or $WERF_REPO_HARBOR_USERNAME).")

	cmdData.StagesStorageRepoHarborUsername = new(string)
	cmd.Flags().StringVarP(
		cmdData.StagesStorageRepoHarborUsername,
		"stages-storage-repo-harbor-username",
		"",
		defaultValue,
		usage,
	)

	_ = cmd.Flags().MarkHidden("stages-storage-repo-harbor-username")
}

func setupStagesStorageRepoHarborPassword(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_STAGES_STORAGE_REPO_HARBOR_PASSWORD"),
		os.Getenv("WERF_REPO_HARBOR_PASSWORD"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("Harbor password for stages storage repo implementation (default $WERF_STAGES_STORAGE_REPO_HARBOR_PASSWORD or $WERF_REPO_HARBOR_PASSWORD).")

	cmdData.StagesStorageRepoHarborPassword = new(string)
	cmd.Flags().StringVarP(
		cmdData.StagesStorageRepoHarborPassword,
		"stages-storage-repo-harbor-password",
		"",
		defaultValue,
		usage,
	)

	_ = cmd.Flags().MarkHidden("stages-storage-repo-harbor-password")
}

func setupStagesStorage(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.StagesStorage, "stages-storage", "s", os.Getenv("WERF_STAGES_STORAGE"), fmt.Sprintf("Docker Repo to store stages or %[1]s for non-distributed build (only %[1]s is supported for now; default $WERF_STAGES_STORAGE environment).\nMore info about stages: https://werf.io/documentation/reference/stages_and_images.html", storage.LocalStagesStorageAddress))
}

func SetupSynchronization(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Synchronization = new(string)

	defaultValue := os.Getenv("WERF_SYNCHRONIZATION")
	if defaultValue == "" {
		defaultValue = ":local"
	}

	cmd.Flags().StringVarP(cmdData.Synchronization, "synchronization", "", defaultValue, "Address of synchronizer for multiple werf processes to work with a single stages storage (default :local or $WERF_SYNCHRONIZATION if set). The same address should be specified for all werf processes that work with a single stages storage. :local address allows execution of werf processes from a single host only.")
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
	setupImagesRepo(cmdData, cmd)
	setupImagesRepoMode(cmdData, cmd)

	setupRepoImplementation(cmdData, cmd)
	setupImagesRepoImplementation(cmdData, cmd)

	setupRepoDockerHubToken(cmdData, cmd)
	setupRepoDockerHubUsername(cmdData, cmd)
	setupRepoDockerHubPassword(cmdData, cmd)
	setupRepoGitHubToken(cmdData, cmd)
	setupRepoHarborUsername(cmdData, cmd)
	setupRepoHarborPassword(cmdData, cmd)
	setupImagesRepoDockerHubToken(cmdData, cmd)
	setupImagesRepoDockerHubUsername(cmdData, cmd)
	setupImagesRepoDockerHubPassword(cmdData, cmd)
	setupImagesRepoGitHubToken(cmdData, cmd)
	setupImagesRepoHarborUsername(cmdData, cmd)
	setupImagesRepoHarborPassword(cmdData, cmd)

	// TODO: add the following options for images repo
	SetupInsecureRegistry(cmdData, cmd)
	SetupSkipTlsVerifyRegistry(cmdData, cmd)
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

func setupImagesRepoImplementation(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_IMAGES_REPO_IMPLEMENTATION"),
		os.Getenv("WERF_REPO_IMPLEMENTATION"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf(`Choose images repo implementation.
The following docker registry implementations are supported: %s.
Default $WERF_IMAGES_REPO_IMPLEMENTATION, $WERF_REPO_IMPLEMENTATION or auto mode (detect implementation by a registry).`,
		strings.Join(docker_registry.ImplementationList(), ", "),
	)

	cmdData.ImagesRepoImplementation = new(string)
	cmd.Flags().StringVarP(
		cmdData.ImagesRepoImplementation,
		"images-repo-implementation",
		"",
		defaultValue,
		usage,
	)
}

func setupImagesRepoDockerHubToken(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_IMAGES_REPO_DOCKER_HUB_TOKEN"),
		os.Getenv("WERF_REPO_DOCKER_HUB_TOKEN"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("Docker Hub token for images repo implementation (default $WERF_IMAGES_REPO_DOCKER_HUB_TOKEN or $WERF_REPO_DOCKER_HUB_TOKEN).")

	cmdData.ImagesRepoDockerHubToken = new(string)
	cmd.Flags().StringVarP(
		cmdData.ImagesRepoDockerHubToken,
		"images-repo-docker-hub-token",
		"",
		defaultValue,
		usage,
	)
}

func setupImagesRepoDockerHubUsername(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_IMAGES_REPO_DOCKER_HUB_USERNAME"),
		os.Getenv("WERF_REPO_DOCKER_HUB_USERNAME"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("Docker Hub username for images repo implementation (default $WERF_IMAGES_REPO_DOCKER_HUB_USERNAME or $WERF_REPO_DOCKER_HUB_USERNAME).")

	cmdData.ImagesRepoDockerHubUsername = new(string)
	cmd.Flags().StringVarP(
		cmdData.ImagesRepoDockerHubUsername,
		"images-repo-docker-hub-username",
		"",
		defaultValue,
		usage,
	)
}

func setupImagesRepoDockerHubPassword(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_IMAGES_REPO_DOCKER_HUB_PASSWORD"),
		os.Getenv("WERF_REPO_DOCKER_HUB_PASSWORD"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("Docker Hub password for images repo implementation (default $WERF_IMAGES_REPO_DOCKER_HUB_PASSWORD or $WERF_REPO_DOCKER_HUB_PASSWORD).")

	cmdData.ImagesRepoDockerHubPassword = new(string)
	cmd.Flags().StringVarP(
		cmdData.ImagesRepoDockerHubPassword,
		"images-repo-docker-hub-password",
		"",
		defaultValue,
		usage,
	)
}

func setupImagesRepoGitHubToken(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_IMAGES_REPO_GITHUB_TOKEN"),
		os.Getenv("WERF_REPO_GITHUB_TOKEN"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("GitHub token for images repo implementation (default $WERF_IMAGES_REPO_GITHUB_TOKEN or $WERF_REPO_GITHUB_TOKEN).")

	cmdData.ImagesRepoGitHubToken = new(string)
	cmd.Flags().StringVarP(
		cmdData.ImagesRepoGitHubToken,
		"images-repo-github-token",
		"",
		defaultValue,
		usage,
	)
}

func setupImagesRepoHarborUsername(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_IMAGES_REPO_HARBOR_USERNAME"),
		os.Getenv("WERF_REPO_HARBOR_USERNAME"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("Harbor username for images repo implementation (default $WERF_IMAGES_REPO_HARBOR_USERNAME or $WERF_REPO_HARBOR_USERNAME).")

	cmdData.ImagesRepoHarborUsername = new(string)
	cmd.Flags().StringVarP(
		cmdData.ImagesRepoHarborUsername,
		"images-repo-harbor-username",
		"",
		defaultValue,
		usage,
	)

	_ = cmd.Flags().MarkHidden("images-repo-harbor-username")
}

func setupImagesRepoHarborPassword(cmdData *CmdData, cmd *cobra.Command) {
	var defaultValue string
	for _, value := range []string{
		os.Getenv("WERF_IMAGES_REPO_HARBOR_PASSWORD"),
		os.Getenv("WERF_REPO_HARBOR_PASSWORD"),
	} {
		if value != "" {
			defaultValue = value
			break
		}
	}

	usage := fmt.Sprintf("Harbor password for images repo implementation (default $WERF_IMAGES_REPO_HARBOR_PASSWORD or $WERF_REPO_HARBOR_PASSWORD).")

	cmdData.ImagesRepoHarborPassword = new(string)
	cmd.Flags().StringVarP(
		cmdData.ImagesRepoHarborPassword,
		"images-repo-harbor-password",
		"",
		defaultValue,
		usage,
	)

	_ = cmd.Flags().MarkHidden("images-repo-harbor-password")
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
	cmd.Flags().BoolVarP(cmdData.DryRun, "dry-run", "", false, "Indicate what the command would do without actually doing that")
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
	cmdData.Set = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.Set, "set", "", []string{}, "Set helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
}

func SetupSetString(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SetString = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SetString, "set-string", "", []string{}, "Set STRING helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
}

func SetupValues(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Values = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.Values, "values", "", []string{}, "Specify helm values in a YAML file or a URL (can specify multiple)")
}

func SetupSecretValues(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SecretValues = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SecretValues, "secret-values", "", []string{}, "Specify helm secret values in a YAML file (can specify multiple)")
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

	imagesRepoImplementation := *cmdData.ImagesRepoImplementation
	if imagesRepoImplementation == "" {
		imagesRepoImplementation = *cmdData.RepoImplementation
	}

	if err := validateRepoImplementation(imagesRepoImplementation); err != nil {
		return nil, err
	}

	imagesRepoDockerHubToken := *cmdData.ImagesRepoDockerHubToken
	if imagesRepoDockerHubToken == "" {
		imagesRepoDockerHubToken = *cmdData.RepoDockerHubToken
	}

	imagesRepoDockerHubUsername := *cmdData.ImagesRepoDockerHubUsername
	if imagesRepoDockerHubUsername == "" {
		imagesRepoDockerHubUsername = *cmdData.RepoDockerHubUsername
	}

	imagesRepoDockerHubPassword := *cmdData.ImagesRepoDockerHubPassword
	if imagesRepoDockerHubPassword == "" {
		imagesRepoDockerHubPassword = *cmdData.RepoDockerHubPassword
	}

	imagesRepoGitHubToken := *cmdData.ImagesRepoGitHubToken
	if imagesRepoGitHubToken == "" {
		imagesRepoGitHubToken = *cmdData.RepoGitHubToken
	}

	imagesRepoHarborUsername := *cmdData.ImagesRepoHarborUsername
	if imagesRepoHarborUsername == "" {
		imagesRepoHarborUsername = *cmdData.RepoHarborUsername
	}

	imagesRepoHarborPassword := *cmdData.ImagesRepoHarborPassword
	if imagesRepoHarborPassword == "" {
		imagesRepoHarborPassword = *cmdData.RepoHarborPassword
	}

	return storage.NewImagesRepo(
		projectName,
		imagesRepoAddress,
		imagesRepoMode,
		storage.ImagesRepoOptions{
			DockerImagesRepoOptions: storage.DockerImagesRepoOptions{
				Implementation: imagesRepoImplementation,
				DockerRegistryOptions: docker_registry.DockerRegistryOptions{
					InsecureRegistry:      *cmdData.InsecureRegistry,
					SkipTlsVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
					DockerHubToken:        imagesRepoDockerHubToken,
					DockerHubUsername:     imagesRepoDockerHubUsername,
					DockerHubPassword:     imagesRepoDockerHubPassword,
					GitHubToken:           imagesRepoGitHubToken,
					HarborUsername:        imagesRepoHarborUsername,
					HarborPassword:        imagesRepoHarborPassword,
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

	stagesStorageRepoImplementation := *cmdData.StagesStorageRepoImplementation
	if stagesStorageRepoImplementation == "" {
		stagesStorageRepoImplementation = *cmdData.RepoImplementation
	}

	if err := validateRepoImplementation(stagesStorageRepoImplementation); err != nil {
		return nil, err
	}

	stagesStorageRepoGitHubToken := *cmdData.StagesStorageRepoGitHubToken
	if stagesStorageRepoGitHubToken == "" {
		stagesStorageRepoGitHubToken = *cmdData.RepoGitHubToken
	}

	stagesStorageRepoDockerHubUsername := *cmdData.StagesStorageRepoDockerHubUsername
	if stagesStorageRepoDockerHubUsername == "" {
		stagesStorageRepoDockerHubUsername = *cmdData.RepoDockerHubUsername
	}

	stagesStorageRepoDockerHubPassword := *cmdData.StagesStorageRepoDockerHubPassword
	if stagesStorageRepoDockerHubPassword == "" {
		stagesStorageRepoDockerHubPassword = *cmdData.RepoDockerHubPassword
	}

	stagesStorageRepoHarborUsername := *cmdData.StagesStorageRepoHarborUsername
	if stagesStorageRepoHarborUsername == "" {
		stagesStorageRepoHarborUsername = *cmdData.RepoHarborUsername
	}

	stagesStorageRepoHarborPassword := *cmdData.StagesStorageRepoHarborPassword
	if stagesStorageRepoHarborPassword == "" {
		stagesStorageRepoHarborPassword = *cmdData.RepoHarborPassword
	}

	return storage.NewStagesStorage(
		stagesStorageAddress,
		containerRuntime,
		storage.StagesStorageOptions{
			RepoStagesStorageOptions: storage.RepoStagesStorageOptions{
				Implementation: stagesStorageRepoImplementation,
				DockerRegistryOptions: docker_registry.DockerRegistryOptions{
					InsecureRegistry:      *cmdData.InsecureRegistry,
					SkipTlsVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
					DockerHubToken:        stagesStorageRepoGitHubToken,
					DockerHubUsername:     stagesStorageRepoDockerHubUsername,
					DockerHubPassword:     stagesStorageRepoDockerHubPassword,
					HarborUsername:        stagesStorageRepoHarborUsername,
					HarborPassword:        stagesStorageRepoHarborPassword,
				},
			},
		},
	)
}

func GetSynchronization(cmdData *CmdData) (string, error) {
	if *cmdData.Synchronization != ":local" {
		return "", fmt.Errorf("only --synchronization=:local is supported for now, got '%s'", *cmdData.Synchronization)
	}
	return *cmdData.Synchronization, nil
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

func GetOptionalWerfConfig(projectDir string, logRenderedFilePath bool) (*config.WerfConfig, error) {
	werfConfigPath, err := GetWerfConfigPath(projectDir, false)
	if err != nil {
		return nil, err
	}
	if werfConfigPath != "" {
		return config.GetWerfConfig(werfConfigPath, logRenderedFilePath)
	}
	return nil, nil
}

func GetRequiredWerfConfig(projectDir string, logRenderedFilePath bool) (*config.WerfConfig, error) {
	werfConfigPath, err := GetWerfConfigPath(projectDir, true)
	if err != nil {
		return nil, err
	}
	return config.GetWerfConfig(werfConfigPath, logRenderedFilePath)
}

func GetWerfConfigPath(projectDir string, required bool) (string, error) {
	for _, werfConfigName := range []string{"werf.yml", "werf.yaml"} {
		werfConfigPath := filepath.Join(projectDir, werfConfigName)
		if exist, err := util.FileExists(werfConfigPath); err != nil {
			return "", err
		} else if exist {
			return werfConfigPath, err
		}
	}

	if required {
		return "", errors.New("werf.yaml not found")
	} else {
		return "", nil
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

func validateRepoImplementation(implementation string) error {
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

func GetStagesStorageCache() storage.StagesStorageCache {
	return storage.NewFileStagesStorageCache(filepath.Join(werf.GetLocalCacheDir(), "stages_storage_v3"))
}
