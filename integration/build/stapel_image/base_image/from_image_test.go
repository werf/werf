package base_image_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
	utilsDocker "github.com/flant/werf/pkg/testing/utils/docker"
)

var fromImageItFunc = func(appConfigName, fromImageConfigName string, extraAfterBuildChecks func(appConfigName, fromImageConfigName string)) {
	By(fmt.Sprintf("fromCacheVersion: %s", "0"))
	stubs.SetEnv("FROM_CACHE_VERSION", "0")

	output := utils.SucceedCommandOutputString(
		testDirPath,
		werfBinPath,
		"build",
	)

	Ω(strings.Count(output, "Building stage from")).Should(Equal(4))

	extraAfterBuildChecks(appConfigName, fromImageConfigName)

	By(fmt.Sprintf("fromCacheVersion: %s", "1"))
	stubs.SetEnv("FROM_CACHE_VERSION", "1")

	output = utils.SucceedCommandOutputString(
		testDirPath,
		werfBinPath,
		"build",
	)

	Ω(strings.Count(output, "Building stage from")).Should(Equal(4))

	extraAfterBuildChecks(appConfigName, fromImageConfigName)
}

var _ = Describe("fromImage", func() {
	BeforeEach(func() {
		testDirPath = utils.FixturePath("from_image")
	})

	It("should be rebuilt", func() {
		fromImageItFunc("app", "fromImage", func(appConfigName, fromImageConfigName string) {
			appImageName := utils.SucceedCommandOutputString(
				testDirPath,
				werfBinPath,
				"stage", "image", appConfigName,
			)

			fromImageName := utils.SucceedCommandOutputString(
				testDirPath,
				werfBinPath,
				"stage", "image", fromImageConfigName,
			)

			Ω(utilsDocker.ImageParent(strings.TrimSpace(appImageName))).Should(Equal(utilsDocker.ImageID(strings.TrimSpace(fromImageName))))
		})
	})
})

var _ = Describe("fromImageArtifact", func() {
	BeforeEach(func() {
		testDirPath = utils.FixturePath("from_image_artifact")
	})

	It("should be rebuilt", func() {
		fromImageItFunc("app", "fromImageArtifact", func(appConfigName, fromImageConfigName string) {})
	})
})
