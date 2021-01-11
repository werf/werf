package cleanup_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/pkg/docker_registry"
)

var _ = forEachDockerRegistryImplementation("purge command", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("purge"), SuiteData.TestDirPath)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"init",
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"add", "werf.yaml",
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"commit", "-m", "Initial commit",
		)
	})

	It("should remove all project data", func() {
		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			SuiteData.WerfBinPath,
			"build",
		)

		Ω(StagesCount()).Should(BeNumerically(">", 0))
		Ω(ManagedImagesCount()).Should(BeNumerically(">", 0))
		Ω(len(ImageMetadata(imageName))).Should(BeNumerically(">", 0))
		Ω(len(ImportMetadataIDs())).Should(BeNumerically(">", 0))

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			SuiteData.WerfBinPath,
			"purge",
		)

		if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
			Ω(StagesCount()).Should(Equal(0))
			Ω(ManagedImagesCount()).Should(Equal(0))
			Ω(len(ImageMetadata(imageName))).Should(Equal(0))
			Ω(len(ImportMetadataIDs())).Should(Equal(0))
		}
	})

	Context("when there is running container based on werf image", func() {
		BeforeEach(func() {
			if SuiteData.StagesStorage.Address() != ":local" {
				Skip(fmt.Sprintf("to test :local storage (%s)", SuiteData.StagesStorage.Address()))
			}
		})

		BeforeEach(func() {
			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"build",
			)

			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"run", "--docker-options", "-d", "--", "/bin/sleep", "30",
			)
		})

		It("should fail with specific error", func() {
			out, err := utils.RunCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"purge",
			)
			Ω(err).ShouldNot(Succeed())
			Ω(string(out)).Should(ContainSubstring("used by container"))
			Ω(string(out)).Should(ContainSubstring("Use --force option to remove all containers that are based on deleting werf docker images"))
		})

		It("should remove project images and related container", func() {
			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"purge", "--force",
			)

			Ω(StagesCount()).Should(Equal(0))
		})
	})
})
