package cleanup_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/prashantv/gostub"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/storage"

	"github.com/flant/werf/pkg/testing/utils"
	utilsDocker "github.com/flant/werf/pkg/testing/utils/docker"
)

// Environment implementation variables
// WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_<implementation name>
// WERF_TEST_<implementation name>_REGISTRY
//
// WERF_TEST_DOCKERHUB_USERNAME
// WERF_TEST_DOCKERHUB_PASSWORD

func TestIntegration(t *testing.T) {
	if !utils.MeetsRequirements(requiredSuiteTools, requiredSuiteEnvs) {
		fmt.Println("Missing required tools")
		os.Exit(1)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Cleanup Suite")
}

var requiredSuiteTools = []string{"git", "docker"}
var requiredSuiteEnvs []string

var tmpDir string
var testDirPath string
var werfBinPath string
var stubs = gostub.New()
var localImagesRepoAddress, localImagesRepoContainerName string

var stagesStorage storage.StagesStorage
var imagesRepo storage.ImagesRepo

var _ = SynchronizedBeforeSuite(func() []byte {
	computedPathToWerf := utils.ProcessWerfBinPath()
	return []byte(computedPathToWerf)
}, func(computedPathToWerf []byte) {
	werfBinPath = string(computedPathToWerf)
	localImagesRepoAddress, localImagesRepoContainerName = utilsDocker.LocalDockerRegistryRun()
})

var _ = SynchronizedAfterSuite(func() {
	utilsDocker.ContainerStopAndRemove(localImagesRepoContainerName)
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	tmpDir = utils.GetTempDir()
	testDirPath = tmpDir

	utils.BeforeEachOverrideWerfProjectName(stubs)
})

var _ = AfterEach(func() {
	err := os.RemoveAll(tmpDir)
	Ω(err).ShouldNot(HaveOccurred())

	stubs.Reset()
})

func forEachDockerRegistryImplementation(description string, body func()) bool {
	for _, name := range implementationListToCheck() {
		implementationName := name

		Describe(fmt.Sprintf("%s (%s)", description, implementationName), func() {
			BeforeEach(func() {
				var stagesStorageAddress string
				var stagesStorageImplementationName string
				var stagesStorageDockerRegistryOptions docker_registry.DockerRegistryOptions

				var imagesRepoAddress string
				var imagesRepoMode string
				var imagesRepoImplementationName string
				var imagesRepoDockerRegistryOptions docker_registry.DockerRegistryOptions

				if implementationName == ":local" {
					stagesStorageAddress = ":local"
					stagesStorageImplementationName = ""
					imagesRepoAddress = strings.Join([]string{localImagesRepoAddress, utils.ProjectName()}, "/")
					imagesRepoMode = storage.MultirepoImagesRepoMode
					imagesRepoImplementationName = ""
				} else {
					stagesStorageAddress = implementationStagesStorageAddress(implementationName)
					stagesStorageImplementationName = "" // TODO
					stagesStorageDockerRegistryOptions = implementationStagesStorageDockerRegistryOptions(implementationName)

					imagesRepoAddress = implementationImagesRepoAddress(implementationName)
					imagesRepoMode = implementationImagesRepoMode(implementationName)
					imagesRepoImplementationName = implementationName
					imagesRepoDockerRegistryOptions = implementationImagesRepoDockerRegistryOptions(implementationName)
				}

				initStagesStorage(stagesStorageAddress, stagesStorageImplementationName, stagesStorageDockerRegistryOptions)
				initImagesRepo(imagesRepoAddress, imagesRepoMode, imagesRepoImplementationName, imagesRepoDockerRegistryOptions)

				stubs.SetEnv("WERF_STAGES_STORAGE", stagesStorageAddress)
				stubs.SetEnv("WERF_IMAGES_REPO", imagesRepoAddress)
				stubs.SetEnv("WERF_IMAGES_REPO_MODE", imagesRepoMode) // TODO
			})

			AfterEach(func() {
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"purge", "--force",
				)

				implementationAfterEach(implementationName)
			})

			body()
		})
	}

	return true
}

