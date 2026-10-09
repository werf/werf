package true_git

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/path_matcher"
	"github.com/werf/werf/v3/test/pkg/suite_init"
)

var _ = ginkgo.Describe("LFS patch preparation", func() {
	suite_init.NewWerfInitData(SuiteData.TmpDirData)
	ginkgo.DescribeTable("ignores excluded unavailable objects with ordinary submodules",
		func(ctx ginkgo.SpecContext, install bool, poolLimit string) {
			isolateGitConfig()
			setEnvForSpec("GIT_CONFIG_COUNT", "1")
			setEnvForSpec("GIT_CONFIG_KEY_0", "protocol.file.allow")
			setEnvForSpec("GIT_CONFIG_VALUE_0", "always")
			setEnvForSpec("GIT_CONFIG_PARAMETERS", "")
			setEnvForSpec("GIT_LFS_SKIP_SMUDGE", "0")
			setEnvForSpec("GIT_LFS_SKIP_DOWNLOAD_ERRORS", "0")
			previousPoolLimit := workTreePoolLimit
			workTreePoolLimit = poolLimit
			ginkgo.DeferCleanup(func() { workTreePoolLimit = previousPoolLimit })
			gomega.Expect(Init(ctx, Options{})).To(gomega.Succeed())
			repoDir := filepath.Join(SuiteData.TestDirPath, "repo")
			subDir := filepath.Join(SuiteData.TestDirPath, "sub-origin")
			cacheDir := filepath.Join(SuiteData.TestDirPath, "cache")
			gitInitRepoWithFile(ctx, repoDir, "wanted.txt", "before\n")
			gitInitRepoWithFile(ctx, subDir, "plain.txt", "ordinary submodule\n")
			writeLFSTestFile(repoDir, "excluded.bin", "unavailable LFS object", false)
			gomega.Expect(os.WriteFile(filepath.Join(repoDir, ".gitattributes"), []byte("*.bin filter=lfs diff=lfs merge=lfs -text\n"), 0o644)).To(gomega.Succeed())
			gitAddSubmoduleSucceed(ctx, repoDir, subDir, "sub")
			gitSucceed(ctx, repoDir, "add", ".")
			gitCommitSucceed(ctx, repoDir, "-m", "ordinary submodule and excluded pointer")
			from := gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")
			gomega.Expect(os.WriteFile(filepath.Join(repoDir, "wanted.txt"), []byte("after\n"), 0o644)).To(gomega.Succeed())
			gitSucceed(ctx, repoDir, "add", "wanted.txt")
			gitCommitSucceed(ctx, repoDir, "-m", "selected update")
			to := gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")
			gitSucceed(ctx, repoDir, "remote", "add", "origin", "file://"+repoDir)
			if install {
				gitSucceed(ctx, repoDir, "lfs", "install", "--local", "--skip-repo")
			}
			matcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: []string{"wanted.txt"}})
			var archive bytes.Buffer
			gomega.Expect(ArchiveWithSubmodules(ctx, &archive, filepath.Join(repoDir, ".git"), filepath.Join(SuiteData.TestDirPath, "archive-cache"), ArchiveOptions{
				Commit: to, PathMatcher: matcher,
			})).To(gomega.Succeed())
			gomega.Expect(readTestTar(archive.Bytes())).To(gomega.Equal(map[string]string{"wanted.txt": "after\n"}))
			for _, direction := range []struct{ from, to, removed, added string }{
				{from, to, "before", "after"},
				{to, from, "after", "before"},
			} {
				var patch bytes.Buffer
				descriptor, err := Patch(ctx, &patch, filepath.Join(repoDir, ".git"), cacheDir, true, PatchOptions{
					FromCommit: direction.from, ToCommit: direction.to, PathMatcher: matcher,
				})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(descriptor.Paths).To(gomega.ConsistOf("wanted.txt"))
				gomega.Expect(patch.String()).To(gomega.ContainSubstring("-" + direction.removed))
				gomega.Expect(patch.String()).To(gomega.ContainSubstring("+" + direction.added))
			}
		},
		ginkgo.Entry("without host filter, one slot", false, "1"),
		ginkgo.Entry("native host filter, one slot", true, "1"),
		ginkgo.Entry("native host filter, two slots", true, "2"),
	)
})
