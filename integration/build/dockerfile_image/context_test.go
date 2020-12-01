package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"path/filepath"

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
		prepareFixturesFunc func()
		expectedDigest      string
	}

	var itBody = func(entry entry) {
		entry.prepareFixturesFunc()

		output, err := utils.RunCommand(
			testDirPath,
			werfBinPath,
			"build", "--debug",
		)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(string(output)).Should(ContainSubstring(entry.expectedDigest))
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
			expectedDigest: "26d19298c5565d81f8559fa8683f708924482831b17904671738fd64",
		}),
	)
})