func initImagesRepo(imagesRepoAddress, imageRepoMode, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) {
	projectName := utils.ProjectName()

	i, err := storage.NewImagesRepo(
		projectName,
		imagesRepoAddress,
		imageRepoMode,
		storage.ImagesRepoOptions{
			DockerImagesRepoOptions: storage.DockerImagesRepoOptions{
				DockerRegistryOptions: dockerRegistryOptions,
				Implementation:        implementationName,
			},
		},
	)
	Ω(err).ShouldNot(HaveOccurred())

	imagesRepo = i
}

func initStagesStorage(stagesStorageAddress string, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) {
	s, err := storage.NewStagesStorage(
		stagesStorageAddress,
		&container_runtime.LocalDockerServerRuntime{},
		storage.StagesStorageOptions{
			RepoStagesStorageOptions: storage.RepoStagesStorageOptions{
				DockerRegistryOptions: dockerRegistryOptions,
				Implementation:        implementationName,
			},
		},
	)
	Ω(err).ShouldNot(HaveOccurred())

	stagesStorage = s
}

func imagesRepoAllImageRepoTags(imageName string) []string {
	tags, err := imagesRepo.GetAllImageRepoTags(imageName)
	Ω(err).ShouldNot(HaveOccurred())
	return tags
}

func stagesStorageRepoImagesCount() int {
	repoImages, err := stagesStorage.GetRepoImages(utils.ProjectName())
	Ω(err).ShouldNot(HaveOccurred())
	return len(repoImages)
}

func stagesStorageManagedImagesCount() int {
	managedImages, err := stagesStorage.GetManagedImages(utils.ProjectName())
	Ω(err).ShouldNot(HaveOccurred())
	return len(managedImages)
}

func implementationListToCheck() []string {
	list := []string{":local"}

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

	return list
}

func implementationStagesStorageAddress(_ string) string {
	return ":local" // TODO
}

func implementationStagesStorageDockerRegistryOptions(_ string) docker_registry.DockerRegistryOptions {
	return docker_registry.DockerRegistryOptions{} // TODO
}

func implementationImagesRepoAddress(implementationName string) string {
	projectName := utils.ProjectName()
	implementationCode := strings.ToUpper(implementationName)

	registryEnvName := fmt.Sprintf(
		"WERF_TEST_%s_REGISTRY",
		implementationCode,
	)

	registry := getRequiredEnv(registryEnvName)

	return strings.Join([]string{registry, projectName}, "/")
}

func implementationImagesRepoMode(implementationName string) string {
	switch implementationName {
	case docker_registry.DockerHubImplementationName, docker_registry.GitHubPackagesImplementationName, docker_registry.QuayImplementationName:
		return storage.MonorepoImagesRepoMode
	default:
		return storage.MultirepoImagesRepoMode
	}
}

func implementationImagesRepoDockerRegistryOptions(implementationName string) docker_registry.DockerRegistryOptions {
	implementationCode := strings.ToUpper(implementationName)

	switch implementationName {
	case docker_registry.DockerHubImplementationName:
		usernameEnvName := fmt.Sprintf(
			"WERF_TEST_%s_USERNAME",
			implementationCode,
		)

		passwordEnvName := fmt.Sprintf(
			"WERF_TEST_%s_PASSWORD",
			implementationCode,
		)

		username := getRequiredEnv(usernameEnvName)
		password := getRequiredEnv(passwordEnvName)

		stubs.SetEnv("WERF_REPO_DOCKER_HUB_USERNAME", username)
		stubs.SetEnv("WERF_REPO_DOCKER_HUB_PASSWORD", password)

		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
			DockerHubUsername:     username,
			DockerHubPassword:     password,
		}
	default:
		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
		}
	}
}

func implementationAfterEach(implementationName string) {
	switch implementationName {
	case docker_registry.DockerHubImplementationName:
		err := imagesRepo.DeleteRepo()
		if err != nil {
			if _, ok := err.(docker_registry.DockerHubNotFoundError); !ok {
				Ω(err).Should(Succeed())
			}
		}
	}
}

func getRequiredEnv(name string) string {
	envValue := os.Getenv(name)
	if envValue == "" {
		panic(fmt.Sprintf("environment variable %s must be specified", name))
	}

	return envValue
}
