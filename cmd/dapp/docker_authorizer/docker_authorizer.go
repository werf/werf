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

	Credentials     *DockerCredentials
	PullCredentials *DockerCredentials
	PushCredentials *DockerCredentials
}

func (a *DockerAuthorizer) LoginForPull(repo string) error {
	fmt.Printf("# Login into docker repo %s for pull\n", repo)
	return a.login(a.PullCredentials, repo)
}

func (a *DockerAuthorizer) LoginForPush(repo string) error {
	fmt.Printf("# Login into docker repo %s for push\n", repo)
	return a.login(a.PushCredentials, repo)
}

func (a *DockerAuthorizer) Login(repo string) error {
	return a.login(a.Credentials, repo)
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

func GetBuildDockerAuthorizer(projectTmpDir, pullUsernameOption, pullPasswordOption string) (*DockerAuthorizer, error) {
	pullCredentials, err := getPullCredentials(pullUsernameOption, pullPasswordOption)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials for pull: %s", err)
	}

	return getDockerAuthorizer(projectTmpDir, nil, pullCredentials, nil)
}

func GetPushDockerAuthorizer(projectTmpDir, pushUsernameOption, pushPasswordOption, repo string) (*DockerAuthorizer, error) {
	pushCredentials, err := getPushCredentials(pushUsernameOption, pushPasswordOption, repo)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials for push: %s", err)
	}

	return getDockerAuthorizer(projectTmpDir, nil, nil, pushCredentials)
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

	return getDockerAuthorizer(projectTmpDir, nil, pullCredentials, pushCredentials)
}

func GetFlushDockerAuthorizer(projectTmpDir, flushUsernameOption, flushPasswordOption string) (*DockerAuthorizer, error) {
	credentials, err := getFlushCredentials(flushUsernameOption, flushPasswordOption)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials for flush: %s", err)
	}

	return getDockerAuthorizer(projectTmpDir, credentials, nil, nil)
}

func GetSyncDockerAuthorizer(projectTmpDir, syncUsernameOption, syncPasswordOption, repo string) (*DockerAuthorizer, error) {
	credentials, err := getSyncCredentials(syncUsernameOption, syncPasswordOption, repo)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials for sync: %s", err)
	}

	return getDockerAuthorizer(projectTmpDir, credentials, nil, nil)
}

func GetCleanupDockerAuthorizer(projectTmpDir, syncUsernameOption, syncPasswordOption, repo string) (*DockerAuthorizer, error) {
	credentials, err := getCleanupCredentials(syncUsernameOption, syncPasswordOption, repo)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials for cleanup: %s", err)
	}

	return getDockerAuthorizer(projectTmpDir, credentials, nil, nil)
}

func GetDeployDockerAuthorizer(projectTmpDir, usernameOption, passwordOption, repo string) (*DockerAuthorizer, error) {
	credentials, err := getDeployCredentials(usernameOption, passwordOption, repo)
	if err != nil {
		return nil, fmt.Errorf("cannot get docker credentials for deploy: %s", err)
	}

	return getDockerAuthorizer(projectTmpDir, credentials, nil, nil)
}

func getDockerAuthorizer(projectTmpDir string, credentials, pullCredentials, pushCredentials *DockerCredentials) (*DockerAuthorizer, error) {
	a := &DockerAuthorizer{Credentials: credentials, PullCredentials: pullCredentials, PushCredentials: pushCredentials}

	if dappDockerConfigEnv := os.Getenv("DAPP_DOCKER_CONFIG"); dappDockerConfigEnv != "" {
		a.HostDockerConfigDir = dappDockerConfigEnv
		a.ExternalDockerConfig = true
	} else {
		if a.Credentials != nil || a.PullCredentials != nil || a.PushCredentials != nil {
			tmpDockerConfigDir := path.Join(projectTmpDir, "docker")

			if err := os.Mkdir(tmpDockerConfigDir, os.ModePerm); err != nil {
				return nil, fmt.Errorf("error creating tmp dir %s for docker config: %s", tmpDockerConfigDir, err)
			}

			fmt.Printf("Using tmp docker config at %s\n", tmpDockerConfigDir)

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

func getPullCredentials(pullUsernameOption, pullPasswordOption string) (*DockerCredentials, error) {
	creds := getSpecifiedCredentials(pullUsernameOption, pullPasswordOption)
	if creds != nil {
		return creds, nil
	}

	return getDefaultAutologinCredentials()
}

func getPushCredentials(usernameOption, passwordOption, repo string) (*DockerCredentials, error) {
	return getDefaultCredentials(usernameOption, passwordOption, repo)
}

func getFlushCredentials(usernameOption, passwordOption string) (*DockerCredentials, error) {
	return getSpecifiedCredentials(usernameOption, passwordOption), nil
}

func getSyncCredentials(usernameOption, passwordOption, repo string) (*DockerCredentials, error) {
	return getDefaultCredentials(usernameOption, passwordOption, repo)
}

func getDeployCredentials(usernameOption, passwordOption, repo string) (*DockerCredentials, error) {
	return getDefaultCredentials(usernameOption, passwordOption, repo)
}

func getCleanupCredentials(usernameOption, passwordOption, repo string) (*DockerCredentials, error) {
	creds := getSpecifiedCredentials(usernameOption, passwordOption)
	if creds != nil {
		return creds, nil
	}

	dappCleanupRegistryPassword := os.Getenv("DAPP_CLEANUP_REGISTRY_PASSWORD")
	if dappCleanupRegistryPassword != "" {
		return &DockerCredentials{Username: "dapp-cleanup", Password: dappCleanupRegistryPassword}, nil
	}

	isGCR, err := isGCR(repo)
	if err != nil {
		return nil, err
	}
	if isGCR {
		return nil, nil
	}

	return getDefaultAutologinCredentials()
}

func getDefaultCredentials(usernameOption, passwordOption, repo string) (*DockerCredentials, error) {
	creds := getSpecifiedCredentials(usernameOption, passwordOption)
	if creds != nil {
		return creds, nil
	}

	isGCR, err := isGCR(repo)
	if err != nil {
		return nil, err
	}
	if isGCR {
		return nil, nil
	}

	return getDefaultAutologinCredentials()
}

func getSpecifiedCredentials(usernameOption, passwordOption string) *DockerCredentials {
	if usernameOption != "" && passwordOption != "" {
		return &DockerCredentials{Username: usernameOption, Password: passwordOption}
	}

	return nil
}

func getDefaultAutologinCredentials() (*DockerCredentials, error) {
	if os.Getenv("DAPP_IGNORE_CI_DOCKER_AUTOLOGIN") == "" {
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
		return docker_registry.IsGCR(repoOption)
	}

	return false, nil
}
