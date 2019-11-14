// +build integration

package base_image

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/integration/utils"
	utilsDocker "github.com/flant/werf/integration/utils/docker"
)

var fromImageItFunc = func(appConfigName, fromImageConfigName string, extraAfterBuildChecks func(appConfigName, fromImageConfigName string)) {
	By(fmt.Sprintf("fromCacheVersion: %s", "0"))
	Ω(os.Setenv("FROM_CACHE_VERSION", "0")).Should(Succeed())

	output := utils.SucceedCommandOutput(
		testDirPath,
		werfBinPath,
		"build",
	)

	Ω(strings.Count(output, "Building stage from")).Should(Equal(4))

	extraAfterBuildChecks(appConfigName, fromImageConfigName)

	By(fmt.Sprintf("fromCacheVersion: %s", "1"))
	Ω(os.Setenv("FROM_CACHE_VERSION", "1")).Should(Succeed())

	output = utils.SucceedCommandOutput(
		testDirPath,
		werfBinPath,
		"build",
	)

	Ω(strings.Count(output, "Building stage from")).Should(Equal(4))

	extraAfterBuildChecks(appConfigName, fromImageConfigName)
}

var _ = Describe("fromImage", func() {
	BeforeEach(func() {
		testDirPath = fixturePath("from_image")
	})

	It("should be rebuilt", func() {
		fromImageItFunc("app", "fromImage", func(appConfigName, fromImageConfigName string) {
			appImageName := utils.SucceedCommandOutput(
				testDirPath,
				werfBinPath,
				"stage", "image", appConfigName,
			)

			fromImageName := utils.SucceedCommandOutput(
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
		testDirPath = fixturePath("from_image_artifact")
	})

	It("should be rebuilt", func() {
		fromImageItFunc("app", "fromImageArtifact", func(appConfigName, fromImageConfigName string) {})
	})
})
