package common_test

import (
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/testing/utils"
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
			"stages", "purge", "--force",
		)
	})

	type entry struct {
		prepareFixturesFunc   func()
		expectedDigest        string
		expectedDarwinDigest  string
		expectedWindowsDigest string
	}

	var itBody = func(entry entry) {
		entry.prepareFixturesFunc()

		output, err := utils.RunCommand(
			testDirPath,
			werfBinPath,
			"build", "--debug",
		)
		Ω(err).ShouldNot(HaveOccurred())

		if runtime.GOOS == "windows" && entry.expectedWindowsDigest != "" {
			Ω(string(output)).Should(ContainSubstring(entry.expectedWindowsDigest))
		} else if runtime.GOOS == "darwin" && entry.expectedDarwinDigest != "" {
			Ω(string(output)).Should(ContainSubstring(entry.expectedDarwinDigest))
		} else {
			Ω(string(output)).Should(ContainSubstring(entry.expectedDigest))
		}
	}

	var _ = DescribeTable("checksum", itBody,
		Entry("without git", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context", "default"), testDirPath)
				Ω(os.RemoveAll(filepath.Join(testDirPath, ".git"))).Should(Succeed())
			},
			expectedDigest:        "10577fbfd229120fa34bc07fd40630af70a8051017b31ec4a86c1f76",
			expectedDarwinDigest:  "6419296f73e469ab97cb99defc7dc20c9ad7e9fbf211539e2d0f6639",
			expectedWindowsDigest: "36407a81113c9555fe5483ab04f42b8004cdbf0120b00bc129118f9b",
		}),
		Entry("with ls-tree", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context", "default"), testDirPath)
			},
			expectedDigest:        "0ee2ba14ff8084049d694748977873c3bcab905cdbe3c1caac8204d3",
			expectedWindowsDigest: "9ba084272d896bc3d5d20ddc98f08edeb8c92de03121fc63a9002025",
		}),
		Entry("with ls-tree and status", entry{
			prepareFixturesFunc: func() {
				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"reset", "HEAD~50",
				)

				utils.CopyIn(utils.FixturePath("context", "default"), testDirPath)
			},
			expectedDigest:        "d4f36d7d05db896ac2067e2e30bea131ce9c32142d6d31f83c7d3d9e",
			expectedWindowsDigest: "51d0ed2fbc218b4eb7860f910bdab9eedaa2528a9fa3b88bbb8eebc4",
		}),
		Entry("with ls-tree, status and ignored files by .gitignore files", entry{
			prepareFixturesFunc: func() {
				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"reset", "HEAD~50",
				)

				utils.CopyIn(utils.FixturePath("context", "default"), testDirPath)
				utils.CopyIn(utils.FixturePath("context", "gitignores"), testDirPath)
			},
			expectedDigest:        "4dac4b7874769660e42856e038261ad80d418a7b6672bd3658d5bd19",
			expectedWindowsDigest: "e3ee8c62496da6a52181cd09e296b63d8fef7e96e04c28fba1cda278",
		}),
	)
})
