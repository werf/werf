package common

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

type CmdData struct {
	Dir     *string
	TmpDir  *string
	HomeDir *string
	SSHKeys *[]string

	Tag          *[]string
	TagGitBranch *string
	TagGitTag    *string
	TagGitCommit *string

	Environment *string
	Release     *string
	Namespace   *string
	KubeContext *string

	StagesRepo *string
	ImagesRepo *string

	DockerConfig *string

	DryRun bool
}

func GetLongCommandDescription(text string) string {
	return logger.FitTextWithIndentWithWidthMaxLimit(text, 0, 100)
}

func SetupDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Dir = new(string)
	cmd.Flags().StringVarP(cmdData.Dir, "dir", "", "", "Change to the specified directory to find werf.yaml config")
}

func SetupTmpDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.TmpDir = new(string)
	cmd.Flags().StringVarP(cmdData.TmpDir, "tmp-dir", "", "", "Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp dir by default)")
}

func SetupHomeDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HomeDir = new(string)
	cmd.Flags().StringVarP(cmdData.HomeDir, "home-dir", "", "", "Use specified dir to store werf cache files and dirs (use WERF_HOME environment or ~/.werf by default)")
}

func SetupSSHKey(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SSHKeys = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SSHKeys, "ssh-key", "", []string{}, "Use only specific ssh keys (system ssh-agent or default keys will be used by default, see https://flant.github.io/werf/reference/toolbox/ssh.html). Option can be specified multiple times to use multiple keys.")
}

func SetupTag(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Tag = new([]string)
	cmdData.TagGitBranch = new(string)
	cmdData.TagGitTag = new(string)
	cmdData.TagGitCommit = new(string)

	cmd.Flags().StringArrayVarP(cmdData.Tag, "tag", "", []string{}, "Add tag (can be used one or more times)")

	cmd.Flags().StringVarP(cmdData.TagGitBranch, "tag-git-branch", "", os.Getenv("WERF_TAG_GIT_BRANCH"), "Use git-branch tagging strategy and tag by the specified git branch (option can be enabled by specifying git branch in the WERF_TAG_GIT_BRANCH environment variable)")
	cmd.Flags().StringVarP(cmdData.TagGitTag, "tag-git-tag", "", os.Getenv("WERF_TAG_GIT_TAG"), "Use git-tag tagging strategy and tag by the specified git tag (option can be enabled by specifying git tag in the WERF_TAG_GIT_TAG environment variable)")
	cmd.Flags().StringVarP(cmdData.TagGitCommit, "tag-git-commit", "", os.Getenv("WERF_TAG_GIT_COMMIT"), "Use git-commit tagging strategy and tag by the specified git commit hash (option can be enabled by specifying git commit hash in the WERF_TAG_GIT_COMMIT environment variable)")
}

func SetupEnvironment(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Environment = new(string)
	cmd.Flags().StringVarP(cmdData.Environment, "env", "", "", "Use specified environment (use WERF_DEPLOY_ENVIRONMENT by default)")
}

func SetupRelease(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Release = new(string)
	cmd.Flags().StringVarP(cmdData.Release, "release", "", "", "Use specified Helm release name (use %project-%environment template by default)")
}

func SetupNamespace(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Namespace = new(string)
	cmd.Flags().StringVarP(cmdData.Namespace, "namespace", "", "", "Use specified Kubernetes namespace (use %project-%environment template by default)")
}

func SetupKubeContext(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeContext = new(string)
	cmd.Flags().StringVarP(cmdData.KubeContext, "kube-context", "", "", "Kubernetes config context")
}

func SetupStagesRepo(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StagesRepo = new(string)
	cmd.Flags().StringVarP(cmdData.StagesRepo, "stages", "s", "", "Docker Repo to store stages or :local for non-distributed build (only :local is supported for now)")
}

func SetupImagesRepo(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ImagesRepo = new(string)
	cmd.Flags().StringVarP(cmdData.ImagesRepo, "images", "i", os.Getenv("WERF_IMAGES_REPO"), "Docker Repo to store images (use WERF_IMAGES_REPO environment by default)")
}

func SetupDryRun(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.DryRun, "dry-run", "", false, "Indicate what the command would do without actually doing that")
}

func SetupDockerConfig(cmdData *CmdData, cmd *cobra.Command) {
	defaultValue := os.Getenv("WERF_DOCKER_CONFIG")
	if defaultValue == "" {
		defaultValue = os.Getenv("DOCKER_CONFIG")
	}

	cmdData.DockerConfig = new(string)
	cmd.Flags().StringVarP(cmdData.DockerConfig, "docker-config", "", defaultValue, "Specify docker config directory path. WERF_DOCKER_CONFIG or DOCKER_CONFIG or ~/.docker will be used by default (in the order of priority).")
}

func GetStagesRepo(cmdData *CmdData) (string, error) {
	if *cmdData.StagesRepo == "" {
		return "", fmt.Errorf("--stages :local param required")
	} else if *cmdData.StagesRepo != ":local" {
		return "", fmt.Errorf("only --stages :local is supported for now, got '%s'", *cmdData.StagesRepo)
	}
	return *cmdData.StagesRepo, nil
}

func GetImagesRepo(projectName string, cmdData *CmdData) (string, error) {
	if *cmdData.ImagesRepo == "" {
		return "", fmt.Errorf("--images REPO param required")
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
			return config.ParseWerfConfig(werfConfigPath)
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

func LogRunningTime(f func() error) error {
	t := time.Now()
	err := f()

	logger.LogServiceLn(fmt.Sprintf("Running time %0.2f seconds", time.Now().Sub(t).Seconds()))

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
