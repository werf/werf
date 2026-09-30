package stage

import (
	"archive/tar"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/git_repo"
)

var _ = ginkgo.Describe("Git archive mapping ownership", func() {
	ginkgo.DescribeTable("ignores stored ownership and applies only the current mapping",
		func(owner, group, credentials string, archiveType git_repo.ArchiveType, destination string) {
			dir := ginkgo.GinkgoT().TempDir()
			archivePath := filepath.Join(dir, "shared.tar")
			file, err := os.Create(archivePath)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			writer := tar.NewWriter(file)
			for _, header := range []*tar.Header{
				{Name: "file with space", Mode: 0o644, Uid: 1234, Gid: 5678},
				{Name: "line\nbreak", Mode: 0o644, Uid: 1234, Gid: 5678},
				{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "/untouched", Mode: 0o777, Uid: 1234, Gid: 5678},
			} {
				gomega.Expect(writer.WriteHeader(header)).To(gomega.Succeed())
			}
			gomega.Expect(writer.Close()).To(gomega.Succeed())
			gomega.Expect(file.Close()).To(gomega.Succeed())
			before, err := os.ReadFile(archivePath)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			mapping := NewGitMapping()
			mapping.Owner, mapping.Group, mapping.To = owner, group, destination
			mapping.ScriptsDir = filepath.Join(dir, "scripts")
			mapping.ContainerScriptsDir = "/scripts"
			commands, err := mapping.applyArchiveCommand(context.Background(), &ContainerFileDescriptor{FilePath: archivePath, ContainerFilePath: "/archives/shared.tar"}, archiveType)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(commands).To(gomega.HaveLen(3))
			gomega.Expect(commands[1]).To(gomega.ContainSubstring("--no-same-owner"))
			gomega.Expect(commands[2]).To(gomega.And(gomega.ContainSubstring("--null"), gomega.ContainSubstring("--no-run-if-empty"), gomega.ContainSubstring("--no-dereference"), gomega.ContainSubstring("-- "+credentials)))
			gomega.Expect(commands[2]).NotTo(gomega.ContainSubstring("--recursive"))
			files, err := os.ReadDir(mapping.ScriptsDir)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(files).To(gomega.HaveLen(1))
			info, err := files[0].Info()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(info.Mode().Perm()).To(gomega.Equal(os.FileMode(0o644)))
			paths, err := os.ReadFile(filepath.Join(mapping.ScriptsDir, files[0].Name()))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(strings.Split(string(paths), "\x00")).To(gomega.Equal([]string{"/app/file with space", "/app/line\nbreak", "/app/link", ""}))
			after, err := os.ReadFile(archivePath)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(after).To(gomega.Equal(before))
		},
		ginkgo.Entry("explicit numeric owner/group", "1001", "1002", "1001:1002", git_repo.DirectoryArchive, "/app"),
		ginkgo.Entry("default ownership after a different mapping", "", "", "0:0", git_repo.DirectoryArchive, "/app"),
		ginkgo.Entry("owner only", "1001", "", "1001:0", git_repo.DirectoryArchive, "/app"),
		ginkgo.Entry("group only", "", "1002", "0:1002", git_repo.DirectoryArchive, "/app"),
		ginkgo.Entry("named credentials resolved inside the image", "app", "staff", "app:staff", git_repo.DirectoryArchive, "/app"),
		ginkgo.Entry("single-file mapping uses its destination parent", "1001", "1002", "1001:1002", git_repo.FileArchive, "/app/renamed"),
	)
})
