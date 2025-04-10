package managed_images_test

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("managed images", func() {
	BeforeEach(func(ctx SpecContext) {
		SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("default"), "initial commit")
	})

	for _, iName := range suite_init.ContainerRegistryImplementationListToCheck(true) {
		implementationName := iName

		Context("["+implementationName+"]", func() {
			BeforeEach(func(ctx SpecContext) {
				repo := fmt.Sprintf("%s/%s", SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryAddress, SuiteData.ProjectName)
				SuiteData.SetupRepo(ctx, repo, implementationName, SuiteData.StubsData)
			})

			It("ls should not return anything", func(ctx SpecContext) {
				output := utils.SucceedCommandOutputString(
					ctx,
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					SuiteData.WerfBinPath,
					"managed-images", "ls",
				)

				Expect(output).Should(BeEmpty())
			})

			It("add should work properly", func(ctx SpecContext) {
				addManagedImage(ctx, "test")
				Expect(isManagedImage(ctx, "test")).Should(BeTrue())
			})

			When("managed-images test has been added", func() {
				managedImage := "test"

				BeforeEach(func(ctx SpecContext) {
					addManagedImage(ctx, managedImage)
				})

				It("ls should return managed image", func(ctx SpecContext) {
					Expect(isManagedImage(ctx, managedImage)).Should(BeTrue())
				})

				It("rm should remove managed-image", func(ctx SpecContext) {
					rmManagedImage(ctx, managedImage)
					Expect(isManagedImage(ctx, managedImage)).Should(BeFalse())
				})
			})

			When("werf images have been built", func() {
				BeforeEach(func(ctx SpecContext) {
					utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build")
				})

				It("ls should return managed image", func(ctx SpecContext) {
					Expect(isManagedImage(ctx, "a")).Should(BeTrue())
					Expect(isManagedImage(ctx, "b")).Should(BeTrue())
					Expect(isManagedImage(ctx, "c")).Should(BeTrue())
					Expect(isManagedImage(ctx, "d")).Should(BeTrue())
				})
			})
		})
	}
})

func addManagedImage(ctx context.Context, imageName string) {
	utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "managed-images", "add", imageName)
}

func rmManagedImage(ctx context.Context, imageName string) {
	utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "managed-images", "rm", imageName)
}

func isManagedImage(ctx context.Context, imageName string) bool {
	output := utils.SucceedCommandOutputString(
		ctx,
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
