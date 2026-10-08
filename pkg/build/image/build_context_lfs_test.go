package image

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("automatic LFS build context", func() {
	ginkgo.It("exports selected objects with excluded .lfsconfig and refreshes changed pointers after warming the cache", func(ctx ginkgo.SpecContext) {
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_COUNT", "0")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_PARAMETERS", "")
		ginkgo.GinkgoT().Setenv("GIT_ATTR_SOURCE", "")
		gomega.Expect(os.Unsetenv("GIT_ATTR_SOURCE")).To(gomega.Succeed())
		content := "Dockerfile LFS content\x00"
		oid := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		pointer := fmt.Sprintf("version https://git-lfs.github.com/spec/v1\noid sha256:%s\nsize %d\n", oid, len(content))
		projectDir := newProjectRepo(ctx, dockerfileProjectFiles("\nimage: app\ndockerfile: Dockerfile\n", map[string]string{
			"Dockerfile":          "FROM scratch\nCOPY assets/wanted.bin /payload/\n",
			"assets/wanted.bin":   pointer,
			"assets/excluded.bin": "version https://git-lfs.github.com/spec/v1\noid sha256:0000000000000000000000000000000000000000000000000000000000000000\nsize 7\n",
			".dockerignore":       "assets/excluded.bin\n.lfsconfig\nignored.txt\n",
			".lfsconfig":          "[lfs]\n",
			"ignored.txt":         "first\n",
		}))
		utils.WriteFile(filepath.Join(projectDir, ".git", "lfs", "objects", oid[:2], oid[2:4], oid), []byte(content))
		utils.RunSucceedCommand(ctx, projectDir, "git", "remote", "add", "origin", "file://"+projectDir)
		archive := NewBuildContextArchive(giterminismManagerOf(ctx, projectDir), ginkgo.GinkgoT().TempDir())
		opts := container_backend.BuildContextArchiveCreateOptions{DockerfileRelToContextPath: "Dockerfile"}
		gomega.Expect(archive.Create(ctx, opts)).To(gomega.Succeed())
		entries := lastEntryContents(openedContextEntries(ctx, archive))
		gomega.Expect(entries).To(gomega.HaveKeyWithValue("assets/wanted.bin", content))
		gomega.Expect(entries).NotTo(gomega.HaveKey("assets/excluded.bin"))

		initialArchive, err := os.Stat(archive.Path())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		commitFiles(ctx, projectDir, map[string]string{"ignored.txt": "second\n"})
		archive = NewBuildContextArchive(giterminismManagerOf(ctx, projectDir), ginkgo.GinkgoT().TempDir())
		gomega.Expect(archive.Create(ctx, opts)).To(gomega.Succeed())
		reusedArchive, err := os.Stat(archive.Path())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(os.SameFile(initialArchive, reusedArchive)).To(gomega.BeTrue(), "excluded ordinary change must reuse the cached archive")

		commitFiles(ctx, projectDir, map[string]string{".lfsconfig": "[lfs]\nurl = file://" + projectDir + "\n"})
		archive = NewBuildContextArchive(giterminismManagerOf(ctx, projectDir), ginkgo.GinkgoT().TempDir())
		gomega.Expect(archive.Create(ctx, opts)).To(gomega.Succeed())
		changedArchive, err := os.Stat(archive.Path())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(os.SameFile(reusedArchive, changedArchive)).To(gomega.BeFalse(), "excluded .lfsconfig must invalidate the cached archive")
		entries = lastEntryContents(openedContextEntries(ctx, archive))
		gomega.Expect(entries).To(gomega.HaveKeyWithValue("assets/wanted.bin", content))
		gomega.Expect(entries).NotTo(gomega.HaveKey(".lfsconfig"))
		gomega.Expect(entries).NotTo(gomega.HaveKey("ignored.txt"))
		gomega.Expect(entries).NotTo(gomega.HaveKey("assets/excluded.bin"))

		content = "updated Dockerfile LFS content\x00"
		oid = fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		utils.WriteFile(filepath.Join(projectDir, ".git", "lfs", "objects", oid[:2], oid[2:4], oid), []byte(content))
		pointer = fmt.Sprintf("version https://git-lfs.github.com/spec/v1\noid sha256:%s\nsize %d\n", oid, len(content))
		commitFiles(ctx, projectDir, map[string]string{"assets/wanted.bin": pointer})
		archive = NewBuildContextArchive(giterminismManagerOf(ctx, projectDir), ginkgo.GinkgoT().TempDir())
		gomega.Expect(archive.Create(ctx, opts)).To(gomega.Succeed())
		entries = lastEntryContents(openedContextEntries(ctx, archive))
		gomega.Expect(entries).To(gomega.HaveKeyWithValue("assets/wanted.bin", content))
		gomega.Expect(entries).NotTo(gomega.HaveKey("assets/excluded.bin"))
	})

	ginkgo.It("ignores excluded unavailable LFS objects with ordinary submodules", func(ctx ginkgo.SpecContext) {
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_COUNT", "1")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_KEY_0", "protocol.file.allow")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_VALUE_0", "always")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_PARAMETERS", "")
		ginkgo.GinkgoT().Setenv("GIT_LFS_SKIP_SMUDGE", "0")
		ginkgo.GinkgoT().Setenv("GIT_LFS_SKIP_DOWNLOAD_ERRORS", "0")
		projectDir := newProjectRepo(ctx, dockerfileProjectFiles("\nimage: app\ndockerfile: Dockerfile\n", map[string]string{
			"Dockerfile":     "FROM scratch\nCOPY wanted.txt /wanted.txt\n",
			"wanted.txt":     "wanted content\n",
			"excluded.bin":   "version https://git-lfs.github.com/spec/v1\noid sha256:0000000000000000000000000000000000000000000000000000000000000000\nsize 7\n",
			".gitattributes": "*.bin filter=lfs diff=lfs merge=lfs -text\n",
			".dockerignore":  "excluded.bin\nsub\n",
		}))
		subDir := ginkgo.GinkgoT().TempDir()
		utils.RunSucceedCommand(ctx, subDir, "git", "init", "--initial-branch=main")
		gomega.Expect(os.WriteFile(filepath.Join(subDir, "plain.txt"), []byte("ordinary submodule\n"), 0o644)).To(gomega.Succeed())
		utils.RunSucceedCommand(ctx, subDir, "git", "add", ".")
		utils.RunSucceedCommand(ctx, subDir, "git", "-c", "user.name=review", "-c", "user.email=review@example.invalid", "-c", "commit.gpgsign=false", "commit", "-m", "ordinary submodule")
		utils.RunSucceedCommand(ctx, projectDir, "git", "submodule", "add", subDir, "sub")
		utils.RunSucceedCommand(ctx, projectDir, "git", "-c", "user.name=review", "-c", "user.email=review@example.invalid", "-c", "commit.gpgsign=false", "commit", "-m", "add ordinary submodule")
		utils.RunSucceedCommand(ctx, projectDir, "git", "remote", "add", "origin", "file://"+projectDir)
		utils.RunSucceedCommand(ctx, projectDir, "git", "lfs", "install", "--local", "--skip-repo")
		archive := NewBuildContextArchive(giterminismManagerOf(ctx, projectDir), ginkgo.GinkgoT().TempDir())
		gomega.Expect(archive.Create(ctx, container_backend.BuildContextArchiveCreateOptions{DockerfileRelToContextPath: "Dockerfile"})).To(gomega.Succeed())
		entries := lastEntryContents(openedContextEntries(ctx, archive))
		gomega.Expect(entries).To(gomega.HaveKeyWithValue("wanted.txt", "wanted content\n"))
		gomega.Expect(entries).NotTo(gomega.HaveKey("excluded.bin"))
	})
})
