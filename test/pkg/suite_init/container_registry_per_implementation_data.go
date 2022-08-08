package suite_init

import (
	"context"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/test/pkg/utils"
	utilsDocker "github.com/werf/werf/test/pkg/utils/docker"
)

const LocalRegistryImplementationName = ":local_container_registry"

type ContainerRegistryPerImplementationData struct {
	ContainerRegistryPerImplementation map[string]*containerRegistryImplementationData
}

func (data *ContainerRegistryPerImplementationData) SetupRepo(ctx context.Context, repo, implementationName string, stubsData *StubsData) bool {
	implData := data.ContainerRegistryPerImplementation[implementationName]

	registry, err := docker_registry.NewDockerRegistry(repo, implData.WerfImplementationName, implData.RegistryOptions)
	Expect(err).Should(Succeed())

	if implementationName == docker_registry.AwsEcrImplementationName {
		err := registry.CreateRepo(ctx, repo)
		Expect(err).Should(Succeed())
	}

	stubsData.Stubs.SetEnv("WERF_REPO", repo)
	stubsData.Stubs.SetEnv("WERF_REPO_IMPLEMENTATION", implData.WerfImplementationName)

	switch implementationName {
	case docker_registry.DockerHubImplementationName:
		stubsData.Stubs.SetEnv("WERF_REPO_DOCKER_HUB_USERNAME", implData.RegistryOptions.DockerHubUsername)
		stubsData.Stubs.SetEnv("WERF_REPO_DOCKER_HUB_PASSWORD", implData.RegistryOptions.DockerHubPassword)
	case docker_registry.GitHubPackagesImplementationName:
		stubsData.Stubs.SetEnv("WERF_REPO_GITHUB_TOKEN", implData.RegistryOptions.GitHubToken)
	case docker_registry.QuayImplementationName:
		stubsData.Stubs.SetEnv("WERF_PARALLEL", "0")
	}

	return false
}

func (data *ContainerRegistryPerImplementationData) TeardownRepo(ctx context.Context, repo, implementationName string, _ *StubsData) bool {
	if implementationName == LocalRegistryImplementationName {
		return true
	}

	registry, err := docker_registry.NewDockerRegistry(repo, implementationName, data.ContainerRegistryPerImplementation[implementationName].RegistryOptions)
	Expect(err).Should(Succeed())

	switch implementationName {
	case docker_registry.AzureCrImplementationName, docker_registry.AwsEcrImplementationName, docker_registry.DockerHubImplementationName, docker_registry.GitHubPackagesImplementationName, docker_registry.HarborImplementationName, docker_registry.QuayImplementationName:
		err := registry.DeleteRepo(ctx, repo)

		switch {
		case err == nil:
		case docker_registry.IsAzureCrRepositoryNotFoundErr(err),
			docker_registry.IsDockerHubRepositoryNotFoundErr(err),
			docker_registry.IsHarborRepositoryNotFoundErr(err),
			docker_registry.IsQuayRepositoryNotFoundErr(err),
			docker_registry.IsSelectelRepositoryNotFoundErr(err):
		default:
			Ω(err).Should(Succeed())
		}
	}

	return true
}

type containerRegistryImplementationData struct {
	RegistryAddress        string
	WerfImplementationName string
	RegistryOptions        docker_registry.DockerRegistryOptions
}

func NewContainerRegistryPerImplementationData(synchronizedSuiteCallbacksData *SynchronizedSuiteCallbacksData, withOptionalContainerRegistry bool) *ContainerRegistryPerImplementationData {
	data := &ContainerRegistryPerImplementationData{}

	synchronizedSuiteCallbacksData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(_ []byte) {
		implementations := ContainerRegistryImplementationListToCheck(false)
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

			registryAddress := containerRegistryImplementationAddress(implementationNameForWerf)

			implData := &containerRegistryImplementationData{
				RegistryAddress:        registryAddress,
				WerfImplementationName: implementationNameForWerf,
				RegistryOptions:        makeContainerRegistryImplementationDockerRegistryOptions(implementationName),
			}

			data.ContainerRegistryPerImplementation[implementationName] = implData
		}

		if len(implementations) == 0 && withOptionalContainerRegistry {
			setupOptionalLocalContainerRegistry(synchronizedSuiteCallbacksData, data)
		}
	})

	return data
}

func setupOptionalLocalContainerRegistry(synchronizedSuiteCallbacksData *SynchronizedSuiteCallbacksData, data *ContainerRegistryPerImplementationData) {
	implementationNameForWerf := "default"

	localRegistryAddress, _, localRegistryContainerName := utilsDocker.LocalDockerRegistryRun()
	registryAddress := localRegistryAddress

	synchronizedSuiteCallbacksData.AppendSynchronizedAfterSuiteAllNodesFunc(func() {
		utilsDocker.ContainerStopAndRemove(localRegistryContainerName)
	})

	implData := &containerRegistryImplementationData{
		RegistryAddress:        registryAddress,
		WerfImplementationName: implementationNameForWerf,
		RegistryOptions:        makeContainerRegistryImplementationDockerRegistryOptions(implementationNameForWerf),
	}

	data.ContainerRegistryPerImplementation[LocalRegistryImplementationName] = implData
}

func makeContainerRegistryImplementationDockerRegistryOptions(implementationName string) docker_registry.DockerRegistryOptions {
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
	case docker_registry.SelectelImplementationName:
		accountEnvName := fmt.Sprintf(
			"WERF_TEST_%s_ACCOUNT",
			implementationCode,
		)

		vpcEnvName := fmt.Sprintf(
			"WERF_TEST_%s_VPC",
			implementationCode,
		)

		vpcIDEnvName := fmt.Sprintf(
			"WERF_TEST_%s_VPC_ID",
			implementationCode,
		)

		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
			SelectelUsername:      utils.GetRequiredEnv(usernameEnvName),
			SelectelPassword:      utils.GetRequiredEnv(passwordEnvName),
			SelectelAccount:       utils.GetRequiredEnv(accountEnvName),
			SelectelVPC:           utils.GetRequiredEnv(vpcEnvName),
			SelectelVPCID:         utils.GetRequiredEnv(vpcIDEnvName),
		}
	default:
		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
		}
	}
}

func containerRegistryImplementationAddress(implementationName string) string {
	implementationCode := strings.ToUpper(implementationName)

	registryEnvName := fmt.Sprintf(
		"WERF_TEST_%s_REGISTRY",
		implementationCode,
	)

	return utils.GetRequiredEnv(registryEnvName)
}

func ContainerRegistryImplementationListToCheck(withOptionalContainerRegistry bool) []string {
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

	if len(list) == 0 && withOptionalContainerRegistry {
		list = append(list, LocalRegistryImplementationName)
	}

	return list
}
