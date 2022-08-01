package common_test

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

var werfRepositoryDir string

func init() {
	var err error
	werfRepositoryDir, err = filepath.Abs("../../../")
	if err != nil {
		panic(err)
	}
}

var _ = XDescribe("persistent stage digests", func() {
	BeforeEach(func() {
		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"clone", werfRepositoryDir, SuiteData.TestDirPath,
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"checkout", "-b", "integration-digest-test", "v1.0.10",
		)

		utils.CopyIn(utils.FixturePath("digest"), SuiteData.TestDirPath)
	})

	type entry struct {
		imageName              string
		expectedDigests        []string
		expectedWindowsDigests []string
		skipOnWindows          bool
	}

	DescribeTable("should not be changed", func(e entry) {
		output, err := utils.RunCommand(
			SuiteData.TestDirPath,
			SuiteData.WerfBinPath,
			"build", e.imageName,
		)

		Ω(err).NotTo(HaveOccurred())

		var expectedDigests []string
		if runtime.GOOS == "windows" && len(e.expectedWindowsDigests) != 0 {
			expectedDigests = e.expectedWindowsDigests
		} else {
			expectedDigests = e.expectedDigests
		}

		for _, digest := range expectedDigests {
			Ω(string(output)).Should(ContainSubstring(digest))
		}
	},
		Entry("dockerfile_image", entry{
			imageName: "dockerfile_image",
			expectedDigests: []string{
				"dockerfile_image/dockerfile", "tag: 66be3adaf40fba215530be3abaa0bdfbc6005abdbd6d5f8957e031db-",
			},
			expectedWindowsDigests: []string{
				"dockerfile_image/dockerfile", "tag: 9642c67560c27be99d21076fcb37dd60959c4269c53870d553e2282b-",
			},
			skipOnWindows: true,
		}),
		Entry("dockerfile_image_based_on_stage", entry{
			imageName: "dockerfile_image_based_on_stage",
			expectedDigests: []string{
				"dockerfile_image_based_on_stage/dockerfile", "tag: c52464f2235835ba66266d7b7f844fa399aa362706644583b7f32293-",
			},
			expectedWindowsDigests: []string{
				"dockerfile_image_based_on_stage/dockerfile", "tag: c52464f2235835ba66266d7b7f844fa399aa362706644583b7f32293-",
			},
			skipOnWindows: true,
		}),
		Entry("stapel_image_shell", entry{
			imageName: "stapel_image_shell",
			expectedDigests: []string{
				"import_artifact/from", "tag: 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47-",
				"import_artifact/install", "tag: f9026091241bb85eac3c2413333b4269cc309dde7a29d7ebffdd05d1-",
				"import_image/from", "tag: 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47-",
				"import_image/install", "tag: f9026091241bb85eac3c2413333b4269cc309dde7a29d7ebffdd05d1-",
				"stapel_image_shell/from", "tag: 4c102ca0d0645f3ba5def446ddebeeec76adffe91264b678507037bd-",
				"stapel_image_shell/beforeInstall", "tag: 4dbbb84d19fcda50b578576c221d14e1b296a2e0437dd44356b4f86f-",
				"stapel_image_shell/importsBeforeInstall", "tag: a83bed392f696e0c6e2f15014706f93aa1a60bacfa52935bd6a22887-",
				"stapel_image_shell/gitArchive", "tag: 417f06df5f1434365e23943f8bc75b6fc8117f48fc124631ceed3c26-",
				"stapel_image_shell/install", "tag: c91779d2713c9cb4f7559d0b6356968ee0de1e026849269db10ff16b-",
				"stapel_image_shell/beforeSetup", "tag: b60efd01e74f1567e427dadadd8f7ab8afa593e6979417864ede004e-",
				"stapel_image_shell/setup", "tag: 2bd3f896c6c0081261961ee72631e4042b983ec8ddde05bcdb6c3b29-",
				"stapel_image_shell/importsAfterSetup", "tag: f94feaa781d594c2e33c450f85527704e8f0395bb3ad8b02ad892bdf-",
				"stapel_image_shell/dockerInstructions", "tag: 5b472b016aa6d05fea9e8b7e8410a43cadfd7832db522fb14460a688-",
			},
		}),
		Entry("stapel_image_ansible", entry{
			imageName: "stapel_image_ansible",
			expectedDigests: []string{
				"import_artifact/from", "tag: 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47-",
				"import_artifact/install", "tag: f9026091241bb85eac3c2413333b4269cc309dde7a29d7ebffdd05d1-",
				"import_image/from", "tag: 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47-",
				"import_image/install", "tag: f9026091241bb85eac3c2413333b4269cc309dde7a29d7ebffdd05d1-",
				"stapel_image_ansible/from", "tag: 4c102ca0d0645f3ba5def446ddebeeec76adffe91264b678507037bd-",
				"stapel_image_ansible/beforeInstall", "tag: 9898ed1d8b3867bb8ef3de46e39caebbef110beefaac2ad992642204-",
				"stapel_image_ansible/importsBeforeInstall", "tag: 8e98cbaebefa3aac5bc733d9457d7ec8ab6e6bb8312d10fdcc4b300a-",
				"stapel_image_ansible/gitArchive", "tag: 2e63f9d4025aa4ac2d323c59ca62311e55357fb7b3013bb9dd08d6b1-",
				"stapel_image_ansible/install", "tag: 5b075521a9f662f405fda91349aacba5c62b7d58fef96ecbac2abe16-",
				"stapel_image_ansible/beforeSetup", "tag: ac07a896a0d254e131ed88085b28fe7ea5ed668214dec66fe7e3c1c9-",
				"stapel_image_ansible/setup", "tag: 498804176d5cbd1b5581fa54062a172adf83a89d53a888eac6d8c6be-",
				"stapel_image_ansible/importsAfterSetup", "tag: ab649da9cd18d665da6f80d7a941360d1312bcd10c3d382c12bdac6c-",
				"stapel_image_ansible/dockerInstructions", "tag: 9a07671791b6d1455239e305465da38bc6e65cec55fc5eba0854dacb-",
			},
		}),
		Entry("stapel_image_from_image", entry{
			imageName: "stapel_image_from_image",
			expectedDigests: []string{
				"import_artifact/from", "tag: 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47-",
				"import_artifact/install", "tag: f9026091241bb85eac3c2413333b4269cc309dde7a29d7ebffdd05d1-",
				"import_image/from", "tag: 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47-",
				"import_image/install", "tag: f9026091241bb85eac3c2413333b4269cc309dde7a29d7ebffdd05d1-",
				"stapel_image_shell/from", "tag: 4c102ca0d0645f3ba5def446ddebeeec76adffe91264b678507037bd-",
				"stapel_image_shell/beforeInstall", "tag: 4dbbb84d19fcda50b578576c221d14e1b296a2e0437dd44356b4f86f-",
				"stapel_image_shell/importsBeforeInstall", "tag: a83bed392f696e0c6e2f15014706f93aa1a60bacfa52935bd6a22887-",
				"stapel_image_shell/gitArchive", "tag: 417f06df5f1434365e23943f8bc75b6fc8117f48fc124631ceed3c26-",
				"stapel_image_shell/install", "tag: c91779d2713c9cb4f7559d0b6356968ee0de1e026849269db10ff16b-",
				"stapel_image_shell/beforeSetup", "tag: b60efd01e74f1567e427dadadd8f7ab8afa593e6979417864ede004e-",
				"stapel_image_shell/setup", "tag: 2bd3f896c6c0081261961ee72631e4042b983ec8ddde05bcdb6c3b29-",
				"stapel_image_shell/importsAfterSetup", "tag: f94feaa781d594c2e33c450f85527704e8f0395bb3ad8b02ad892bdf-",
				"stapel_image_shell/dockerInstructions", "tag: 5b472b016aa6d05fea9e8b7e8410a43cadfd7832db522fb14460a688-",
				"stapel_image_from_image/from", "tag: ec92ec7f41214433b845eba1feb0cb120c8bad981487832a6f54cbd2-",
			},
		}),
		Entry("stapel_image_from_artifact", entry{
			imageName: "stapel_image_from_artifact",
			expectedDigests: []string{
				"import_artifact/from", "tag: 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47-",
				"import_artifact/install", "tag: f9026091241bb85eac3c2413333b4269cc309dde7a29d7ebffdd05d1-",
				"import_image/from", "tag: 2bc41fbd00277e3021c613fcfdcef2716ed893bee29b36f928136e47-",
				"import_image/install", "tag: f9026091241bb85eac3c2413333b4269cc309dde7a29d7ebffdd05d1-",
				"stapel_artifact_shell/from", "tag: 4c102ca0d0645f3ba5def446ddebeeec76adffe91264b678507037bd-",
				"stapel_artifact_shell/beforeInstall", "tag: 4dbbb84d19fcda50b578576c221d14e1b296a2e0437dd44356b4f86f-",
				"stapel_artifact_shell/importsBeforeInstall", "tag: a83bed392f696e0c6e2f15014706f93aa1a60bacfa52935bd6a22887-",
				"stapel_artifact_shell/gitArchive", "tag: 417f06df5f1434365e23943f8bc75b6fc8117f48fc124631ceed3c26-",
				"stapel_artifact_shell/install", "tag: c91779d2713c9cb4f7559d0b6356968ee0de1e026849269db10ff16b-",
				"stapel_artifact_shell/beforeSetup", "tag: b60efd01e74f1567e427dadadd8f7ab8afa593e6979417864ede004e-",
				"stapel_artifact_shell/setup", "tag: 2bd3f896c6c0081261961ee72631e4042b983ec8ddde05bcdb6c3b29-",
				"stapel_artifact_shell/importsAfterSetup", "tag: f94feaa781d594c2e33c450f85527704e8f0395bb3ad8b02ad892bdf-",
				"stapel_image_from_artifact/from", "tag: fd76841961a39e279d67d71540b00b4d8dba16c46a95a062f598d20d-",
			},
		}),
	)
})
