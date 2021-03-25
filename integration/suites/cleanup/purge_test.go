package cleanup_test

import (
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
})
