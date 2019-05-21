package common

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"
	cleanup "github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

type CmdData struct {
	Dir     *string
	TmpDir  *string
	HomeDir *string
	SSHKeys *[]string

	TagCustom    *[]string
	TagGitBranch *string
	TagGitTag    *string
	TagGitCommit *string

	Environment                 *string
	Release                     *string
	Namespace                   *string
	AddAnnotations              *[]string
	AddLabels                   *[]string
	KubeContext                 *string
	KubeConfig                  *string
	HelmReleaseStorageNamespace *string
	HelmReleaseStorageType      *string

	Set             *[]string
	SetString       *[]string
	Values          *[]string
	SecretValues    *[]string
	IgnoreSecretKey *bool

	StagesStorage *string
	ImagesRepo    *string

	DockerConfig *string
	InsecureRepo *bool
	DryRun       *bool

	GitTagStrategyLimit         *int64
	GitTagStrategyExpiryDays    *int64
	GitCommitStrategyLimit      *int64
	GitCommitStrategyExpiryDays *int64

	LogPretty        *bool
	LogColorMode     *string
	LogProjectDir    *bool
	LogTerminalWidth *int64
}

const (
	CleaningCommandsForceOptionDescription = "Remove containers that are based on deleting werf docker images"
)

func GetLongCommandDescription(text string) string {
	return logboek.FitText(text, logboek.FitTextOptions{MaxWidth: 100})
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
	cmd.Flags().StringArrayVarP(cmdData.SSHKeys, "ssh-key", "", []string{}, "Use only specific ssh keys (Defaults to system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see https://werf.io/reference/toolbox/ssh.html). Option can be specified multiple times to use multiple keys.")
}

func SetupImagesCleanupPolicies(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.GitTagStrategyLimit = new(int64)
	cmdData.GitTagStrategyExpiryDays = new(int64)
	cmdData.GitCommitStrategyLimit = new(int64)
	cmdData.GitCommitStrategyExpiryDays = new(int64)

	cmd.Flags().Int64VarP(cmdData.GitTagStrategyLimit, "git-tag-strategy-limit", "", -1, "Keep max number of images published with the git-tag tagging strategy in the images repo. No limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_TAG_STRATEGY_LIMIT.")
	cmd.Flags().Int64VarP(cmdData.GitTagStrategyExpiryDays, "git-tag-strategy-expiry-days", "", -1, "Keep images published with the git-tag tagging strategy in the images repo for the specified maximum days since image published. Republished image will be kept specified maximum days since new publication date. No days limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS.")
	cmd.Flags().Int64VarP(cmdData.GitCommitStrategyLimit, "git-commit-strategy-limit", "", -1, "Keep max number of images published with the git-commit tagging strategy in the images repo. No limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_COMMIT_STRATEGY_LIMIT.")
	cmd.Flags().Int64VarP(cmdData.GitCommitStrategyExpiryDays, "git-commit-strategy-expiry-days", "", -1, "Keep images published with the git-commit tagging strategy in the images repo for the specified maximum days since image published. Republished image will be kept specified maximum days since new publication date. No days limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS.")
}

func SetupTag(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.TagCustom = new([]string)
	cmdData.TagGitBranch = new(string)
	cmdData.TagGitTag = new(string)
	cmdData.TagGitCommit = new(string)

	cmd.Flags().StringArrayVarP(cmdData.TagCustom, "tag-custom", "", []string{}, "Use custom tagging strategy and tag by the specified arbitrary tags. Option can be used multiple times to produce multiple images with the specified tags.")
	cmd.Flags().StringVarP(cmdData.TagGitBranch, "tag-git-branch", "", os.Getenv("WERF_TAG_GIT_BRANCH"), "Use git-branch tagging strategy and tag by the specified git branch (option can be enabled by specifying git branch in the $WERF_TAG_GIT_BRANCH)")
	cmd.Flags().StringVarP(cmdData.TagGitTag, "tag-git-tag", "", os.Getenv("WERF_TAG_GIT_TAG"), "Use git-tag tagging strategy and tag by the specified git tag (option can be enabled by specifying git tag in the $WERF_TAG_GIT_TAG)")
	cmd.Flags().StringVarP(cmdData.TagGitCommit, "tag-git-commit", "", os.Getenv("WERF_TAG_GIT_COMMIT"), "Use git-commit tagging strategy and tag by the specified git commit hash (option can be enabled by specifying git commit hash in the $WERF_TAG_GIT_COMMIT)")
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

func SetupStagesStorage(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.StagesStorage, "stages-storage", "s", os.Getenv("WERF_STAGES_STORAGE"), "Docker Repo to store stages or :local for non-distributed build (only :local is supported for now; default $WERF_STAGES_STORAGE environment).\nMore info about stages: https://werf.io/reference/build/stages_and_images.html")
}

func SetupImagesRepo(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ImagesRepo = new(string)
	cmd.Flags().StringVarP(cmdData.ImagesRepo, "images-repo", "i", os.Getenv("WERF_IMAGES_REPO"), "Docker Repo to store images (default $WERF_IMAGES_REPO)")
}

