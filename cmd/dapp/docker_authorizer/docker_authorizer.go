package docker_authorizer

import (
	"fmt"
	"os"
	"path"

	"github.com/flant/dapp/pkg/docker"

	"github.com/flant/dapp/pkg/docker_registry"
)

type DockerCredentials struct {
	Username, Password string
}

type DockerAuthorizer struct {
	HostDockerConfigDir  string
	ExternalDockerConfig bool

	PullCredentials *DockerCredentials
	PushCredentials *DockerCredentials
}

func (a *DockerAuthorizer) LoginForPull(repo string) error {
	return a.login(a.PullCredentials, repo)
}

func (a *DockerAuthorizer) LoginForPush(repo string) error {
	return a.login(a.PushCredentials, repo)
}

func (a *DockerAuthorizer) login(creds *DockerCredentials, repo string) error {
	if a.ExternalDockerConfig || creds == nil {
		return nil
	}
	return docker.Login(creds.Username, creds.Password, repo)
}

func GetBuildDockerAuthorizer(projectTmpDir, pullUsernameOption, pullPasswordOption string) (*DockerAuthorizer, error) {
	pullCredentials, err := getPullCredentials(pullUsernameOption, pullPasswordOption)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials for pull: %s", err)
	}

	return getDockerAuthorizer(projectTmpDir, pullCredentials, nil)
}

func GetPushDockerAuthorizer(projectTmpDir, pushUsernameOption, pushPasswordOption, repo string) (*DockerAuthorizer, error) {
	pushCredentials, err := getPushCredentials(pushUsernameOption, pushPasswordOption, repo)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials for push: %s", err)
	}

	return getDockerAuthorizer(projectTmpDir, nil, pushCredentials)
}

func GetBPDockerAuthorizer(projectTmpDir, pullUsernameOption, pullPasswordOption, pushUsernameOption, pushPasswordOption, repo string) (*DockerAuthorizer, error) {
	pullCredentials, err := getPullCredentials(pullUsernameOption, pullPasswordOption)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials for pull: %s", err)
	}

	pushCredentials, err := getPushCredentials(pushUsernameOption, pushPasswordOption, repo)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials for push: %s", err)
	}

	return getDockerAuthorizer(projectTmpDir, pullCredentials, pushCredentials)
}

func getDockerAuthorizer(projectTmpDir string, pullCredentials, pushCredentials *DockerCredentials) (*DockerAuthorizer, error) {
	a := &DockerAuthorizer{PullCredentials: pullCredentials, PushCredentials: pushCredentials}

	if dappDockerConfigEnv := os.Getenv("DAPP_DOCKER_CONFIG"); dappDockerConfigEnv != "" {
		a.HostDockerConfigDir = dappDockerConfigEnv
		a.ExternalDockerConfig = true
	} else {
		if a.PullCredentials != nil || a.PushCredentials != nil {
			tmpDockerConfigDir := path.Join(projectTmpDir, "docker")

			if err := os.Mkdir(tmpDockerConfigDir, os.ModePerm); err != nil {
				return nil, fmt.Errorf("error creating tmp dir %s for docker config: %s", tmpDockerConfigDir, err)
			}

			fmt.Printf("Using tmp docker config at %s\n", tmpDockerConfigDir)

			a.HostDockerConfigDir = tmpDockerConfigDir
		} else {
			a.HostDockerConfigDir = path.Join(os.Getenv("HOME"), ".docker")
			a.ExternalDockerConfig = true
		}
	}

	return a, nil
}

func getPullCredentials(pullUsernameOption, pullPasswordOption string) (*DockerCredentials, error) {
	if pullUsernameOption != "" && pullPasswordOption != "" {
		return &DockerCredentials{Username: pullUsernameOption, Password: pullPasswordOption}, nil
	}

	if os.Getenv("DAPP_IGNORE_CI_DOCKER_AUTOLOGIN") == "" {
		ciRegistryEnv := os.Getenv("CI_REGISTRY")
		ciJobTokenEnv := os.Getenv("CI_JOB_TOKEN")
		if ciRegistryEnv != "" && ciJobTokenEnv != "" {
			return &DockerCredentials{Username: "gitlab-ci-token", Password: ciJobTokenEnv}, nil
		}
	}

	return nil, nil
}

func getPushCredentials(pushUsernameOption, pushPasswordOption, repo string) (*DockerCredentials, error) {
	if pushUsernameOption != "" && pushPasswordOption != "" {
		return &DockerCredentials{Username: pushUsernameOption, Password: pushPasswordOption}, nil
	}

	isGCR, err := isGCR(repo)
	if err != nil {
		return nil, err
	}
	if !isGCR && os.Getenv("DAPP_IGNORE_CI_DOCKER_AUTOLOGIN") == "" {
		ciRegistryEnv := os.Getenv("CI_REGISTRY")
		ciJobTokenEnv := os.Getenv("CI_JOB_TOKEN")
		if ciRegistryEnv != "" && ciJobTokenEnv != "" {
			return &DockerCredentials{Username: "gitlab-ci-token", Password: ciJobTokenEnv}, nil
		}
	}

	return nil, nil
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
