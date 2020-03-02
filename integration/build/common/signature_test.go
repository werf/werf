package common_test

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
)

var werfRepositoryDir string

func init() {
	var err error
	werfRepositoryDir, err = filepath.Abs("../../../")
	if err != nil {
		panic(err)
	}
}

var _ = Describe("persistent stage signatures", func() {
	BeforeEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"clone", werfRepositoryDir, testDirPath,
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"checkout", "-b", "integration-signature-test", "v1.0.10",
		)

		utils.CopyIn(utils.FixturePath("signature"), testDirPath)

		stubs.SetEnv("WERF_LOG_VERBOSE", "1")
	})

	type entry struct {
		imageName                 string
		expectedSignatures        []string
		expectedWindowsSignatures []string
		skipOnWindows             bool
	}

	DescribeTable("should not be changed", func(e entry) {
		output, err := utils.RunCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", e.imageName,
		)

		Ω(err).Should(HaveOccurred())

		var expectedSignatures []string
		if runtime.GOOS == "windows" && len(e.expectedWindowsSignatures) != 0 {
			expectedSignatures = e.expectedWindowsSignatures
		} else {
			expectedSignatures = e.expectedSignatures
		}

		for _, signature := range expectedSignatures {
			Ω(string(output)).Should(ContainSubstring(signature))
		}
	},
		Entry("dockerfile_image", entry{
			imageName: "dockerfile_image",
			expectedSignatures: []string{
				"stage dockerfile_image/dockerfile with signature 66be3adaf40fba215530be3abaa0bdfbc6005abdbd6d5f8957e031db",
			},
			expectedWindowsSignatures: []string{
				"stage dockerfile_image/dockerfile with signature 9642c67560c27be99d21076fcb37dd60959c4269c53870d553e2282b",
			},
			skipOnWindows: true,
		}),
		Entry("dockerfile_image_based_on_stage", entry{
			imageName: "dockerfile_image_based_on_stage",
			expectedSignatures: []string{
				"stage dockerfile_image_based_on_stage/dockerfile with signature de15809392f59fef94bceffe1e25d87f6762281b5fa2f7435bf014f3",
			},
			expectedWindowsSignatures: []string{
				"stage dockerfile_image_based_on_stage/dockerfile with signature de15809392f59fef94bceffe1e25d87f6762281b5fa2f7435bf014f3",
			},
			skipOnWindows: true,
		}),
		Entry("stapel_image_shell", entry{
			imageName: "stapel_image_shell",
			expectedSignatures: []string{
				"stage import_artifact/from with signature 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47",
				"stage import_image/from with signature 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47",
				"stage stapel_image_shell/from with signature 397ac70071d48fdf1b586124cb4ded4743af5f5607b51181ac3a050b",
				"stage stapel_image_shell/beforeInstall with signature 4291ffe66dfc08a229e4b3490eafc5bef1f216e98fa4045c9f91adab",
				"stage stapel_image_shell/importsBeforeInstall with signature eabe68ed5026dbabe78ff8c99e5f3f6666f366b167c796c4f691fe60",
				"stage stapel_image_shell/gitArchive with signature e92aa641a479c867cc08c64b02289055d0ebb05bf02189bb4b357427",
				"stage stapel_image_shell/install with signature c95e58cb28fc319b83c4641791478f8170897eaa13b82e3b7ff035aa",
				"stage stapel_image_shell/beforeSetup with signature f2b08e786597679441e3074146498f47826ee56ab370a62d8e907793",
				"stage stapel_image_shell/setup with signature a699d04d07a2704b9daa80882a2342bf39a093534387e0cabfe28065",
				"stage stapel_image_shell/importsAfterSetup with signature d8078a794fc9e445a21646d54e40a874f49cbe3d3cafef1dab35ab55",
				"stage stapel_image_shell/dockerInstructions with signature 8e63dbfd3f18f150804c9623452b9dd430cff79f56ed543dee55ae8e",
			},
		}),
		Entry("stapel_image_ansible", entry{
			imageName: "stapel_image_ansible",
			expectedSignatures: []string{
				"stage import_artifact/from with signature 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47",
				"stage import_image/from with signature 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47",
				"stage stapel_image_ansible/from with signature 397ac70071d48fdf1b586124cb4ded4743af5f5607b51181ac3a050b",
				"stage stapel_image_ansible/beforeInstall with signature 4dc6392b23a9faa06ec3035056732e98bd476e2e54b38ca6632606eb",
				"stage stapel_image_ansible/importsBeforeInstall with signature a94062e7af8ee11778b6af3ac3360108dd3553d977aadde7a912da69",
				"stage stapel_image_ansible/gitArchive with signature 3735d7d22cdb3d8a3b0689fb7b187541df5c2af883dc1987d90e60c4",
				"stage stapel_image_ansible/install with signature 552d0582c955c0c4a1390f8b215bafc18a136b0f8344b097dbab5d34",
				"stage stapel_image_ansible/beforeSetup with signature c8ee225afdcb74a95df003326f3e723031cb42fa8c894890afedde6c",
				"stage stapel_image_ansible/setup with signature 99844134b6ebd84d88fcf06d62044570dedc8714ad788b0655c60bcd",
				"stage stapel_image_ansible/importsAfterSetup with signature f32efc0bbcb684e6ee12094ee125fcbf2d584684efef5f976090cd2a",
				"stage stapel_image_ansible/dockerInstructions with signature 85c2b1cc9216db8df8a05df567ae96fa271681a449bc2e7566c48089",
			},
		}),
		Entry("stapel_image_from_image", entry{
			imageName: "stapel_image_from_image",
			expectedSignatures: []string{
				"stage import_artifact/from with signature 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47",
				"stage import_image/from with signature 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47",
				"stage stapel_image_shell/from with signature 397ac70071d48fdf1b586124cb4ded4743af5f5607b51181ac3a050b",
				"stage stapel_image_shell/beforeInstall with signature 4291ffe66dfc08a229e4b3490eafc5bef1f216e98fa4045c9f91adab",
				"stage stapel_image_shell/importsBeforeInstall with signature eabe68ed5026dbabe78ff8c99e5f3f6666f366b167c796c4f691fe60",
				"stage stapel_image_shell/gitArchive with signature e92aa641a479c867cc08c64b02289055d0ebb05bf02189bb4b357427",
				"stage stapel_image_shell/install with signature c95e58cb28fc319b83c4641791478f8170897eaa13b82e3b7ff035aa",
				"stage stapel_image_shell/beforeSetup with signature f2b08e786597679441e3074146498f47826ee56ab370a62d8e907793",
				"stage stapel_image_shell/setup with signature a699d04d07a2704b9daa80882a2342bf39a093534387e0cabfe28065",
				"stage stapel_image_shell/importsAfterSetup with signature d8078a794fc9e445a21646d54e40a874f49cbe3d3cafef1dab35ab55",
				"stage stapel_image_shell/dockerInstructions with signature 8e63dbfd3f18f150804c9623452b9dd430cff79f56ed543dee55ae8e",
				"stage stapel_image_from_image/from with signature abb47ce1e3b4f27c21d1643c4fff6cb19ee573b1d25bed9a4fd19635",
			},
		}),
		Entry("stapel_image_from_artifact", entry{
			imageName: "stapel_image_from_artifact",
			expectedSignatures: []string{
				"stage import_artifact/from with signature 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47",
				"stage import_image/from with signature 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47",
				"stage stapel_artifact_shell/from with signature 397ac70071d48fdf1b586124cb4ded4743af5f5607b51181ac3a050b",
				"stage stapel_artifact_shell/beforeInstall with signature 4291ffe66dfc08a229e4b3490eafc5bef1f216e98fa4045c9f91adab",
				"stage stapel_artifact_shell/importsBeforeInstall with signature eabe68ed5026dbabe78ff8c99e5f3f6666f366b167c796c4f691fe60",
				"stage stapel_artifact_shell/gitArchive with signature e92aa641a479c867cc08c64b02289055d0ebb05bf02189bb4b357427",
				"stage stapel_artifact_shell/install with signature c95e58cb28fc319b83c4641791478f8170897eaa13b82e3b7ff035aa",
				"stage stapel_artifact_shell/beforeSetup with signature f2b08e786597679441e3074146498f47826ee56ab370a62d8e907793",
				"stage stapel_artifact_shell/setup with signature a699d04d07a2704b9daa80882a2342bf39a093534387e0cabfe28065",
				"stage stapel_artifact_shell/importsAfterSetup with signature d8078a794fc9e445a21646d54e40a874f49cbe3d3cafef1dab35ab55",
				"stage stapel_image_from_artifact/from with signature 097fa71cef389a87cad9538b202dd954492451673e80cfa1cb3e95a1",
			},
		}),
	)
})
