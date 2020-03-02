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

	var _ = DescribeTable("checksum", itBody,
		Entry("with files read", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context"), testDirPath)
				Ω(os.RemoveAll(filepath.Join(testDirPath, ".git"))).Should(Succeed())
			},
			expectedSignature:        "02473513cc5a901dd98785602a36ff2e192d40260054634cf81fa41c",
			expectedWindowsSignature: "d08e5f4366f31e10f23beca819a01fc51eeabcf6c6ec6dfd17646da7",
		}),
		Entry("with ls-tree", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context"), testDirPath)
			},
			expectedSignature:        "0ee2ba14ff8084049d694748977873c3bcab905cdbe3c1caac8204d3",
			expectedWindowsSignature: "9ba084272d896bc3d5d20ddc98f08edeb8c92de03121fc63a9002025",
		}),
		Entry("with ls-tree and status", entry{
			prepareFixturesFunc: func() {
				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"reset", "HEAD~50",
				)

				utils.CopyIn(utils.FixturePath("context"), testDirPath)
			},
			expectedSignature:        "d4f36d7d05db896ac2067e2e30bea131ce9c32142d6d31f83c7d3d9e",
			expectedWindowsSignature: "51d0ed2fbc218b4eb7860f910bdab9eedaa2528a9fa3b88bbb8eebc4",
		}),
	)
})
