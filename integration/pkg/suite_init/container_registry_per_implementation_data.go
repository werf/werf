package suite_init

import (
	"fmt"
	"os"
	"strings"

	"github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/pkg/docker_registry"
)

type ContainerRegistryPerImplementationData struct {
	ContainerRegistryPerImplementation map[string]*containerRegistryImplementationData
}

func (data *ContainerRegistryPerImplementationData) ActivateImplementationWerfEnvironmentParams(implementationName string, stubsData *StubsData) bool {
	implData := data.ContainerRegistryPerImplementation[implementationName]

	stubsData.Stubs.SetEnv("WERF_REPO_IMPLEMENTATION", implData.ImplementationName)

	switch implementationName {
	case docker_registry.DockerHubImplementationName:
		stubsData.Stubs.SetEnv("WERF_REPO_DOCKER_HUB_USERNAME", implData.RegistryOptions.DockerHubUsername)
		stubsData.Stubs.SetEnv("WERF_REPO_DOCKER_HUB_PASSWORD", implData.RegistryOptions.DockerHubPassword)
	case docker_registry.GitHubPackagesImplementationName:
		stubsData.Stubs.SetEnv("WERF_REPO_GITHUB_TOKEN", implData.RegistryOptions.GitHubToken)
	}

	return true
}

type containerRegistryImplementationData struct {
	RegistryAddress    string
	ImplementationName string
	RegistryOptions    docker_registry.DockerRegistryOptions
}

func NewContainerRegistryPerImplementationData(synchronizedSuiteCallbacksData *SynchronizedSuiteCallbacksData) *ContainerRegistryPerImplementationData {
	data := &ContainerRegistryPerImplementationData{}

	synchronizedSuiteCallbacksData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(_ []byte) {
		implementations := ContainerRegistryImplementationListToCheck()
		gomega.Expect(len(implementations)).NotTo(gomega.Equal(0), "expected at least one of WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_<IMPLEMENTATION>=1 to be set, supported implementations: %v", docker_registry.ImplementationList())

		data.ContainerRegistryPerImplementation = make(map[string]*containerRegistryImplementationData)

		for _, implementationName := range implementations {
			isNotSupported := true
			for _, name := range docker_registry.ImplementationList() {
				if name == implementationName {
					isNotSupported = false
				}
			}
			var implementationNameForWerf string
			if isNotSupported {
				implementationNameForWerf = "default"
			} else {
				implementationNameForWerf = implementationName
			}

			registryAddress := ContainerRegistryImplementationAddress(implementationNameForWerf)

			implData := &containerRegistryImplementationData{
				RegistryAddress:    registryAddress,
				ImplementationName: implementationNameForWerf,
				RegistryOptions:    ContainerRegistryImplementationDockerRegistryOptionsAndSetEnvs(implementationNameForWerf),
			}

			data.ContainerRegistryPerImplementation[implementationName] = implData
		}
	})

	return data
}

func ContainerRegistryImplementationDockerRegistryOptionsAndSetEnvs(implementationName string) docker_registry.DockerRegistryOptions {
	implementationCode := strings.ToUpper(implementationName)

	usernameEnvName := fmt.Sprintf(
		"WERF_TEST_%s_USERNAME",
		implementationCode,
	)

	passwordEnvName := fmt.Sprintf(
		"WERF_TEST_%s_PASSWORD",
		implementationCode,
	)

	tokenEnvName := fmt.Sprintf(
		"WERF_TEST_%s_TOKEN",
		implementationCode,
	)

	switch implementationName {
	case docker_registry.DockerHubImplementationName:
		username := utils.GetRequiredEnv(usernameEnvName)
		password := utils.GetRequiredEnv(passwordEnvName)

		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
			DockerHubUsername:     username,
			DockerHubPassword:     password,
		}
	case docker_registry.GitHubPackagesImplementationName:
		token := utils.GetRequiredEnv(tokenEnvName)

		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
			GitHubToken:           token,
		}
	case docker_registry.HarborImplementationName:
		username := utils.GetRequiredEnv(usernameEnvName)
		password := utils.GetRequiredEnv(passwordEnvName)

		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
			HarborUsername:        username,
			HarborPassword:        password,
		}
	case docker_registry.QuayImplementationName:
		tokenEnvName := fmt.Sprintf(
			"WERF_TEST_%s_TOKEN",
			implementationCode,
		)

		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
			QuayToken:             utils.GetRequiredEnv(tokenEnvName),
		}
	default:
		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
		}
	}
}

func ContainerRegistryImplementationAddress(implementationName string) string {
	implementationCode := strings.ToUpper(implementationName)

	registryEnvName := fmt.Sprintf(
		"WERF_TEST_%s_REGISTRY",
		implementationCode,
	)

	return utils.GetRequiredEnv(registryEnvName)
}

func ContainerRegistryImplementationListToCheck() []string {
	var list []string

	for _, implementationName := range docker_registry.ImplementationList() {
		implementationCode := strings.ToUpper(implementationName)
		implementationFlagEnvName := fmt.Sprintf(
			"WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_%s",
			implementationCode,
		)

		if os.Getenv(implementationFlagEnvName) == "1" {
			list = append(list, implementationName)
		}
	}

environLoop:
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		envName := parts[0]
		envValue := parts[1]

		if strings.HasPrefix(envName, "WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_") && envValue == "1" {
			implementationName := strings.ToLower(strings.TrimPrefix(envName, "WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_"))

			for _, name := range list {
				if name == implementationName {
					continue environLoop
				}
			}

			list = append(list, implementationName)
		}
	}

	return list
}
