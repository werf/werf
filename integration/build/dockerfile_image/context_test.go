package common_test

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/utils"
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
			"purge", "--force",
		)
	})

	type entry struct {
		prepareFixturesFunc   func()
		expectedWindowsDigest string
		expectedUnixDigest    string
		expectedDigest        string
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
		} else if entry.expectedUnixDigest != "" {
			Ω(string(output)).Should(ContainSubstring(entry.expectedUnixDigest))
		} else {
			Ω(string(output)).Should(ContainSubstring(entry.expectedDigest))
		}
	}

	var _ = DescribeTable("checksum", itBody,
		Entry("base", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context", "base"), testDirPath)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"add", "werf.yaml", ".dockerignore", "Dockerfile",
				)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"commit", "-m", "+",
				)
			},
			expectedDigest: "26f6bd1d7de41678c4dcfae8a3785d9655ee6b13c16e4498abb43d0b",
		}),
		Entry("contextAdd", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context", "context_add_file"), testDirPath)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"add", "werf.yaml", ".dockerignore", "Dockerfile",
				)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"commit", "-m", "+",
				)
			},
			expectedWindowsDigest: "f4f979dc59d00427ae092d7b98a082143504e3fafbe5e01ef913e5a5",
			expectedUnixDigest:    "4d168006f579e786eb009927a517582ddb40ad2199aa4ad806e38d0b",
		}),
	)
})
