package cleanup_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/testing/utils"
)

var _ = forEachDockerRegistryImplementation("purging stages", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("default"), testDirPath)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"init",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"add", "werf.yaml",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"commit", "-m", "Initial commit",
		)
	})

	for _, werfCommand := range [][]string{
		{"stages", "purge"},
		{"purge"},
	} {
		description := strings.Join(werfCommand, " ") + " command"
		werfCommand := werfCommand

		Describe(description, func() {
			It("should remove project images", func() {
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				Ω(stagesStorageRepoImagesCount()).Should(BeNumerically(">", 0))
				Ω(stagesStorageManagedImagesCount()).Should(BeNumerically(">", 0))

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					werfCommand...,
				)

				Ω(stagesStorageRepoImagesCount()).Should(Equal(0))
				Ω(stagesStorageManagedImagesCount()).Should(Equal(0))
			})

			Context("when there is running container based on werf image", func() {
				BeforeEach(func() {
					if stagesStorage.Address() != ":local" {
						Skip(fmt.Sprintf("to test :local stages storage (%s)", stagesStorage.Address()))
					}
				})

				BeforeEach(func() {
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"run", "--docker-options", "-d", "--", "/bin/sleep", "30",
					)

					stubs.SetEnv("WERF_LOG_PRETTY", "0")
				})

				It("should fail with specific error", func() {
					out, err := utils.RunCommand(
						testDirPath,
						werfBinPath,
						werfCommand...,
					)
					Ω(err).ShouldNot(Succeed())
					Ω(string(out)).Should(ContainSubstring("used by container"))
					Ω(string(out)).Should(ContainSubstring("Use --force option to remove all containers that are based on deleting werf docker images"))
				})

				It("should remove project images and container", func() {
					werfArgs := append(werfCommand, "--force")
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						werfArgs...,
					)

					Ω(stagesStorageRepoImagesCount()).Should(Equal(0))
				})
			})
		})
	}
})
