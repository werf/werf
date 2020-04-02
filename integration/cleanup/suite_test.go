package cleanup_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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
var testImplementation string
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
				testImplementation = implementationName

				var stagesStorageAddress string
				var stagesStorageImplementationName string
				var stagesStorageDockerRegistryOptions docker_registry.DockerRegistryOptions

				var imagesRepoAddress string
				var imagesRepoMode string
				var imagesRepoImplementationName string
				var imagesRepoDockerRegistryOptions docker_registry.DockerRegistryOptions

				if implementationName == ":local" {
					stagesStorageAddress = ":local"

					imagesRepoAddress = strings.Join([]string{localImagesRepoAddress, utils.ProjectName()}, "/")
					imagesRepoMode = docker_registry.MultirepoRepoMode
				} else {
					stagesStorageAddress = implementationStagesStorageAddress(implementationName)
					stagesStorageDockerRegistryOptions = implementationStagesStorageDockerRegistryOptions(implementationName)

					imagesRepoAddress = implementationImagesRepoAddress(implementationName)
					imagesRepoMode = "auto"
					imagesRepoDockerRegistryOptions = implementationImagesRepoDockerRegistryOptions(implementationName)
				}

				isNotSupported := true
				for _, name := range docker_registry.ImplementationList() {
					if name == implementationName {
						isNotSupported = false
					}
				}

				if isNotSupported {
					imagesRepoImplementationName = "default"
					// TODO: stagesStorageImplementationName = "default"
					// TODO: stubs.SetEnv("WERF_STAGES_STORAGE_IMPLEMENTATION", imagesRepoImplementationName)
					stubs.SetEnv("WERF_IMAGES_REPO_IMPLEMENTATION", "default")
				}

				initStagesStorage(stagesStorageAddress, stagesStorageImplementationName, stagesStorageDockerRegistryOptions)
				initImagesRepo(imagesRepoAddress, imagesRepoMode, imagesRepoImplementationName, imagesRepoDockerRegistryOptions)

				stubs.SetEnv("WERF_STAGES_STORAGE", stagesStorageAddress)
				stubs.SetEnv("WERF_IMAGES_REPO", imagesRepoAddress)
			})

			BeforeEach(func() {
				implementationBeforeEach(implementationName)
			})

			AfterEach(func() {
			afterEach:
				combinedOutput, err := utils.RunCommand(
					testDirPath,
					werfBinPath,
					"purge", "--force",
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
	repoImages, err := stagesStorage.GetAllStages(utils.ProjectName())
	Ω(err).ShouldNot(HaveOccurred())
	return len(repoImages)
}

func stagesStorageManagedImagesCount() int {
	managedImages, err := stagesStorage.GetManagedImages(utils.ProjectName())
	Ω(err).ShouldNot(HaveOccurred())
	return len(managedImages)
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
		return []string{":local"}
	}
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

func implementationImagesRepoDockerRegistryOptions(implementationName string) docker_registry.DockerRegistryOptions {
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
	case docker_registry.GitHubPackagesImplementationName:
		token := getRequiredEnv(tokenEnvName)

		stubs.SetEnv("WERF_REPO_GITHUB_TOKEN", token)

		return docker_registry.DockerRegistryOptions{
			InsecureRegistry:      false,
			SkipTlsVerifyRegistry: false,
			GitHubToken:           token,
		}
	case docker_registry.HarborImplementationName:
		username := getRequiredEnv(usernameEnvName)
		password := getRequiredEnv(passwordEnvName)

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
			QuayToken:             getRequiredEnv(tokenEnvName),
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
		err := imagesRepo.CreateImageRepo("image")
		Ω(err).Should(Succeed())
	default:
	}
}

func implementationAfterEach(implementationName string) {
	switch implementationName {
	case docker_registry.AwsEcrImplementationName, docker_registry.DockerHubImplementationName, docker_registry.GitHubPackagesImplementationName, docker_registry.HarborImplementationName, docker_registry.QuayImplementationName:
		if implementationName == docker_registry.HarborImplementationName {
			// API cannot delete repository without any tags
			// {"code":404,"message":"no tags found for repository test2/werf-test-none-7872-wfdy8uyupu/image"}

			Ω(utilsDocker.Pull("flant/werf-test:hello-world")).Should(Succeed(), "docker pull")
			Ω(utilsDocker.CliTag("flant/werf-test:hello-world", imagesRepo.ImageRepositoryName("image"))).Should(Succeed(), "docker tag")
			defer func() {
				Ω(utilsDocker.CliRmi(imagesRepo.ImageRepositoryName("image"))).Should(Succeed(), "docker rmi")
			}()

			Ω(utilsDocker.CliPush(imagesRepo.ImageRepositoryName("image"))).Should(Succeed(), "docker push")
		}

		err := imagesRepo.DeleteImageRepo("image")
		switch err := err.(type) {
		case nil, docker_registry.DockerHubNotFoundError, docker_registry.HarborNotFoundError, docker_registry.QuayNotFoundError:
		default:
			Ω(err).Should(Succeed())
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
