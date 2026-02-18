package common_test

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

var werfRepositoryDir string

func init() {
	var err error
	werfRepositoryDir, err = filepath.Abs("../../../../../")
	if err != nil {
		panic(err)
	}
}

var _ = Describe("context", func() {
	BeforeEach(func(ctx SpecContext) {
		SuiteData.WerfRepoWorktreeDir = filepath.Join(SuiteData.TestDirPath, "werf_repo_worktree")

		utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "clone", werfRepositoryDir, SuiteData.WerfRepoWorktreeDir)

		utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "checkout", "-b", "integration-context-test", "v1.0.10")
	})

	AfterEach(func(ctx SpecContext) {
		utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, SuiteData.WerfBinPath, "host", "purge", "--force")
	})

	type entry struct {
		prepareFixturesFunc   func(ctx SpecContext)
		expectedWindowsDigest string
		expectedUnixDigest    string
		expectedDigest        string
	}

	itBody := func(ctx SpecContext, entry entry) {
		entry.prepareFixturesFunc(ctx)

		output, err := utils.RunCommand(
			ctx,
			SuiteData.WerfRepoWorktreeDir,
			SuiteData.WerfBinPath,
			"build", "--debug",
		)
		Expect(err).ShouldNot(HaveOccurred())

		if runtime.GOOS == "windows" && entry.expectedWindowsDigest != "" {
			Expect(string(output)).Should(ContainSubstring(entry.expectedWindowsDigest))
		} else if entry.expectedUnixDigest != "" {
			Expect(string(output)).Should(ContainSubstring(entry.expectedUnixDigest))
		} else {
			Expect(string(output)).Should(ContainSubstring(entry.expectedDigest))
		}
	}

	_ = DescribeTable("checksum", itBody,
		Entry("base", entry{
			prepareFixturesFunc: func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("context", "base"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "add", "werf.yaml", ".dockerignore", "Dockerfile")
				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "commit", "-m", "+")
			},
			expectedDigest: "176144ade9f12b623accfe76b0a47094e5368be135f7d1066673e6d8",
		}),
		Entry("file from contextAddFile added to context", entry{
			prepareFixturesFunc: func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("context", "context_add_file"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile")
				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "commit", "-m", "+")
			},
			expectedWindowsDigest: "b1c6be25d30d2de58df66e46dc8a328176cc2744dc3bfc2ae8d2917b",
			expectedUnixDigest:    "15e0c2b78e9ee4f81c40df8a3c766c932728009d50a206082365276a",
		}),
		Entry("symlinks from contextAddFiles added to context as is", entry{
			prepareFixturesFunc: func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("context", "context_add_symlinks"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile")
				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "commit", "-m", "+")
			},
			expectedWindowsDigest: "b602288def378e30337507f00a9dfb618eee38a8a880411b88353470",
			expectedUnixDigest:    "e9ee32dfdeac4859dd532d1ce3453b4880a087460e17930eb94df876",
		}),
		Entry("dir from contextAddFiles added to context", entry{
			prepareFixturesFunc: func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("context", "context_add_dir"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile")
				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "commit", "-m", "+")
			},
			expectedWindowsDigest: "d04740140584330309c128175dd9c714aaa8bd536b83609f2ab14e4a",
			expectedUnixDigest:    "7cc4dce1ff9891ed3b70f4e476e16a7d0d42c88b5678c819711b9757",
		}),
		Entry("specified files from dir allowed in allowContextAddFiles added to context", entry{
			prepareFixturesFunc: func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("context", "context_add_dir_partially"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile")
				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "commit", "-m", "+")
			},
			expectedWindowsDigest: "1cae63a8395a5cdb32936ec77e09e1954ec7698a75e0cedba2c11eff",
			expectedUnixDigest:    "b23c7172bc4ad67104017f75ad444993f1686727c2ad07fdaa7d6f0b",
		}),
	)
})
