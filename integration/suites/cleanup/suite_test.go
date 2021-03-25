package cleanup_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/werf/werf/integration/pkg/suite_init"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"

	"github.com/werf/werf/integration/pkg/utils"
	utilsDocker "github.com/werf/werf/integration/pkg/utils/docker"
)

// Environment implementation variables
// WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_<implementation name>
// WERF_TEST_<implementation name>_REGISTRY
//
// export WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_ECR
// export WERF_TEST_ECR_REGISTRY
//
// export WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_DOCKERHUB
// export WERF_TEST_DOCKERHUB_REGISTRY
// export WERF_TEST_DOCKERHUB_USERNAME
// export WERF_TEST_DOCKERHUB_PASSWORD
//
// export WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_GITHUB
// export WERF_TEST_GITHUB_REGISTRY
// export WERF_TEST_GITHUB_TOKEN
//
// export WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_HARBOR
// export WERF_TEST_HARBOR_REGISTRY
// export WERF_TEST_HARBOR_USERNAME
// export WERF_TEST_HARBOR_PASSWORD
//
// export WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_QUAY
// export WERF_TEST_QUAY_REGISTRY
// export WERF_TEST_QUAY_TOKEN

const imageName = "image"

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Cleanup suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"git", "docker"},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData struct {
	suite_init.SuiteData
	LocalRegistryRepoAddress   string
	LocalRegistryContainerName string
	TestImplementation         string
	StagesStorage              storage.StagesStorage
}

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
var _ = SuiteData.SetupTmp(suite_init.NewTmpDirData())

var _ = SuiteData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(_ []byte) {
	SuiteData.LocalRegistryRepoAddress, SuiteData.LocalRegistryContainerName = utilsDocker.LocalDockerRegistryRun()
})

var _ = SuiteData.AppendSynchronizedAfterSuiteAllNodesFunc(func() {
	utilsDocker.ContainerStopAndRemove(SuiteData.LocalRegistryContainerName)
})

func forEachDockerRegistryImplementation(description string, body func()) bool {
	for _, name := range implementationListToCheck() {
		implementationName := name

		Describe(fmt.Sprintf("[%s] %s", implementationName, description), func() {
			BeforeEach(func() {
				SuiteData.TestImplementation = implementationName

				var stagesStorageAddress string
				var stagesStorageImplementationName string
				var stagesStorageDockerRegistryOptions docker_registry.DockerRegistryOptions

				if implementationName == ":local_with_stages_storage_repo" {
					stagesStorageAddress = strings.Join([]string{SuiteData.LocalRegistryRepoAddress, utils.ProjectName(), "stages"}, "/")
					stagesStorageDockerRegistryOptions = docker_registry.DockerRegistryOptions{}
				} else {
					stagesStorageAddress = implementationStagesStorageAddress(implementationName)
					implementationDockerRegistryOptions := implementationDockerRegistryOptionsAndSetEnvs(implementationName)
					stagesStorageDockerRegistryOptions = implementationDockerRegistryOptions
				}

				isNotSupported := true
				for _, name := range docker_registry.ImplementationList() {
					if name == implementationName {
						isNotSupported = false
					}
				}

				if isNotSupported {
					stagesStorageImplementationName = "default"
					SuiteData.Stubs.SetEnv("WERF_REPO_IMPLEMENTATION", stagesStorageImplementationName)
				}

				initStagesStorage(stagesStorageAddress, stagesStorageImplementationName, stagesStorageDockerRegistryOptions)

				SuiteData.Stubs.SetEnv("WERF_REPO", stagesStorageAddress)
			})

			BeforeEach(func() {
				implementationBeforeEach(implementationName)
			})

			AfterEach(func() {
			afterEach:
				combinedOutput, err := utils.RunCommand(
					SuiteData.TestDirPath,
					SuiteData.WerfBinPath,
					"purge",
				)

				if err != nil {
					if implementationName == docker_registry.QuayImplementationName {
						if strings.Contains(string(combinedOutput), "TAG_EXPIRED") {
							time.Sleep(5)
							goto afterEach
						}
					}

					Ω(err).ShouldNot(HaveOccurred())
				}

				implementationAfterEach(implementationName)
			})

			body()
		})
	}

	return true
}

func initStagesStorage(stagesStorageAddress string, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) {
	SuiteData.StagesStorage = utils.NewStagesStorage(stagesStorageAddress, implementationName, dockerRegistryOptions)
}

func initLocalStagesStorage() {
	SuiteData.StagesStorage = utils.NewStagesStorage(":local", "default", docker_registry.DockerRegistryOptions{})
}

func StagesCount() int {
	return utils.StagesCount(context.Background(), SuiteData.StagesStorage)
}

func ManagedImagesCount() int {
	return utils.ManagedImagesCount(context.Background(), SuiteData.StagesStorage)
}

func ImageMetadata(imageName string) map[string][]string {
	return utils.ImageMetadata(context.Background(), SuiteData.StagesStorage, imageName)
}

func RmImportMetadata(importSourceID string) {
	utils.RmImportMetadata(context.Background(), SuiteData.StagesStorage, importSourceID)
}

func ImportMetadataIDs() []string {
	return utils.ImportMetadataIDs(context.Background(), SuiteData.StagesStorage)
}

func implementationListToCheck() []string {
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

	if len(list) != 0 {
		return list
	} else {
		return []string{
			":local_with_stages_storage_repo",
		}
	}
}

func implementationStagesStorageAddress(implementationName string) string {
	projectName := utils.ProjectName()
	implementationCode := strings.ToUpper(implementationName)

	registryEnvName := fmt.Sprintf(
		"WERF_TEST_%s_REGISTRY",
		implementationCode,
	)

	registry := utils.GetRequiredEnv(registryEnvName)

	return fmt.Sprintf("%s/%s-stages", registry, projectName)
}

func implementationDockerRegistryOptionsAndSetEnvs(implementationName string) docker_registry.DockerRegistryOptions {
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

		SuiteData.Stubs.SetEnv("WERF_REPO_DOCKER_HUB_USERNAME", username)
		SuiteData.Stubs.SetEnv("WERF_REPO_DOCKER_HUB_PASSWORD", password)

		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
			DockerHubUsername:     username,
			DockerHubPassword:     password,
		}
	case docker_registry.GitHubPackagesImplementationName:
		token := utils.GetRequiredEnv(tokenEnvName)

		SuiteData.Stubs.SetEnv("WERF_REPO_GITHUB_TOKEN", token)

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

func implementationBeforeEach(implementationName string) {
	switch implementationName {
	case docker_registry.AwsEcrImplementationName:
		err := SuiteData.StagesStorage.CreateRepo(context.Background())
		Ω(err).Should(Succeed())
	case docker_registry.QuayImplementationName:
		SuiteData.Stubs.SetEnv("WERF_PARALLEL", "0")
	default:
	}
}

func implementationAfterEach(implementationName string) {
	switch implementationName {
	case docker_registry.AzureCrImplementationName, docker_registry.AwsEcrImplementationName, docker_registry.DockerHubImplementationName, docker_registry.GitHubPackagesImplementationName, docker_registry.HarborImplementationName, docker_registry.QuayImplementationName:
		err := SuiteData.StagesStorage.DeleteRepo(context.Background())
		switch err := err.(type) {
		case nil, docker_registry.AzureCrNotFoundError, docker_registry.DockerHubNotFoundError, docker_registry.HarborNotFoundError, docker_registry.QuayNotFoundError:
		default:
			Ω(err).Should(Succeed())
		}
	default:
	}
}
