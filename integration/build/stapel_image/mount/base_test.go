// +build integration

package mount_test

import (
	"os"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/flant/werf/integration/utils"
)

type entry struct {
	fixturePath                       string
	expectedFirstBuildOutputMatchers  []types.GomegaMatcher
	expectedSecondBuildOutputMatchers []types.GomegaMatcher
}

var itBody = func(e entry) {
	utils.CopyIn(e.fixturePath, testDirPath)

	Ω(os.Setenv("FROM_CACHE_VERSION", "1"))

	output := utils.SucceedCommandOutputString(
		testDirPath,
		werfBinPath,
		"build",
	)

	for _, match := range e.expectedFirstBuildOutputMatchers {
		Ω(output).Should(match)
	}

	Ω(os.Setenv("FROM_CACHE_VERSION", "2"))

	output = utils.SucceedCommandOutputString(
		testDirPath,
		werfBinPath,
		"build",
	)

	for _, match := range e.expectedSecondBuildOutputMatchers {
		Ω(output).Should(match)
	}

	utils.RunSucceedContainerCommandWithStapel(werfBinPath, testDirPath, []string{}, []string{"[[ -z \"$(ls -A /mount)\" ]]"})
}

var _ = DescribeTable("base", itBody,
	Entry("tmp_dir", entry{
		fixturePath: fixturePath("tmp_dir"),
		expectedFirstBuildOutputMatchers: []types.GomegaMatcher{
			ContainSubstring("Result number is 2"),
		},
		expectedSecondBuildOutputMatchers: []types.GomegaMatcher{
			ContainSubstring("Result number is 2"),
		},
	}),
	Entry("build_dir", entry{
		fixturePath: fixturePath("build_dir"),
		expectedFirstBuildOutputMatchers: []types.GomegaMatcher{
			ContainSubstring("Result number is 2"),
		},
		expectedSecondBuildOutputMatchers: []types.GomegaMatcher{
			ContainSubstring("Result number is 4"),
		},
	}),
	Entry("from_path", entry{
		fixturePath: fixturePath("from_path"),
		expectedFirstBuildOutputMatchers: []types.GomegaMatcher{
			ContainSubstring("Result number is 4"),
		},
		expectedSecondBuildOutputMatchers: []types.GomegaMatcher{
			ContainSubstring("Result number is 6"),
		},
	}))
