package common_test

import (
	"os"
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

var _ = Describe("context", func() {
	BeforeEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"clone", werfRepositoryDir, testDirPath,
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"checkout", "-b", "integration-context-test", "v1.0.10",
		)
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	type entry struct {
		prepareFixturesFunc      func()
		expectedSignature        string
		expectedWindowsSignature string
	}

	var itBody = func(entry entry) {
		entry.prepareFixturesFunc()

		output, err := utils.RunCommand(
			testDirPath,
			werfBinPath,
			"build", "-s", ":local",
		)
		Ω(err).ShouldNot(HaveOccurred())

		if runtime.GOOS != "windows" {
			Ω(string(output)).Should(ContainSubstring(entry.expectedSignature))
		} else {
			Ω(string(output)).Should(ContainSubstring(entry.expectedWindowsSignature))
		}
	}

	var _ = DescribeTable("context", itBody,
		Entry("checksum with files read", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context"), testDirPath)
				Ω(os.RemoveAll(filepath.Join(testDirPath, ".git"))).Should(Succeed())
			},
			expectedSignature:        "c425035208193d0589aa4e4187575db3fc1d4ffa78f2101f020fadff",
			expectedWindowsSignature: "8bdd7688bad2d9dd6d77108a5e68c1e5748bd08efcfcb5f86a92f504",
		}),
		Entry("checksum with ls-tree", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context"), testDirPath)
			},
			expectedSignature:        "6d2596f2519f3fa66f3b24e370d86441065f5b1db80e157bfcad2458 ",
			expectedWindowsSignature: "f267b44d1503dd6e0d1e940af53d2ce85244da93ee74ea4e5854b2f1",
		}),
		Entry("checksum with ls-tree and status", entry{
			prepareFixturesFunc: func() {
				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"reset", "HEAD~50",
				)

				utils.CopyIn(utils.FixturePath("context"), testDirPath)
			},
			expectedSignature:        "4315b177e36f26ed9132986bdcbef6ce60f4891d733d944ebd80d598",
			expectedWindowsSignature: "ec6330c3760c0a771a79d312dba2bb790ed74a866b5586061ee932d1",
		}),
	)
})
