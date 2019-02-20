package common

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/flant/kubedog/pkg/kube"
	cleanup "github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/logger"
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

	Environment *string
	Release     *string
	Namespace   *string
	KubeContext *string
	KubeConfig  *string

	StagesStorage *string
	ImagesRepo    *string

	DockerConfig *string
	InsecureRepo *bool
	DryRun       *bool

	GitTagStrategyLimit         *int64
	GitTagStrategyExpiryDays    *int64
	GitCommitStrategyLimit      *int64
	GitCommitStrategyExpiryDays *int64

	DisablePrettyLog *bool
}

func GetLongCommandDescription(text string) string {
	return logger.FitText(text, logger.FitTextOptions{MaxWidth: 100})
}

func SetupDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Dir = new(string)
	cmd.Flags().StringVarP(cmdData.Dir, "dir", "", "", "Change to the specified directory to find werf.yaml config")
}

func SetupTmpDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.TmpDir = new(string)
	cmd.Flags().StringVarP(cmdData.TmpDir, "tmp-dir", "", "", "Use specified dir to store tmp files and dirs (default $WERF_TMP environment or system tmp dir)")
}

func SetupHomeDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HomeDir = new(string)
	cmd.Flags().StringVarP(cmdData.HomeDir, "home-dir", "", "", "Use specified dir to store werf cache files and dirs (default $WERF_HOME environment or ~/.werf)")
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

	cmd.Flags().Int64VarP(cmdData.GitTagStrategyLimit, "git-tag-strategy-limit", "", -1, "Keep max number of images published with the git-tag tagging strategy in the images repo. No limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_TAG_STRATEGY_LIMIT environment variable.")
	cmd.Flags().Int64VarP(cmdData.GitTagStrategyExpiryDays, "git-tag-strategy-expiry-days", "", -1, "Keep images published with the git-tag tagging strategy in the images repo for the specified maximum days since image published. Republished image will be kept specified maximum days since new publication date. No days limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS environment variable.")
	cmd.Flags().Int64VarP(cmdData.GitCommitStrategyLimit, "git-commit-strategy-limit", "", -1, "Keep max number of images published with the git-commit tagging strategy in the images repo. No limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_COMMIT_STRATEGY_LIMIT environment variable.")
	cmd.Flags().Int64VarP(cmdData.GitCommitStrategyExpiryDays, "git-commit-strategy-expiry-days", "", -1, "Keep images published with the git-commit tagging strategy in the images repo for the specified maximum days since image published. Republished image will be kept specified maximum days since new publication date. No days limit by default, -1 disables the limit. Value can be specified by the $WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS environment variable.")
}

func SetupTag(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.TagCustom = new([]string)
	cmdData.TagGitBranch = new(string)
	cmdData.TagGitTag = new(string)
	cmdData.TagGitCommit = new(string)

	cmd.Flags().StringArrayVarP(cmdData.TagCustom, "tag-custom", "", []string{}, "Use custom tagging strategy and tag by the specified arbitrary tags. Option can be used multiple times to produce multiple images with the specified tags.")
	cmd.Flags().StringVarP(cmdData.TagGitBranch, "tag-git-branch", "", os.Getenv("WERF_TAG_GIT_BRANCH"), "Use git-branch tagging strategy and tag by the specified git branch (option can be enabled by specifying git branch in the $WERF_TAG_GIT_BRANCH environment variable)")
	cmd.Flags().StringVarP(cmdData.TagGitTag, "tag-git-tag", "", os.Getenv("WERF_TAG_GIT_TAG"), "Use git-tag tagging strategy and tag by the specified git tag (option can be enabled by specifying git tag in the $WERF_TAG_GIT_TAG environment variable)")
	cmd.Flags().StringVarP(cmdData.TagGitCommit, "tag-git-commit", "", os.Getenv("WERF_TAG_GIT_COMMIT"), "Use git-commit tagging strategy and tag by the specified git commit hash (option can be enabled by specifying git commit hash in the $WERF_TAG_GIT_COMMIT environment variable)")
}

func SetupEnvironment(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Environment = new(string)
	cmd.Flags().StringVarP(cmdData.Environment, "env", "", os.Getenv("WERF_ENV"), "Use specified environment (default $WERF_ENV)")
}

func SetupRelease(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Release = new(string)
	cmd.Flags().StringVarP(cmdData.Release, "release", "", "", "Use specified Helm release name (default %project-%environment template)")
}

func SetupNamespace(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Namespace = new(string)
	cmd.Flags().StringVarP(cmdData.Namespace, "namespace", "", "", "Use specified Kubernetes namespace (default %project-%environment template)")
}

func SetupKubeContext(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeContext = new(string)
	cmd.Flags().StringVarP(cmdData.KubeContext, "kube-context", "", "", "Kubernetes config context")
}

func SetupKubeConfig(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeConfig = new(string)
	cmd.Flags().StringVarP(cmdData.KubeConfig, "kube-config", "", "", "Kubernetes config file path")
}

func SetupStagesStorage(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.StagesStorage, "stages-storage", "s", os.Getenv("WERF_STAGES_STORAGE"), "Docker Repo to store stages or :local for non-distributed build (only :local is supported for now; default $WERF_STAGES_STORAGE environment).\nMore info about stages: https://werf.io/reference/build/stages.html")
}

func SetupImagesRepo(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ImagesRepo = new(string)
	cmd.Flags().StringVarP(cmdData.ImagesRepo, "images-repo", "i", os.Getenv("WERF_IMAGES_REPO"), "Docker Repo to store images (default $WERF_IMAGES_REPO environment)")
}

func SetupInsecureRepo(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.InsecureRepo = new(bool)
	cmd.Flags().BoolVarP(cmdData.InsecureRepo, "insecure-repo", "", false, "Allow usage of insecure docker repos")
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

func SetupDisablePrettyLog(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DisablePrettyLog = new(bool)
	cmd.Flags().BoolVarP(cmdData.DisablePrettyLog, "disable-pretty-log", "", getBoolEnvironment("WERF_DISABLE_PRETTY_LOG"), `Disable emojis, auto line wrapping and replace log process border characters with spaces (default $WERF_DISABLE_PRETTY_LOG).`)
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

func GetKubeContext(kubeContextOption string) string {
	kubeContext := os.Getenv("KUBECONTEXT")
	if kubeContext == "" {
		return kubeContextOption
	}
	return kubeContext
}

func ApplyDisablePrettyLog(cmdData *CmdData) {
	if *cmdData.DisablePrettyLog {
		logging.DisablePrettyLog()
	}
}

func LogRunningTime(f func() error) error {
	t := time.Now()
	err := f()

	logger.LogHighlightLn(fmt.Sprintf("Running time %0.2f seconds", time.Now().Sub(t).Seconds()))

	return err
}

func LogVersion() {
	logger.LogInfoF("Version: %s\n", werf.Version)
}

func LogProjectDir(dir string) {
	if os.Getenv("WERF_LOG_PROJECT_DIR") != "" {
		logger.LogInfoF("Using project dir: %s\n", dir)
	}
}
