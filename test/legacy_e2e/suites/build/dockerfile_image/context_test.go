package common_test

import (
	"path/filepath"

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
		prepareFixturesFunc func(ctx SpecContext)
		expectedDigest      string
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

		Expect(string(output)).Should(ContainSubstring(entry.expectedDigest))
	}

	_ = DescribeTable("checksum", itBody,
		Entry("base", entry{
			prepareFixturesFunc: func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("context", "base"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "add", "werf.yaml", ".dockerignore", "Dockerfile")
				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "commit", "-m", "+")
			},
			expectedDigest: "cfd0121a39722eb3a9f45e91ea0535ec121cb15128540fd4337d880d",
		}),
		Entry("file from contextAddFile added to context", entry{
			prepareFixturesFunc: func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("context", "context_add_file"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile")
				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "commit", "-m", "+")
			},
			expectedDigest: "c6f1b7b3ee57e1d77db2e8d432920e9f8287da3edbba0d4d54f38dcf",
		}),
		Entry("symlinks from contextAddFiles added to context as is", entry{
			prepareFixturesFunc: func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("context", "context_add_symlinks"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile")
				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "commit", "-m", "+")
			},
			expectedDigest: "2d08a1690ca94ee41b8c4249348080ce14cfbe033307e00abbd4d873",
		}),
		Entry("dir from contextAddFiles added to context", entry{
			prepareFixturesFunc: func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("context", "context_add_dir"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile")
				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "commit", "-m", "+")
			},
			expectedDigest: "94f895c460edba91900451e8b2d21bbe3a0e52d2998d93e5f5222efd",
		}),
		Entry("specified files from dir allowed in allowContextAddFiles added to context", entry{
			prepareFixturesFunc: func(ctx SpecContext) {
				utils.CopyIn(utils.FixturePath("context", "context_add_dir_partially"), SuiteData.WerfRepoWorktreeDir)

				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "add", "werf.yaml", "werf-giterminism.yaml", ".dockerignore", "Dockerfile")
				utils.RunSucceedCommand(ctx, SuiteData.WerfRepoWorktreeDir, "git", "commit", "-m", "+")
			},
			expectedDigest: "04cadae79b55aae66127e3a4423a9fa9a6a3d52d289d46f074a19cfe",
		}),
	)
})
