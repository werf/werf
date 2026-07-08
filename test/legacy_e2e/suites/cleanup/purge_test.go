package cleanup_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("purge command", func() {
	for _, iName := range suite_init.ContainerRegistryImplementationListToCheck(true) {
		implementationName := iName

		Context("["+implementationName+"]", func() {
			BeforeEach(perImplementationBeforeEach(implementationName))

			BeforeEach(func() {
				SuiteData.Stubs.SetEnv("WERF_WITHOUT_KUBE", "1")
			})

			BeforeEach(func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("purge"), SuiteData.TestDirPath)

				utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "init")

				utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "add", "werf.yaml", "werf-giterminism.yaml")

				utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "-m", "Initial commit")
			})

			It("should remove all project data", func(ctx SpecContext) {
				utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build", "--add-custom-tag", fmt.Sprintf(customTagValueFormat, "1"))

				Expect(StagesCount(ctx)).Should(BeNumerically(">", 0))
				Expect(ManagedImagesCount(ctx)).Should(BeNumerically(">", 0))
				Expect(len(ImageMetadata(ctx, imageName))).Should(BeNumerically(">", 0))
				Expect(len(CustomTags(ctx))).Should(BeNumerically(">", 0))
				Expect(len(CustomTagsMetadataList(ctx))).Should(BeNumerically(">", 0))

				utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "purge")

				if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
					Expect(StagesCount(ctx)).Should(Equal(0))
					Expect(ManagedImagesCount(ctx)).Should(Equal(0))
					Expect(len(ImageMetadata(ctx, imageName))).Should(Equal(0))
					Expect(len(CustomTags(ctx))).Should(Equal(0))
					Expect(len(CustomTagsMetadataList(ctx))).Should(Equal(0))
				}
			})

			Context("with --meta-repo", func() {
				BeforeEach(func(ctx SpecContext) {
					SetupMetaRepo(ctx, implementationName)
				})

				It("should remove managed-image, image-metadata and custom-tag metadata from meta-repo on purge", func(ctx SpecContext) {
					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build", "--add-custom-tag", fmt.Sprintf(customTagValueFormat, "1"))

					Expect(StagesCount(ctx)).Should(BeNumerically(">", 0))
					Expect(MetaManagedImagesCount(ctx)).Should(BeNumerically(">", 0), "managed-image records land in meta-repo")
					Expect(ManagedImagesCount(ctx)).Should(Equal(0), "stages-repo MUST NOT contain managed-image records")
					Expect(len(MetaImageMetadata(ctx, imageName))).Should(BeNumerically(">", 0), "image-metadata records land in meta-repo")
					Expect(len(ImageMetadata(ctx, imageName))).Should(Equal(0), "stages-repo MUST NOT contain image-metadata records")
					Expect(len(CustomTags(ctx))).Should(BeNumerically(">", 0), "custom-tag alias images live in stages-repo")
					Expect(len(MetaCustomTagsMetadataList(ctx))).Should(BeNumerically(">", 0), "custom-tag metadata records land in meta-repo")
					Expect(len(CustomTagsMetadataList(ctx))).Should(Equal(0), "stages-repo MUST NOT contain custom-tag metadata")

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "purge")

					if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
						Expect(StagesCount(ctx)).Should(Equal(0))
						Expect(MetaManagedImagesCount(ctx)).Should(Equal(0), "managed-image records purged from meta-repo")
						Expect(len(MetaImageMetadata(ctx, imageName))).Should(Equal(0), "image-metadata records purged from meta-repo")
						Expect(len(CustomTags(ctx))).Should(Equal(0), "custom-tag alias images removed from stages-repo")
						Expect(len(MetaCustomTagsMetadataList(ctx))).Should(Equal(0), "custom-tag metadata records purged from meta-repo")
					}
				})
			})
		})
	}
})
