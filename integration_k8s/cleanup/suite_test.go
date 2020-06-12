package cleanup_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/prashantv/gostub"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/testing/utils"
)

func TestIntegration(t *testing.T) {
	if !utils.MeetsRequirements(requiredSuiteTools, requiredSuiteEnvs) {
		fmt.Println("Missing required tools")
		os.Exit(1)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Cleanup Suite")
}

var requiredSuiteTools = []string{"git", "docker"}
var requiredSuiteEnvs = []string{
	"WERF_TEST_K8S_DOCKER_REGISTRY",
	"WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME",
	"WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD",
}

var tmpDir string
var testDirPath string
var werfBinPath string
var stubs = gostub.New()

var stagesStorage storage.StagesStorage
var imagesRepo storage.ImagesRepo

var _ = SynchronizedBeforeSuite(func() []byte {
	computedPathToWerf := utils.ProcessWerfBinPath()
	return []byte(computedPathToWerf)
}, func(computedPathToWerf []byte) {
	werfBinPath = string(computedPathToWerf)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	tmpDir = utils.GetTempDir()
	testDirPath = tmpDir

	utils.BeforeEachOverrideWerfProjectName(stubs)

	imagesRepoAddress := fmt.Sprintf("%s/%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), utils.ProjectName())
	imagesRepo = utils.NewImagesRepo(imagesRepoAddress, "multirepo", "default", docker_registry.DockerRegistryOptions{})

	stagesStorageRepoAddress := fmt.Sprintf("%s/%s/%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), utils.ProjectName(), "stages")
	stagesStorage = utils.NewStagesStorage(stagesStorageRepoAddress, "default", docker_registry.DockerRegistryOptions{})

	stubs.SetEnv("WERF_STAGES_STORAGE", stagesStorageRepoAddress)
	stubs.SetEnv("WERF_IMAGES_REPO", imagesRepoAddress)
})

var _ = AfterEach(func() {
	utils.RunSucceedCommand(
		testDirPath,
		werfBinPath,
		"purge", "--force",
	)

	err := os.RemoveAll(tmpDir)
	Ω(err).ShouldNot(HaveOccurred())

	stubs.Reset()
})

func imagesRepoAllImageRepoTags(imageName string) []string {
	return utils.ImagesRepoAllImageRepoTags(imagesRepo, imageName)
}

func stagesStorageRepoImagesCount() int {
	return utils.StagesStorageRepoImagesCount(stagesStorage)
}
