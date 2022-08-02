package managed_images_test

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/suite_init"
	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("managed images", func() {
	BeforeEach(func() {
		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("default"), "initial commit")
	})

	for _, iName := range suite_init.ContainerRegistryImplementationListToCheck(true) {
		implementationName := iName

		Context("["+implementationName+"]", func() {
			BeforeEach(func() {
				repo := fmt.Sprintf("%s/%s", SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryAddress, SuiteData.ProjectName)
				SuiteData.SetupRepo(context.Background(), repo, implementationName, SuiteData.StubsData)
			})

			It("ls should not return anything", func() {
				output := utils.SucceedCommandOutputString(
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
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
						SuiteData.GetProjectWorktree(SuiteData.ProjectName),
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
	}
})

func addManagedImage(imageName string) {
	utils.RunSucceedCommand(
		SuiteData.GetProjectWorktree(SuiteData.ProjectName),
		SuiteData.WerfBinPath,
		"managed-images", "add", imageName,
	)
}

func rmManagedImage(imageName string) {
	utils.RunSucceedCommand(
		SuiteData.GetProjectWorktree(SuiteData.ProjectName),
		SuiteData.WerfBinPath,
		"managed-images", "rm", imageName,
	)
}

func isManagedImage(imageName string) bool {
	output := utils.SucceedCommandOutputString(
		SuiteData.GetProjectWorktree(SuiteData.ProjectName),
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
