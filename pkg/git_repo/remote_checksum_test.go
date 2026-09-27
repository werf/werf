package git_repo

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/git_repo/repo_handle"
	"github.com/werf/werf/v3/pkg/path_matcher"
	"github.com/werf/werf/v3/pkg/true_git"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("Remote object checksums", func() {
	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
		gomega.Expect(Init(&fakeGitDataManager{})).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("matches worktree checksums without materializing files",
		func(ctx ginkgo.SpecContext, shallow bool) {
			source := newChecksumTestRepo(ctx)
			gomega.Expect(os.Mkdir(filepath.Join(source, "nested"), 0o755)).To(gomega.Succeed())
			gomega.Expect(os.WriteFile(filepath.Join(source, "nested", "data"), []byte("first"), 0o644)).To(gomega.Succeed())
			gomega.Expect(os.WriteFile(filepath.Join(source, "run"), []byte("#!/bin/sh\n"), 0o755)).To(gomega.Succeed())
			gomega.Expect(os.Symlink("nested/data", filepath.Join(source, "link"))).To(gomega.Succeed())
			commitChecksumTestRepo(ctx, source)
			firstCommit := utils.GetHeadCommit(ctx, source)
			gomega.Expect(os.WriteFile(filepath.Join(source, "nested", "data"), []byte("second"), 0o644)).To(gomega.Succeed())
			commitChecksumTestRepo(ctx, source)
			secondCommit := utils.GetHeadCommit(ctx, source)

			remote, err := OpenRemoteRepo("checksums", source, nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			if shallow {
				remote.Commit = firstCommit
			}
			gomega.Expect(remote.CloneAndFetch(ctx)).To(gomega.Succeed())
			if shallow {
				remote.Commit = secondCommit
				gomega.Expect(remote.CloneAndFetch(ctx)).To(gomega.Succeed())
				gomega.Expect(remote.mirrorKind()).To(gomega.Equal(mirrorKindShallow))
				gomega.Expect(filepath.Join(remote.GetClonePath(), "shallow")).To(gomega.BeARegularFile())
			}

			var checksums []string
			for _, commit := range []string{firstCommit, secondCommit} {
				opts := ChecksumOptions{Commit: commit, LsTreeOptions: LsTreeOptions{PathMatcher: path_matcher.NewTruePathMatcher(), AllFiles: true}}
				checksum, err := remote.GetOrCreateChecksum(ctx, opts)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(checksum).NotTo(gomega.BeEmpty())
				checksums = append(checksums, checksum)
			}
			gomega.Expect(checksums[0]).NotTo(gomega.Equal(checksums[1]))
			_, err = os.Stat(remote.getWorkTreeCacheDir(remote.getRepoID()))
			gomega.Expect(os.IsNotExist(err)).To(gomega.BeTrue())

			for i, commit := range []string{firstCommit, secondCommit} {
				err := true_git.WithWorkTree(ctx, remote.GetClonePath(), remote.getWorkTreeCacheDir(remote.getRepoID()), commit, true_git.WithWorkTreeOptions{}, func(worktree string) error {
					repository, err := true_git.GitOpenWithCustomWorktreeDir(remote.GetClonePath(), worktree)
					if err != nil {
						return err
					}
					handle, err := repo_handle.NewHandle(repository)
					if err != nil {
						return err
					}
					opts := ChecksumOptions{Commit: commit, LsTreeOptions: LsTreeOptions{PathMatcher: path_matcher.NewTruePathMatcher(), AllFiles: true}}
					checksum, err := remote.CreateChecksum(ctx, handle, opts)
					if err != nil {
						return err
					}
					gomega.Expect(checksum).To(gomega.Equal(checksums[i]))
					return nil
				})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}
		},
		ginkgo.Entry("full mirror", false),
		ginkgo.Entry("shallow mirror", true),
	)

	ginkgo.It("retains submodule file contents in checksums", func(ctx ginkgo.SpecContext) {
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_COUNT", "1")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_KEY_0", "protocol.file.allow")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_VALUE_0", "always")
		submodule := newChecksumTestRepo(ctx)
		gomega.Expect(os.WriteFile(filepath.Join(submodule, "data"), []byte("first"), 0o644)).To(gomega.Succeed())
		commitChecksumTestRepo(ctx, submodule)
		source := newChecksumTestRepo(ctx)
		utils.RunSucceedCommand(ctx, source, "git", "submodule", "add", submodule, "sub")
		commitChecksumTestRepo(ctx, source)
		remote, err := OpenRemoteRepo("submodule-checksums", source, nil)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(remote.CloneAndFetch(ctx)).To(gomega.Succeed())
		opts := ChecksumOptions{Commit: utils.GetHeadCommit(ctx, source), LsTreeOptions: LsTreeOptions{PathScope: "sub", PathMatcher: path_matcher.NewTruePathMatcher(), AllFiles: true}}
		first, err := remote.GetOrCreateChecksum(ctx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(remote.getWorkTreeCacheDir(remote.getRepoID())).To(gomega.BeADirectory())
		gomega.Expect(os.WriteFile(filepath.Join(source, "sub", "data"), []byte("second"), 0o644)).To(gomega.Succeed())
		commitChecksumTestRepo(ctx, filepath.Join(source, "sub"))
		utils.RunSucceedCommand(ctx, submodule, "git", "fetch", filepath.Join(source, "sub"), "HEAD")
		commitChecksumTestRepo(ctx, source)
		gomega.Expect(remote.CloneAndFetch(ctx)).To(gomega.Succeed())
		opts.Commit = utils.GetHeadCommit(ctx, source)
		second, err := remote.GetOrCreateChecksum(ctx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(second).NotTo(gomega.Equal(first))
	})
})
