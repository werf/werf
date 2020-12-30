package suite_init

import (
	"fmt"
	"os"
	"strings"

	"github.com/prashantv/gostub"

	"github.com/onsi/ginkgo"
	"github.com/werf/werf/integration/utils"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
)

type ContainerRegistryPerImplementationData struct {
	ContainerRegistryPerImplementation map[string]*containerRegistryImplementationData
}

type containerRegistryImplementationData struct {
	Repo               string
	ImplementationName string
	RegistryOptions    docker_registry.DockerRegistryOptions
	StagesStorage      storage.StagesStorage
}

func NewContainerRegistryPerImplementationData(projectNameData *ProjectNameData, stubsData *StubsData, implementationName string) *ContainerRegistryPerImplementationData {
	data := &ContainerRegistryPerImplementationData{}

	ginkgo.BeforeEach(func() {
		if data.ContainerRegistryPerImplementation == nil {
			data.ContainerRegistryPerImplementation = make(map[string]*containerRegistryImplementationData)
		}

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

		implData := &containerRegistryImplementationData{
			Repo:               ContainerRegistryImplementationStagesStorageAddress(implementationNameForWerf, projectNameData.ProjectName),
			ImplementationName: implementationNameForWerf,
			RegistryOptions:    ContainerRegistryImplementationDockerRegistryOptionsAndSetEnvs(implementationNameForWerf, stubsData.Stubs),
		}

		implData.StagesStorage = utils.NewStagesStorage(implData.Repo, implementationNameForWerf, implData.RegistryOptions)

		data.ContainerRegistryPerImplementation[implementationName] = implData

		stubsData.Stubs.SetEnv("WERF_REPO", implData.Repo)
		stubsData.Stubs.SetEnv("WERF_REPO_IMPLEMENTATION", implData.ImplementationName)
	})

	return data
}

func ContainerRegistryImplementationDockerRegistryOptionsAndSetEnvs(implementationName string, stubs *gostub.Stubs) docker_registry.DockerRegistryOptions {
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

		stubs.SetEnv("WERF_REPO_DOCKER_HUB_USERNAME", username)
		stubs.SetEnv("WERF_REPO_DOCKER_HUB_PASSWORD", password)

		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
			DockerHubUsername:     username,
			DockerHubPassword:     password,
		}
	case docker_registry.GitHubPackagesImplementationName:
		token := utils.GetRequiredEnv(tokenEnvName)

		stubs.SetEnv("WERF_REPO_GITHUB_TOKEN", token)

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

func ContainerRegistryImplementationStagesStorageAddress(implementationName, projectName string) string {
	implementationCode := strings.ToUpper(implementationName)

	registryEnvName := fmt.Sprintf(
		"WERF_TEST_%s_REGISTRY",
		implementationCode,
	)

	registry := utils.GetRequiredEnv(registryEnvName)

	return fmt.Sprintf("%s/%s", registry, projectName)
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
