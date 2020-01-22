package common_test

import (
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
)

var _ = Describe("persistent stage signatures", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("signature"), testDirPath)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"init",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"add", "-A",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"commit", "-m", "Initial commit",
		)
	})

	type entry struct {
		imageName          string
		expectedSignatures []string
		skipOnWindows      bool
	}

	DescribeTable("should not be changed", func(e entry) {
		if e.skipOnWindows && runtime.GOOS == "windows" {
			Skip("skip on windows")
		}

		output, err := utils.RunCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", e.imageName,
		)

		Ω(err).Should(HaveOccurred())

		for _, signature := range e.expectedSignatures {
			Ω(string(output)).Should(ContainSubstring(signature))
		}
	},
		Entry("dockerfile_image", entry{
			imageName: "dockerfile_image",
			expectedSignatures: []string{
				"dockerfile:             2dbd745e5f0b42a6eed2c244e5a2d5c74fe3a29dbfe88d25ea4205c4011d5f77",
			},
			skipOnWindows: true,
		}),
		Entry("dockerfile_image_based_on_stage", entry{
			imageName: "dockerfile_image_based_on_stage",
			expectedSignatures: []string{
				"dockerfile:             5c4d5ceed2f87b8819c3e1a93232877cf1ed53faed5b97031d79658ecc5d55db",
			},
			skipOnWindows: true,
		}),
		Entry("stapel_image_shell", entry{
			imageName: "stapel_image_shell",
			expectedSignatures: []string{
				"from:                   4a10aead4134628c8e3b326a934e0a0068b28f824216a30b163b795e20401222",
				"from:                   ce404c2be4cf10404b1dd231691acb74ab780e31e8b5a687480ccea2c2be69cb",
				"beforeInstall:          00612d8175a189b28078400638571e0311eecd289df39b7c4e82e3fa246e76e4",
				"importsBeforeInstall:   4e8a2199b5852119b15124a97e08b8016f9c6cf4eec9d7b335a3632365a526e4",
				"install:                46d7e5256d350557064528320a9e9f6d4ede12e1071389abf9e9c7626342a8bf",
				"beforeSetup:            a883b7a0e30c66109377b779855c60615e1b731adaa1031b396797bf4ecb497a",
				"setup:                  d4624fe646ceff8341d240f75e8e00fb581164ba24e1b45de716b20de35316c9",
				"importsAfterSetup:      1bfe02b030a0bfca7ddfb68650f67b63ed8fda6a440044d9a86168b059ffab18",
				"dockerInstructions:     e8be54d6d3de841bd81ac6df9395a1e31e362638b12122d4c8835c1c07552822",
			},
		}),
		Entry("stapel_image_ansible", entry{
			imageName: "stapel_image_ansible",
			expectedSignatures: []string{
				"from:                   4a10aead4134628c8e3b326a934e0a0068b28f824216a30b163b795e20401222",
				"from:                   ce404c2be4cf10404b1dd231691acb74ab780e31e8b5a687480ccea2c2be69cb",
				"beforeInstall:          fb440453199cda29a121c8672c281b8e4c4b4c0ced28f0f37ed0ec230ae1cfa9",
				"importsBeforeInstall:   6ccb2730bb8f40c3f24be980c669939cd9fd11cf4e74ffb470b7a877703328db",
				"install:                873da454ce602eee95faf6d08993da6b767c595e5ad7ce3b09f8f02512bb878b",
				"beforeSetup:            bed74bae853ea7d042e933f4a78a09a63785cebe7ff49c643c8cb24f8aa29b51",
				"setup:                  11d9dab6b39938d495357999557e0e0ece362dc02b2c8d8aebac523baded0fbf",
				"importsAfterSetup:      73e6cbd236ec78dd57ccdc26b524d1b44d91e15e550646a6557946a6e4e9101b",
				"dockerInstructions:     39a038f34ee2d2a79f83911bddd568e7b0828da99f68822c0e09990de677e9bb",
			},
		}),
		Entry("stapel_image_from_image", entry{
			imageName: "stapel_image_from_image",
			expectedSignatures: []string{
				"from:                   4a10aead4134628c8e3b326a934e0a0068b28f824216a30b163b795e20401222",
				"from:                   ce404c2be4cf10404b1dd231691acb74ab780e31e8b5a687480ccea2c2be69cb",
				"beforeInstall:          00612d8175a189b28078400638571e0311eecd289df39b7c4e82e3fa246e76e4",
				"importsBeforeInstall:   4e8a2199b5852119b15124a97e08b8016f9c6cf4eec9d7b335a3632365a526e4",
				"install:                46d7e5256d350557064528320a9e9f6d4ede12e1071389abf9e9c7626342a8bf",
				"beforeSetup:            a883b7a0e30c66109377b779855c60615e1b731adaa1031b396797bf4ecb497a",
				"setup:                  d4624fe646ceff8341d240f75e8e00fb581164ba24e1b45de716b20de35316c9",
				"importsAfterSetup:      1bfe02b030a0bfca7ddfb68650f67b63ed8fda6a440044d9a86168b059ffab18",
				"dockerInstructions:     e8be54d6d3de841bd81ac6df9395a1e31e362638b12122d4c8835c1c07552822",
				"from:                   928c083f5a0d876cf2ba334ac5f9a9b31664cf5805a1e735d51161d346f9b775",
			},
		}),
		Entry("stapel_image_from_artifact", entry{
			imageName: "stapel_image_from_artifact",
			expectedSignatures: []string{
				"from:                   4a10aead4134628c8e3b326a934e0a0068b28f824216a30b163b795e20401222",
				"from:                   ce404c2be4cf10404b1dd231691acb74ab780e31e8b5a687480ccea2c2be69cb",
				"beforeInstall:          00612d8175a189b28078400638571e0311eecd289df39b7c4e82e3fa246e76e4",
				"importsBeforeInstall:   4e8a2199b5852119b15124a97e08b8016f9c6cf4eec9d7b335a3632365a526e4",
				"install:                46d7e5256d350557064528320a9e9f6d4ede12e1071389abf9e9c7626342a8bf",
				"beforeSetup:            a883b7a0e30c66109377b779855c60615e1b731adaa1031b396797bf4ecb497a",
				"setup:                  d4624fe646ceff8341d240f75e8e00fb581164ba24e1b45de716b20de35316c9",
				"importsAfterSetup:      1bfe02b030a0bfca7ddfb68650f67b63ed8fda6a440044d9a86168b059ffab18",
				"from:                   b5b1babc9125e58adc29572951df6347aa68ba4cefcd64c4125b1f63116a4070",
			},
		}))
})
