package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/docker_registry"
	"github.com/flant/dapp/pkg/git_repo"
	"k8s.io/kubernetes/pkg/util/file"

	"github.com/flant/dapp/pkg/slug"
	"github.com/spf13/cobra"
)

var rootCmdData struct {
	Name    string
	Dir     string
	TmpDir  string
	HomeDir string
	SSHKeys []string
}

func main() {
	cmd := &cobra.Command{
		Use: "dapp",
	}

	cmd.AddCommand(
		newBuildCmd(),
		newPushCmd(),
		newBPCmd(),
	)

	cmd.PersistentFlags().StringVarP(&rootCmdData.Name, "name", "", "", `Use custom dapp name.
Chaging default name will cause full cache rebuild.
By default dapp name is the last element of remote.origin.url from project git,
or it is the name of the directory where Dappfile resides.`)
	cmd.PersistentFlags().StringVarP(&rootCmdData.Dir, "dir", "", "", "Change to the specified directory to find dappfile")
	cmd.PersistentFlags().StringVarP(&rootCmdData.TmpDir, "tmp-dir", "", "", "Use specified dir to store tmp files and dirs (use system tmp dir by default)")
	cmd.PersistentFlags().StringVarP(&rootCmdData.HomeDir, "home-dir", "", "", "Use specified dir to store dapp cache files and dirs (use ~/.dapp by default)")
	cmd.PersistentFlags().StringArrayVarP(&rootCmdData.SSHKeys, "ssh-key", "", []string{}, "Enable only specified ssh keys (use system ssh-agent by default)")

	err := cmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func getProjectName(projectDir string) (string, error) {
	name := path.Base(projectDir)

	if rootCmdData.Name != "" {
		name = rootCmdData.Name
	} else {
		exist, err := isGitOwnRepoExists(projectDir)
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

func parseDappfile(projectDir string) ([]*config.Dimg, error) {
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

func getProjectDir() (string, error) {
	if rootCmdData.Dir != "" {
		return rootCmdData.Dir, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return currentDir, nil
}

func getProjectTmpDir() (string, error) {
	return ioutil.TempDir(dapp.GetTmpDir(), "dapp-")
}

func getProjectBuildDir(projectName string) (string, error) {
	projectBuildDir := path.Join(dapp.GetHomeDir(), "build", projectName)

	if err := os.MkdirAll(projectBuildDir, os.ModePerm); err != nil {
		return "", err
	}

	return projectBuildDir, nil
}

func isGitOwnRepoExists(projectDir string) (bool, error) {
	fileInfo, err := os.Stat(path.Join(projectDir, ".git"))
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	return fileInfo.IsDir(), nil
}

func gitOwnRepoOriginUrl(projectDir string) (string, error) {
	localGitRepo := &git_repo.Local{
		Path:   projectDir,
		GitDir: path.Join(projectDir, ".git"),
	}

	remoteOriginUrl, err := localGitRepo.RemoteOriginUrl()
	if err != nil {
		return "", nil
	}

	return remoteOriginUrl, nil
}

func hostDockerConfigDir(projectTmpDir string, usernameOption, passwordOption, repoOption string) (string, error) {
	dappDockerConfigEnv := os.Getenv("DAPP_DOCKER_CONFIG")

	username, password, err := dockerCredentials(usernameOption, passwordOption, repoOption)
	if err != nil {
		return "", err
	}
	areDockerCredentialsNotEmpty := username != "" && password != ""

	if areDockerCredentialsNotEmpty && repoOption != "" {
		tmpDockerConfigDir := path.Join(projectTmpDir, "docker")

		if err := os.Mkdir(tmpDockerConfigDir, os.ModePerm); err != nil {
			return "", err
		}

		return tmpDockerConfigDir, nil
	} else if dappDockerConfigEnv != "" {
		return dappDockerConfigEnv, nil
	} else {
		return path.Join(os.Getenv("HOME"), ".docker"), nil
	}
}

func getRepoName(repoOption string) (string, error) {
	if repoOption != "" {
		return repoOption, nil
	}

	ciRegistryImage := os.Getenv("CI_REGISTRY_IMAGE")
	if ciRegistryImage != "" {
		return ciRegistryImage, nil
	}

	// TODO: repo should be fully optional for render/lint commands

	return "", fmt.Errorf("CI_REGISTRY_IMAGE variable or repo option required!")
}

func dockerCredentials(usernameOption, passwordOption, repoOption string) (string, string, error) {
	if usernameOption != "" && passwordOption != "" {
		return usernameOption, passwordOption, nil
	} else if os.Getenv("DAPP_DOCKER_CONFIG") != "" {
		return "", "", nil
	} else {
		isGCR, err := isGCR(repoOption)
		if err != nil {
			return "", "", err
		}

		dappIgnoreCIDockerAutologinEnv := os.Getenv("DAPP_IGNORE_CI_DOCKER_AUTOLOGIN")
		if isGCR || dappIgnoreCIDockerAutologinEnv != "" {
			return "", "", nil
		}

		ciRegistryEnv := os.Getenv("CI_REGISTRY")
		ciJobTokenEnv := os.Getenv("CI_JOB_TOKEN")
		if ciRegistryEnv != "" && ciJobTokenEnv != "" {
			return "gitlab-ci-token", ciJobTokenEnv, nil
		}
	}

	return "", "", nil
}

func isGCR(repoOption string) (bool, error) {
	if repoOption != "" {
		if repoOption == ":minikube" {
			return false, nil
		}

		return docker_registry.IsGCR(repoOption)
	}

	return false, nil
}

func getDockerAuthorizer(usernameOption, passwordOption, repoOption string) (*DockerAuthorizer, error) {
	username, password, err := dockerCredentials(usernameOption, passwordOption, repoOption)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials: %s", err)
	}

	return NewDockerAuthorizer(username, password), nil
}

type DockerAuthorizer struct {
	Username, Password string
}

func NewDockerAuthorizer(username, password string) *DockerAuthorizer {
	return &DockerAuthorizer{Username: username, Password: password}
}

func (a *DockerAuthorizer) LoginBaseImage(repo string) error {
	return nil
}

func (a *DockerAuthorizer) LoginForPushes(repo string) error {
	return nil
}
