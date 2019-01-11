package common

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/util/file"

	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/git_repo"
	"github.com/flant/dapp/pkg/slug"
	"github.com/flant/kubedog/pkg/kube"
)

type CmdData struct {
	Name    *string
	Dir     *string
	TmpDir  *string
	HomeDir *string
	SSHKeys *[]string

	Tag        *[]string
	TagBranch  *bool
	TagBuildID *bool
	TagCI      *bool
	TagCommit  *bool

	Environment *string
	Release     *string
	Namespace   *string
	KubeContext *string
}

func SetupName(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Name = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.Name, "name", "", "", `Use custom dapp name.
Chaging default name will cause full cache rebuild.
By default dapp name is the last element of remote.origin.url from project git,
or it is the name of the directory where Dappfile resides.`)
}

func SetupDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Dir = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.Dir, "dir", "", "", "Change to the specified directory to find dappfile")
}

func SetupTmpDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.TmpDir = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.TmpDir, "tmp-dir", "", "", "Use specified dir to store tmp files and dirs (use system tmp dir by default)")
}

func SetupHomeDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HomeDir = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.HomeDir, "home-dir", "", "", "Use specified dir to store dapp cache files and dirs (use ~/.dapp by default)")
}

func SetupSSHKey(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SSHKeys = new([]string)
	cmd.PersistentFlags().StringArrayVarP(cmdData.SSHKeys, "ssh-key", "", []string{}, "Enable only specified ssh keys (use system ssh-agent by default)")
}

func SetupTag(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Tag = new([]string)
	cmdData.TagBranch = new(bool)
	cmdData.TagBuildID = new(bool)
	cmdData.TagCI = new(bool)
	cmdData.TagCommit = new(bool)

	cmd.PersistentFlags().StringArrayVarP(cmdData.Tag, "tag", "", []string{}, "Add tag (can be used one or more times)")
	cmd.PersistentFlags().BoolVarP(cmdData.TagBranch, "tag-branch", "", false, "Tag by git branch")
	cmd.PersistentFlags().BoolVarP(cmdData.TagBuildID, "tag-build-id", "", false, "Tag by CI build id")
	cmd.PersistentFlags().BoolVarP(cmdData.TagCI, "tag-ci", "", false, "Tag by CI branch and tag")
	cmd.PersistentFlags().BoolVarP(cmdData.TagCommit, "tag-commit", "", false, "Tag by git commit")
}

func SetupEnvironment(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Environment = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.Release, "environment", "", "", "Use specified environment (use CI_ENVIRONMENT_SLUG by default). Environment is a required parameter and should be specified with option or CI_ENVIRONMENT_SLUG variable.")
}

func SetupRelease(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Release = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.Release, "release", "", "", "Use specified Helm release name (use %project-%environment template by default)")
}

func SetupNamespace(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Namespace = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.Release, "namespace", "", "", "Use specified Kubernetes namespace (use %project-%environment template by default)")
}

func SetupKubeContext(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeContext = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.Release, "kube-context", "", "", "Kubernetes config context")
}

func GetProjectName(cmdData *CmdData, projectDir string) (string, error) {
	name := path.Base(projectDir)

	if *cmdData.Name != "" {
		name = *cmdData.Name
	} else {
		exist, err := IsGitOwnRepoExists(projectDir)
		if err != nil {
			return "", err
		}

		if exist {
			remoteOriginUrl, err := gitOwnRepoOriginUrl(projectDir)
			if err != nil {
				return "", err
			}

			if remoteOriginUrl != "" {
				parts := strings.Split(remoteOriginUrl, "/")
				repoName := parts[len(parts)-1]

				gitEnding := ".git"
				if strings.HasSuffix(repoName, gitEnding) {
					repoName = repoName[0 : len(repoName)-len(gitEnding)]
				}

				name = repoName
			}
		}
	}

	return slug.Slug(name), nil
}

func GetDappfile(projectDir string) (*config.Dappfile, error) {
	for _, dappfileName := range []string{"dappfile.yml", "dappfile.yaml"} {
		dappfilePath := path.Join(projectDir, dappfileName)
		if exist, err := file.FileExists(dappfilePath); err != nil {
			return nil, err
		} else if exist {
			return config.ParseDimgs(dappfilePath)
		}
	}

	return nil, errors.New("dappfile.y[a]ml not found")
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

func GetProjectBuildDir(projectName string) (string, error) {
	projectBuildDir := path.Join(dapp.GetHomeDir(), "builds", projectName)

	if err := os.MkdirAll(projectBuildDir, os.ModePerm); err != nil {
		return "", err
	}

	return projectBuildDir, nil
}

func IsGitOwnRepoExists(projectDir string) (bool, error) {
	fileInfo, err := os.Stat(path.Join(projectDir, ".git"))
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	return fileInfo.IsDir(), nil
}

func GetLocalGitRepo(projectDir string) *git_repo.Local {
	return &git_repo.Local{
		Path:   projectDir,
		GitDir: path.Join(projectDir, ".git"),
	}
}

func gitOwnRepoOriginUrl(projectDir string) (string, error) {
	remoteOriginUrl, err := GetLocalGitRepo(projectDir).RemoteOriginUrl()
	if err != nil {
		return "", nil
	}

	return remoteOriginUrl, nil
}

func GetRequiredRepoName(projectName, repoOption string) (string, error) {
	res := GetOptionalRepoName(projectName, repoOption)
	if res == "" {
		return "", fmt.Errorf("CI_REGISTRY_IMAGE variable or --repo option required!")
	}
	return res, nil
}

func GetOptionalRepoName(projectName, repoOption string) string {
	if repoOption == ":minikube" {
		return fmt.Sprintf("localhost:5000/%s", projectName)
	} else if repoOption != "" {
		return repoOption
	}

	ciRegistryImage := os.Getenv("CI_REGISTRY_IMAGE")
	if ciRegistryImage != "" {
		return ciRegistryImage
	}

	return ""
}

func GetNamespace(namespaceOption string) string {
	if namespaceOption == "" {
		return kube.DefaultNamespace
	}
	return namespaceOption
}
