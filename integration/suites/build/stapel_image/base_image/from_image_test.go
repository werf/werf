package base_image_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
	utilsDocker "github.com/werf/werf/test/pkg/utils/docker"
)

var fromImageItFunc = func(appConfigName, fromImageConfigName string, extraAfterBuildChecks func(appConfigName, fromImageConfigName string)) {
	By(fmt.Sprintf("fromCacheVersion: %s", "0"))
	SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "0")

	output := utils.SucceedCommandOutputString(
		SuiteData.TestDirPath,
		SuiteData.WerfBinPath,
		"build",
	)

	Ω(strings.Count(output, fmt.Sprintf("Building stage %s/from", appConfigName))).Should(Equal(2))

	extraAfterBuildChecks(appConfigName, fromImageConfigName)

	By(fmt.Sprintf("fromCacheVersion: %s", "1"))
	SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")

	output = utils.SucceedCommandOutputString(
		SuiteData.TestDirPath,
		SuiteData.WerfBinPath,
		"build",
	)

	Ω(strings.Count(output, fmt.Sprintf("Building stage %s/from", appConfigName))).Should(Equal(2))

	extraAfterBuildChecks(appConfigName, fromImageConfigName)
}

var _ = XDescribe("fromImage", func() {
	BeforeEach(func() {
		SuiteData.TestDirPath = utils.FixturePath("from_image")
	})

	It("should be rebuilt", func() {
		fromImageItFunc("app", "fromImage", func(appConfigName, fromImageConfigName string) {
			appImageName := utils.SucceedCommandOutputString(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"stage", "image", appConfigName,
			)

			fromImageName := utils.SucceedCommandOutputString(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"stage", "image", fromImageConfigName,
			)

			Ω(utilsDocker.ImageParent(strings.TrimSpace(appImageName))).Should(Equal(utilsDocker.ImageID(strings.TrimSpace(fromImageName))))
		})
	})
})