func SetupInsecureRepo(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.InsecureRepo = new(bool)
	cmd.Flags().BoolVarP(cmdData.InsecureRepo, "insecure-repo", "", getBoolEnvironment("WERF_INSECURE_REPO"), "Allow usage of insecure docker repos (default $WERF_INSECURE_REPO)")
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

	desc := "Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or ~/.docker (in the order of priority)."

	if extraDesc != "" {
		desc += "\n"
		desc += extraDesc
	}

	cmd.Flags().StringVarP(cmdData.DockerConfig, "docker-config", "", defaultValue, desc)
}

func SetupLogOptions(cmdData *CmdData, cmd *cobra.Command) {
	SetupLogColor(cmdData, cmd)
	SetupLogPretty(cmdData, cmd)
	SetupTerminalWidth(cmdData, cmd)
}

func SetupLogColor(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogColorMode = new(string)

	logColorEnvironmentValue := os.Getenv("WERF_LOG_COLOR_MODE")

	defaultValue := "auto"
	if logColorEnvironmentValue != "" {
		defaultValue = logColorEnvironmentValue
	}

	cmd.Flags().StringVarP(cmdData.LogColorMode, "log-color-mode", "", defaultValue, `Set log color mode.
Supported on, off and auto (based on the stdout's file descriptor referring to a terminal) modes.
Default $WERF_LOG_COLOR_MODE or auto mode.`)
}

func SetupLogPretty(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogPretty = new(bool)

	var defaultValue bool
	if os.Getenv("WERF_LOG_PRETTY") != "" {
		defaultValue = getBoolEnvironment("WERF_LOG_PRETTY")
	} else {
		defaultValue = true
	}

	cmd.Flags().BoolVarP(cmdData.LogPretty, "log-pretty", "", defaultValue, `Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or true).`)
}

func SetupTerminalWidth(cmdData *CmdData, cmd *cobra.Command) {
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
	cmd.Flags().BoolVarP(cmdData.IgnoreSecretKey, "ignore-secret-key", "", getBoolEnvironment("WERF_IGNORE_SECRET_KEY"), "Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)")
}

func SetupLogProjectDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogProjectDir = new(bool)
	cmd.Flags().BoolVarP(cmdData.LogProjectDir, "log-project-dir", "", getBoolEnvironment("WERF_LOG_PROJECT_DIR"), `Print current project directory path (default $WERF_LOG_PROJECT_DIR)`)
}

func getBoolEnvironment(environmentName string) bool {
	switch os.Getenv(environmentName) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
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

	return res, nil
}

func GetStagesRepo(cmdData *CmdData) (string, error) {
	if *cmdData.StagesStorage == "" {
		return "", fmt.Errorf("--stages-storage :local param required")
	} else if *cmdData.StagesStorage != ":local" {
		return "", fmt.Errorf("only --stages-storage :local is supported for now, got '%s'", *cmdData.StagesStorage)
	}
	return *cmdData.StagesStorage, nil
}

func GetImagesRepo(projectName string, cmdData *CmdData) (string, error) {
	if *cmdData.ImagesRepo == "" {
		return "", fmt.Errorf("--images-repo REPO param required")
	}
	return GetOptionalImagesRepo(projectName, cmdData), nil
}

func GetOptionalImagesRepo(projectName string, cmdData *CmdData) string {
	repoOption := *cmdData.ImagesRepo

	if repoOption == ":minikube" {
		return fmt.Sprintf("werf-registry.kube-system.svc.cluster.local:5000/%s", projectName)
	} else if repoOption != "" {
		return repoOption
	}

	return ""
}

func GetWerfConfig(projectDir string) (*config.WerfConfig, error) {
	for _, werfConfigName := range []string{"werf.yml", "werf.yaml"} {
		werfConfigPath := path.Join(projectDir, werfConfigName)
		if exist, err := util.FileExists(werfConfigPath); err != nil {
			return nil, err
		} else if exist {
			return config.GetWerfConfig(werfConfigPath)
		}
	}

	return nil, errors.New("werf.yaml not found")
}

func GetProjectDir(cmdData *CmdData) (string, error) {
	if *cmdData.Dir != "" {
		return *cmdData.Dir, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return currentDir, nil
}

func GetNamespace(namespaceOption string) string {
	if namespaceOption == "" {
		return kube.DefaultNamespace
	}
	return namespaceOption
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

	logboek.LogHighlightLn(fmt.Sprintf("Running time %0.2f seconds", time.Now().Sub(t).Seconds()))

	return err
}

func LogVersion() {
	logboek.LogF("Version: %s\n", werf.Version)
}

func LogError(format string, a ...interface{}) {
	_ = logboek.WithoutIndent(func() error {
		msg := fmt.Sprintf(format, a...)
		msg = strings.TrimSuffix(msg, "\n")

		logboek.LogErrorLn(msg)

		return nil
	})
}
