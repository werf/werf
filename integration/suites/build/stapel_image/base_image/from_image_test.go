package base_image_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
	utilsDocker "github.com/werf/werf/v2/test/pkg/utils/docker"
)

var fromImageItFunc = func(ctx SpecContext, appConfigName, fromImageConfigName string, extraAfterBuildChecks func(appConfigName, fromImageConfigName string)) {
	By(fmt.Sprintf("fromCacheVersion: %s", "0"))
	SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "0")

	output := utils.SucceedCommandOutputString(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

	Expect(strings.Count(output, fmt.Sprintf("Building stage %s/from", appConfigName))).Should(Equal(2))

	extraAfterBuildChecks(appConfigName, fromImageConfigName)

	By(fmt.Sprintf("fromCacheVersion: %s", "1"))
	SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")

	output = utils.SucceedCommandOutputString(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

	Expect(strings.Count(output, fmt.Sprintf("Building stage %s/from", appConfigName))).Should(Equal(2))

	extraAfterBuildChecks(appConfigName, fromImageConfigName)
}

var _ = XDescribe("fromImage", func() {
	BeforeEach(func() {
		SuiteData.TestDirPath = utils.FixturePath("from_image")
	})

	It("should be rebuilt", func(ctx SpecContext) {
		fromImageItFunc(ctx, "app", "fromImage", func(appConfigName, fromImageConfigName string) {
			appImageName := utils.SucceedCommandOutputString(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "stage", "image", appConfigName)

			fromImageName := utils.SucceedCommandOutputString(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "stage", "image", fromImageConfigName)

			Expect(utilsDocker.ImageParent(strings.TrimSpace(appImageName))).Should(Equal(utilsDocker.ImageID(strings.TrimSpace(fromImageName))))
		})
	})
})

var _ = XDescribe("from anywhere", func() {
	BeforeEach(func() {
		SuiteData.TestDirPath = utils.FixturePath("from_anywhere")
	})

	It("should resolve and chain correctly", func(ctx SpecContext) {
		trimID := func(imageName string) string {
			return strings.TrimSpace(utils.SucceedCommandOutputString(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "stage", "image", imageName))
		}

		externalImageID := trimID("FromExternalImage")
		fromImageID := trimID("FromImage")
		fromImageAliasID := trimID("FromImageAlias")

		Expect(utilsDocker.ImageParent(fromImageID)).Should(Equal(utilsDocker.ImageID(externalImageID)))

		Expect(utilsDocker.ImageParent(fromImageAliasID)).Should(Equal(utilsDocker.ImageID(fromImageID)))
	})
})
