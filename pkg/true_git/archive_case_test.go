package true_git

import (
	"archive/tar"
	"bytes"
	"io"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/test/pkg/suite_init"
)

var _ = ginkgo.DescribeTable("grouping paths for a case-insensitive filesystem",
	func(paths []string, expected [][]string) {
		gomega.Expect(groupPathsCaseInsensitively(paths)).To(gomega.Equal(expected))
	},
	ginkgo.Entry("no paths", nil, [][]string(nil)),
	ginkgo.Entry("distinct paths stay in one group", []string{"a.h", "b.h", "Dir/x", "dir/y"}, [][]string{{"a.h", "b.h", "Dir/x", "dir/y"}}),
	ginkgo.Entry("paths equal ignoring case go to separate groups", []string{"a.h", "A.h", "b.h"}, [][]string{{"a.h", "b.h"}, {"A.h"}}),
	ginkgo.Entry("every further variant opens a new group", []string{"inc/x.h", "inc/X.h", "inc/x.H", "inc/y.h", "INC/Y.h"}, [][]string{{"inc/x.h", "inc/y.h"}, {"inc/X.h", "INC/Y.h"}, {"inc/x.H"}}),
	ginkgo.Entry("file and directory equal ignoring case go to separate groups", []string{"a", "A/b"}, [][]string{{"a"}, {"A/b"}}),
	ginkgo.Entry("directories equal ignoring case share a group", []string{"a", "A/b", "a/c"}, [][]string{{"a"}, {"A/b", "a/c"}}),
	ginkgo.Entry("directories equal ignoring case stay in one group", []string{"A/x", "a/y"}, [][]string{{"A/x", "a/y"}}),
	ginkgo.Entry("nested file and directory equal ignoring case go to separate groups", []string{"d/e", "D/E/f"}, [][]string{{"d/e"}, {"D/E/f"}}),
	ginkgo.Entry("directory before file equal ignoring case go to separate groups", []string{"A/b", "a"}, [][]string{{"A/b"}, {"a"}}),
	ginkgo.Entry("file and a deeper ancestor directory equal ignoring case go to separate groups", []string{"a", "A/b/c"}, [][]string{{"a"}, {"A/b/c"}}),
	ginkgo.Entry("deeper ancestor directory before file equal ignoring case go to separate groups", []string{"A/b/c", "a"}, [][]string{{"A/b/c"}, {"a"}}),
)

var _ = ginkgo.Describe("archive content of paths differing only by case", func() {
	suite_init.NewWerfInitData(SuiteData.TmpDirData)

	ginkgo.DescribeTable("reads every tree entry from its own object",
		func(ctx ginkgo.SpecContext, contentChecksum string) {
			isolateGitConfig()
			gomega.Expect(Init(ctx, Options{})).To(gomega.Succeed())
			commitTime := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
			setEnvForSpec("GIT_COMMITTER_DATE", commitTime.Format(time.RFC3339))

			repoDir := filepath.Join(SuiteData.TestDirPath, "repo")
			gitInitRepo(ctx, repoDir)
			for path, entry := range map[string]struct{ mode, content string }{
				".gitattributes":    {"100644", "*.crlf text eol=crlf\n"},
				"inc/xt_connmark.h": {"100644", "lower\n"},
				"inc/xt_CONNMARK.h": {"100644", "UPPER\n"},
				"inc/eol.crlf":      {"100644", "a\nb\n"},
				"inc/link":          {"120000", "xt_CONNMARK.h"},
			} {
				blob := gitSucceedWithStdin(ctx, repoDir, entry.content, "hash-object", "-w", "--stdin")
				gitSucceed(ctx, repoDir, "update-index", "--add", "--cacheinfo", entry.mode+","+blob+","+path)
			}
			tree := gitSucceedTrimmed(ctx, repoDir, "write-tree")
			commit := gitSucceedTrimmed(ctx, repoDir, "commit-tree", tree, "-m", "case collision")

			var output bytes.Buffer
			gomega.Expect(Archive(ctx, &output, filepath.Join(repoDir, ".git"), filepath.Join(SuiteData.TestDirPath, "worktree-cache"), ArchiveOptions{
				Commit:          commit,
				ContentChecksum: contentChecksum,
				PathScope:       "inc",
				PathMatcher:     path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: "inc"}),
			})).To(gomega.Succeed())

			files := map[string]string{}
			links := map[string]string{}
			modTimes := map[string]time.Time{}
			reader := tar.NewReader(&output)
			for {
				header, err := reader.Next()
				if err == io.EOF {
					break
				}
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				modTimes[header.Name] = header.ModTime.UTC()
				switch header.Typeflag {
				case tar.TypeReg:
					content, err := io.ReadAll(reader)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					files[header.Name] = string(content)
				case tar.TypeSymlink:
					links[header.Name] = header.Linkname
				}
			}
			gomega.Expect(files).To(gomega.Equal(map[string]string{
				"xt_connmark.h": "lower\n",
				"xt_CONNMARK.h": "UPPER\n",
				"eol.crlf":      "a\r\nb\r\n",
			}))
			gomega.Expect(links).To(gomega.Equal(map[string]string{"link": "xt_CONNMARK.h"}))
			for name, modTime := range modTimes {
				gomega.Expect(modTime).To(gomega.Equal(commitTime), "mtime of %q", name)
			}
		},
		ginkgo.Entry("Stapel", ""),
		ginkgo.Entry("Dockerfile context", "context-checksum"),
	)

	ginkgo.DescribeTable("exports a file and a directory equal ignoring case",
		func(ctx ginkgo.SpecContext, contentChecksum string) {
			isolateGitConfig()
			gomega.Expect(Init(ctx, Options{})).To(gomega.Succeed())

			repoDir := filepath.Join(SuiteData.TestDirPath, "repo")
			gitInitRepo(ctx, repoDir)
			for path, content := range map[string]string{
				"inc/a":   "file\n",
				"inc/A/b": "nested\n",
			} {
				blob := gitSucceedWithStdin(ctx, repoDir, content, "hash-object", "-w", "--stdin")
				gitSucceed(ctx, repoDir, "update-index", "--add", "--cacheinfo", "100644,"+blob+","+path)
			}
			tree := gitSucceedTrimmed(ctx, repoDir, "write-tree")
			commit := gitSucceedTrimmed(ctx, repoDir, "commit-tree", tree, "-m", "case collision")

			var output bytes.Buffer
			gomega.Expect(Archive(ctx, &output, filepath.Join(repoDir, ".git"), filepath.Join(SuiteData.TestDirPath, "worktree-cache"), ArchiveOptions{
				Commit:          commit,
				ContentChecksum: contentChecksum,
				PathScope:       "inc",
				PathMatcher:     path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: "inc"}),
			})).To(gomega.Succeed())

			files := map[string]string{}
			reader := tar.NewReader(&output)
			for {
				header, err := reader.Next()
				if err == io.EOF {
					break
				}
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				if header.Typeflag == tar.TypeReg {
					content, err := io.ReadAll(reader)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					files[header.Name] = string(content)
				}
			}
			gomega.Expect(files).To(gomega.Equal(map[string]string{
				"a":   "file\n",
				"A/b": "nested\n",
			}))
		},
		ginkgo.Entry("Stapel", ""),
		ginkgo.Entry("Dockerfile context", "context-checksum"),
	)
})
