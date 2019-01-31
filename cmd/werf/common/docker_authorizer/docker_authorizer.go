package docker_authorizer

import (
	"fmt"
	"os"
	"path"

	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/logger"
)

type DockerCredentials struct {
	Username, Password string
}

type DockerAuthorizer struct {
	HostDockerConfigDir  string
	ExternalDockerConfig bool

	Credentials     *DockerCredentials
	PullCredentials *DockerCredentials
	PushCredentials *DockerCredentials
}

func (a *DockerAuthorizer) LoginForPull(repo string) error {
	err := a.login(a.PullCredentials, repo)
	if err != nil {
		return err
	}

	logger.LogInfoF("Login into docker repo '%s' for pull\n", repo)

	return nil
}

func (a *DockerAuthorizer) LoginForPush(repo string) error {
	err := a.login(a.PushCredentials, repo)
	if err != nil {
		return err
	}

	logger.LogInfoF("Login into docker repo '%s' for push\n", repo)

	return nil
}

func (a *DockerAuthorizer) Login(repo string) error {
	err := a.login(a.Credentials, repo)
	if err != nil {
		return err
	}

	logger.LogInfoF("Login into docker repo '%s'\n", repo)

	return nil
}

func (a *DockerAuthorizer) login(creds *DockerCredentials, repo string) error {
	if a.ExternalDockerConfig || creds == nil {
		return nil
	}

	if err := docker.Login(creds.Username, creds.Password, repo); err != nil {
		return err
	}

	if err := docker.Init(a.HostDockerConfigDir); err != nil {
		return err
	}

	return nil
}

func GetBuildStagesDockerAuthorizer(projectTmpDir, pullUsernameOption, pullPasswordOption string) (*DockerAuthorizer, error) {
	pullCredentials := getSpecifiedCredentials(pullUsernameOption, pullPasswordOption)
	return getDockerAuthorizer(projectTmpDir, nil, pullCredentials, nil)
}

func GetImagePublishDockerAuthorizer(projectTmpDir, pushUsernameOption, pushPasswordOption string) (*DockerAuthorizer, error) {
	pushCredentials := getSpecifiedCredentials(pushUsernameOption, pushPasswordOption)
	return getDockerAuthorizer(projectTmpDir, nil, nil, pushCredentials)
}

func GetBuildAndPublishDockerAuthorizer(projectTmpDir, pullUsernameOption, pullPasswordOption, pushUsernameOption, pushPasswordOption string) (*DockerAuthorizer, error) {
	pullCredentials := getSpecifiedCredentials(pullUsernameOption, pullPasswordOption)
	pushCredentials := getSpecifiedCredentials(pushUsernameOption, pushPasswordOption)
	return getDockerAuthorizer(projectTmpDir, nil, pullCredentials, pushCredentials)
}

func GetCommonDockerAuthorizer(projectTmpDir, username, password string) (*DockerAuthorizer, error) {
	credentials := getSpecifiedCredentials(username, password)
	return getDockerAuthorizer(projectTmpDir, credentials, nil, nil)
}

func getSpecifiedCredentials(usernameOption, passwordOption string) *DockerCredentials {
	if usernameOption != "" && passwordOption != "" {
		return &DockerCredentials{Username: usernameOption, Password: passwordOption}
	}

	return nil
}

func getDockerAuthorizer(projectTmpDir string, credentials, pullCredentials, pushCredentials *DockerCredentials) (*DockerAuthorizer, error) {
	a := &DockerAuthorizer{Credentials: credentials, PullCredentials: pullCredentials, PushCredentials: pushCredentials}

	if werfDockerConfigEnv := os.Getenv("WERF_DOCKER_CONFIG"); werfDockerConfigEnv != "" {
		a.HostDockerConfigDir = werfDockerConfigEnv
		a.ExternalDockerConfig = true
	} else {
		if a.Credentials != nil || a.PullCredentials != nil || a.PushCredentials != nil {
			tmpDockerConfigDir := path.Join(projectTmpDir, "docker")

			if err := os.Mkdir(tmpDockerConfigDir, os.ModePerm); err != nil {
				return nil, fmt.Errorf("error creating tmp dir %s for docker config: %s", tmpDockerConfigDir, err)
			}

			logger.LogInfoF("Using tmp docker config at %s\n", tmpDockerConfigDir)

			a.HostDockerConfigDir = tmpDockerConfigDir
		} else {
			a.HostDockerConfigDir = GetHomeDockerConfigDir()
			a.ExternalDockerConfig = true
		}
	}

	if err := docker.Init(a.HostDockerConfigDir); err != nil {
		return nil, err
	}

	os.Setenv("DOCKER_CONFIG", a.HostDockerConfigDir)

	return a, nil
}

func GetHomeDockerConfigDir() string {
	return path.Join(os.Getenv("HOME"), ".docker")
}
