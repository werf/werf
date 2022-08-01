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
	werfRepositoryDir, err = filepath.Abs("../../../../")
	if err != nil {
		panic(err)
	}
}

var _ = Describe("context", func() {
	BeforeEach(func() {
		SuiteData.WerfRepoWorktreeDir = filepath.Join(SuiteData.TestDirPath, "werf_repo_worktree")

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"clone", werfRepositoryDir, SuiteData.WerfRepoWorktreeDir,
		)

		utils.RunSucceedCommand(
			SuiteData.WerfRepoWorktreeDir,
			"git",
			"checkout", "-b", "integration-context-test", "v1.0.10",
		)
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			SuiteData.WerfRepoWorktreeDir,
			SuiteData.WerfBinPath,
			"host", "purge", "--force",
		)
	})

	type entry struct {
		prepareFixturesFunc   func()
		expectedWindowsDigest string
		expectedUnixDigest    string
		expectedDigest        string
	}

	itBody := func(entry entry) {
		entry.prepareFixturesFunc()

		output, err := utils.RunCommand(
			SuiteData.WerfRepoWorktreeDir,
			SuiteData.WerfBinPath,
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

	_ = DescribeTable("checksum", itBody,
		Entry("base", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context", "base"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(
					SuiteData.WerfRepoWorktreeDir,
					"git",
					"add", "werf.yaml", ".dockerignore", "Dockerfile",
				)
				utils.RunSucceedCommand(
					SuiteData.WerfRepoWorktreeDir,
					"git",
					"commit", "-m", "+",
				)
			},
			expectedDigest: "26f6bd1d7de41678c4dcfae8a3785d9655ee6b13c16e4498abb43d0b",
		}),
		Entry("file from contextAddFile added to context", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context", "context_add_file"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(
					SuiteData.WerfRepoWorktreeDir,
					"git",
					"add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile",
				)
				utils.RunSucceedCommand(
					SuiteData.WerfRepoWorktreeDir,
					"git",
					"commit", "-m", "+",
				)
			},
			expectedWindowsDigest: "b1c6be25d30d2de58df66e46dc8a328176cc2744dc3bfc2ae8d2917b",
			expectedUnixDigest:    "48a81bd49a6d299f78b463628ef6dd2436c2fce6736f2ad624b92e7f",
		}),
		Entry("symlinks from contextAddFiles added to context as is", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context", "context_add_symlinks"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(
					SuiteData.WerfRepoWorktreeDir,
					"git",
					"add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile",
				)
				utils.RunSucceedCommand(
					SuiteData.WerfRepoWorktreeDir,
					"git",
					"commit", "-m", "+",
				)
			},
			expectedWindowsDigest: "b602288def378e30337507f00a9dfb618eee38a8a880411b88353470",
			expectedUnixDigest:    "7e860ed9abcaa83496e6422cbc4d819dff064a1cc91ce618fb8dcfb6",
		}),
		Entry("dir from contextAddFiles added to context", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context", "context_add_dir"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(
					SuiteData.WerfRepoWorktreeDir,
					"git",
					"add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile",
				)
				utils.RunSucceedCommand(
					SuiteData.WerfRepoWorktreeDir,
					"git",
					"commit", "-m", "+",
				)
			},
			expectedWindowsDigest: "d04740140584330309c128175dd9c714aaa8bd536b83609f2ab14e4a",
			expectedUnixDigest:    "786fca63d59552405ce81b2361b9a396a93d7fcab7a5c84fbe6a8e48",
		}),
		Entry("specified files from dir allowed in allowContextAddFiles added to context", entry{
			prepareFixturesFunc: func() {
				utils.CopyIn(utils.FixturePath("context", "context_add_dir_partially"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(
					SuiteData.WerfRepoWorktreeDir,
					"git",
					"add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile",
				)
				utils.RunSucceedCommand(
					SuiteData.WerfRepoWorktreeDir,
					"git",
					"commit", "-m", "+",
				)
			},
			expectedWindowsDigest: "1cae63a8395a5cdb32936ec77e09e1954ec7698a75e0cedba2c11eff",
			expectedUnixDigest:    "164c8fdaecfd09657e1c6d8a9c780aa814814faf93c50901db38770f",
		}),
	)
})
