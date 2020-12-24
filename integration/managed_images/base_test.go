package managed_images_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/utils"
)

var _ = Describe("managed images", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("default"), SuiteData.TestDirPath)
	})

	It("ls should not return anything", func() {
		output := utils.SucceedCommandOutputString(
			SuiteData.TestDirPath,
			SuiteData.WerfBinPath,
			"managed-images", "ls",
		)

		Ω(output).Should(BeEmpty())
	})

	It("add should work properly", func() {
		addManagedImage("test")
		Ω(isManagedImage("test")).Should(BeTrue())
	})

	When("managed-images test has been added", func() {
		managedImage := "test"

		BeforeEach(func() {
			addManagedImage(managedImage)
		})

		It("ls should return managed image", func() {
			Ω(isManagedImage(managedImage)).Should(BeTrue())
		})

		It("rm should remove managed-image", func() {
			rmManagedImage(managedImage)
			Ω(isManagedImage(managedImage)).Should(BeFalse())
		})
	})

	When("werf images have been built", func() {
		BeforeEach(func() {
			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"build",
			)
		})

		It("ls should return managed image", func() {
			Ω(isManagedImage("a")).Should(BeTrue())
			Ω(isManagedImage("b")).Should(BeTrue())
			Ω(isManagedImage("c")).Should(BeTrue())
			Ω(isManagedImage("d")).Should(BeFalse())
		})
	})
})

func addManagedImage(imageName string) {
	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		SuiteData.WerfBinPath,
		"managed-images", "add", imageName,
	)
}

func rmManagedImage(imageName string) {
	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		SuiteData.WerfBinPath,
		"managed-images", "rm", imageName,
	)
}

func isManagedImage(imageName string) bool {
	output := utils.SucceedCommandOutputString(
		SuiteData.TestDirPath,
		SuiteData.WerfBinPath,
		"managed-images", "ls",
	)

	for _, managedImage := range strings.Fields(output) {
		if managedImage == imageName {
			return true
		}
	}

	return false
}
