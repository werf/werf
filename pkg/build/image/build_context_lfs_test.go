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
		content := "Dockerfile LFS content\x00"
		oid := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		pointer := fmt.Sprintf("version https://git-lfs.github.com/spec/v1\noid sha256:%s\nsize %d\n", oid, len(content))
		projectDir := newProjectRepo(ctx, dockerfileProjectFiles("\nimage: app\ndockerfile: Dockerfile\n", map[string]string{
			"Dockerfile":          "FROM scratch\nCOPY assets/wanted.bin /payload/\n",
			"assets/wanted.bin":   pointer,
			"assets/excluded.bin": "version https://git-lfs.github.com/spec/v1\noid sha256:0000000000000000000000000000000000000000000000000000000000000000\nsize 7\n",
			".dockerignore":       "assets/excluded.bin\n.lfsconfig\n",
		}))
		utils.WriteFile(filepath.Join(projectDir, ".git", "lfs", "objects", oid[:2], oid[2:4], oid), []byte(content))
		utils.RunSucceedCommand(ctx, projectDir, "git", "remote", "add", "origin", "file://"+projectDir)
		archive := NewBuildContextArchive(giterminismManagerOf(ctx, projectDir), ginkgo.GinkgoT().TempDir())
		opts := container_backend.BuildContextArchiveCreateOptions{DockerfileRelToContextPath: "Dockerfile"}
		gomega.Expect(archive.Create(ctx, opts)).To(gomega.Succeed())
		entries := lastEntryContents(openedContextEntries(ctx, archive))
		gomega.Expect(entries).To(gomega.HaveKeyWithValue("assets/wanted.bin", content))
		gomega.Expect(entries).NotTo(gomega.HaveKey("assets/excluded.bin"))

		commitFiles(ctx, projectDir, map[string]string{".lfsconfig": "[lfs]\nurl = file://" + projectDir + "\n"})
		archive = NewBuildContextArchive(giterminismManagerOf(ctx, projectDir), ginkgo.GinkgoT().TempDir())
		gomega.Expect(archive.Create(ctx, opts)).To(gomega.Succeed())
		entries = lastEntryContents(openedContextEntries(ctx, archive))
		gomega.Expect(entries).To(gomega.HaveKeyWithValue("assets/wanted.bin", content))
		gomega.Expect(entries).NotTo(gomega.HaveKey(".lfsconfig"))
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
})
