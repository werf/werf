package cleanup_with_k8s_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("cleanup command", func() {
	const (
		artifactCacheVersion1             = "1"
		artifactCacheVersion2             = "2"
		artifactData1                     = "1"
		artifactData2                     = "2"
		expectedStageCountAfterFirstBuild = 5
	)

	setImageCredentialsEnv := func() {
		SuiteData.Stubs.SetEnv("WERF_SET_IMAGE_CREDENTIALS_REGISTRY", fmt.Sprintf("imageCredentials.registry=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY")))
		SuiteData.Stubs.SetEnv("WERF_SET_IMAGE_CREDENTIALS_USERNAME", fmt.Sprintf("imageCredentials.username=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME")))
		SuiteData.Stubs.SetEnv("WERF_SET_IMAGE_CREDENTIALS_PASSWORD", fmt.Sprintf("imageCredentials.password=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD")))
	}

	setupProject := func() {
		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("default"), "initial commit")
		setImageCredentialsEnv()
	}

	runCommand := func(args ...string) {
		utils.RunSucceedCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, args...)
	}

	BeforeEach(func() {
		setupProject()
	})

	AfterEach(func() {
		runCommand("dismiss", "--with-namespace")
		runCommand("purge")
	})

	Context("kubernetes based policy", func() {
		BeforeEach(func() {
			SuiteData.Stubs.SetEnv("ARTIFACT_CACHE_VERSION", artifactCacheVersion1)
			SuiteData.Stubs.SetEnv("ARTIFACT_DATA", artifactData1)
			runCommand("build")
		})

		It("should remove all stages", func() {
			Expect(StagesCount()).Should(Equal(expectedStageCountAfterFirstBuild))
			runCommand("cleanup")
			Expect(StagesCount()).Should(Equal(0))
		})

		It("should keep all stages", func() {
			runCommand("converge", "--require-built-images")
			Expect(StagesCount()).Should(Equal(expectedStageCountAfterFirstBuild))
			runCommand("cleanup")
			Expect(StagesCount()).Should(Equal(expectedStageCountAfterFirstBuild))
		})

		Context("artifact", func() {
			BeforeEach(func() {
				SuiteData.Stubs.SetEnv("ARTIFACT_CACHE_VERSION", artifactCacheVersion2)
			})

			It("should keep both by import checksum", func() {
				SuiteData.Stubs.SetEnv("ARTIFACT_DATA", artifactData1)
				runCommand("converge")

				Expect(StagesCount()).Should(Equal(expectedStageCountAfterFirstBuild + 2))
				Expect(len(ImportMetadataIDs())).Should(Equal(2))

				runCommand("cleanup")

				Expect(StagesCount()).Should(Equal(expectedStageCountAfterFirstBuild + 2))
				Expect(len(ImportMetadataIDs())).Should(Equal(2))
			})

			It("should keep one", func() {
				SuiteData.Stubs.SetEnv("ARTIFACT_DATA", artifactData2)
				runCommand("converge")

				Expect(StagesCount()).Should(Equal(expectedStageCountAfterFirstBuild + 3))
				Expect(len(ImportMetadataIDs())).Should(Equal(2))

				runCommand("cleanup")

				Expect(StagesCount()).Should(Equal(expectedStageCountAfterFirstBuild))
				Expect(len(ImportMetadataIDs())).Should(Equal(1))
			})
		})
	})
})
