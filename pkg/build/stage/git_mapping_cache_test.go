package stage

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/true_git"
)

var _ = ginkgo.Describe("Git mapping cache eviction", func() {
	ginkgo.It("keeps a prepared archive readable after its shared cache is removed", func() {
		root := ginkgo.GinkgoT().TempDir()
		cache := filepath.Join(root, "cache")
		gomega.Expect(os.MkdirAll(cache, 0o700)).To(gomega.Succeed())
		originalManager := git_repo.CommonGitDataManager
		git_repo.CommonGitDataManager = gitdata.NewGitDataManager(cache, cache, root)
		ginkgo.DeferCleanup(func() { git_repo.CommonGitDataManager = originalManager })
		source := filepath.Join(cache, "stage.tar")
		gomega.Expect(os.WriteFile(source, []byte("stage contents"), 0o600)).To(gomega.Succeed())
		gm := NewGitMapping()
		gm.ScriptsDir = filepath.Join(root, "scripts")
		gm.ContainerScriptsDir = "/scripts"
		gm.ContainerArchivesDir = "/shared-archives"
		file, err := gm.prepareArchiveFile(&git_repo.ArchiveFile{FilePath: source})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(os.RemoveAll(cache)).To(gomega.Succeed())
		data, err := os.ReadFile(file.FilePath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(string(data)).To(gomega.Equal("stage contents"))
		relative, err := filepath.Rel(gm.ScriptsDir, file.FilePath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(file.ContainerFilePath).To(gomega.Equal(filepath.ToSlash(filepath.Join(gm.ContainerScriptsDir, relative))))
	})
	ginkgo.It("keeps a prepared v2 text patch after shared-cache eviction", func() {
		root := ginkgo.GinkgoT().TempDir()
		source := filepath.Join(root, "stage.patch")
		gomega.Expect(os.WriteFile(source, []byte("text patch"), 0o600)).To(gomega.Succeed())
		gm := NewGitMapping()
		gm.ScriptsDir = filepath.Join(root, "scripts")
		gm.ContainerScriptsDir = "/scripts"
		file, err := gm.preparePatchFile(&git_repo.PatchFile{FilePath: source})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(os.Remove(source)).To(gomega.Succeed())
		gomega.Expect(os.ReadFile(file.FilePath)).To(gomega.Equal([]byte("text patch")))
	})

	ginkgo.It("keeps each pinned input immutable when another stage uses the same filename", func() {
		root := ginkgo.GinkgoT().TempDir()
		gm := NewGitMapping()
		gm.ScriptsDir = filepath.Join(root, "scripts")
		gm.ContainerScriptsDir = "/scripts"
		var pins []*ContainerFileDescriptor
		for _, value := range []string{"first", "second"} {
			dir := filepath.Join(root, value)
			gomega.Expect(os.Mkdir(dir, 0o700)).To(gomega.Succeed())
			source := filepath.Join(dir, "stage.tar")
			gomega.Expect(os.WriteFile(source, []byte(value), 0o600)).To(gomega.Succeed())
			pin, err := gm.pinGitDataFile(source)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			pins = append(pins, pin)
			gomega.Expect(os.RemoveAll(dir)).To(gomega.Succeed())
		}
		gomega.Expect(pins[0].FilePath).NotTo(gomega.Equal(pins[1].FilePath))
		for i, value := range []string{"first", "second"} {
			data, err := os.ReadFile(pins[i].FilePath)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(string(data)).To(gomega.Equal(value))
		}
	})

	ginkgo.It("copies the input when the command directory is on another filesystem", func() {
		if _, err := os.Stat("/dev/shm"); os.IsNotExist(err) {
			ginkgo.Skip("cross-device fixture needs /dev/shm")
		} else {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}
		sourceDir, err := os.MkdirTemp("/dev/shm", "werf-git-pin-test-")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.DeferCleanup(func() { gomega.Expect(os.RemoveAll(sourceDir)).To(gomega.Succeed()) })
		source := filepath.Join(sourceDir, "stage.tar")
		gomega.Expect(os.WriteFile(source, []byte("cross-device"), 0o600)).To(gomega.Succeed())
		root := ginkgo.GinkgoT().TempDir()
		err = os.Link(source, filepath.Join(root, "link-probe"))
		if err == nil {
			ginkgo.Skip("temporary directories share a filesystem")
		}
		gomega.Expect(errors.Is(err, syscall.EXDEV)).To(gomega.BeTrue(), "%v", err)
		gm := NewGitMapping()
		gm.ScriptsDir = filepath.Join(root, "scripts")
		gm.ContainerScriptsDir = "/scripts"
		pin, err := gm.pinGitDataFile(source)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(os.Remove(source)).To(gomega.Succeed())
		data, err := os.ReadFile(pin.FilePath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(string(data)).To(gomega.Equal("cross-device"))
	})

	ginkgo.It("writes patch sidecars into the command directory instead of the shared cache", func() {
		root := ginkgo.GinkgoT().TempDir()
		cache := filepath.Join(root, "cache")
		patchDir := filepath.Join(cache, "repo", "ab")
		gomega.Expect(os.MkdirAll(patchDir, 0o700)).To(gomega.Succeed())
		originalManager := git_repo.CommonGitDataManager
		git_repo.CommonGitDataManager = gitdata.NewGitDataManager(cache, cache, root)
		ginkgo.DeferCleanup(func() { git_repo.CommonGitDataManager = originalManager })

		patchPath := filepath.Join(patchDir, "abcd.patch")
		gomega.Expect(os.WriteFile(patchPath, []byte("patch payload"), 0o644)).To(gomega.Succeed())

		gm := NewGitMapping()
		gm.ScriptsDir = filepath.Join(root, "scripts")
		gm.ContainerScriptsDir = "/scripts"
		gm.ContainerPatchesDir = "/patches"
		gm.To = "/app"
		gm.SetGitRepo(&gitRepoNameStub{})
		patch := &git_repo.PatchFile{FilePath: patchPath, Descriptor: &true_git.PatchDescriptor{Paths: []string{"a.txt"}}}

		before := listTree(cache)

		pathsListFile, err := gm.preparePatchPathsListFile(patch)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(listTree(cache)).To(gomega.Equal(before))

		for _, file := range []*ContainerFileDescriptor{pathsListFile} {
			relative, err := filepath.Rel(gm.ScriptsDir, file.FilePath)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(relative).NotTo(gomega.HavePrefix(".."))
			gomega.Expect(file.FilePath).To(gomega.BeAnExistingFile())
			gomega.Expect(file.ContainerFilePath).To(gomega.Equal(filepath.ToSlash(filepath.Join(gm.ContainerScriptsDir, relative))))
		}

		gomega.Expect(os.ReadFile(pathsListFile.FilePath)).To(gomega.Equal([]byte("/app/a.txt")))
	})

	ginkgo.DescribeTable("preserves the input file mode when it has to copy across filesystems",
		func(mode os.FileMode) {
			root := ginkgo.GinkgoT().TempDir()
			source := filepath.Join(root, "stage.tar")
			gomega.Expect(os.WriteFile(source, []byte("contents"), 0o600)).To(gomega.Succeed())
			gomega.Expect(os.Chmod(source, mode)).To(gomega.Succeed())

			originalLink := linkFile
			linkFile = func(string, string) error { return syscall.EXDEV }
			ginkgo.DeferCleanup(func() { linkFile = originalLink })

			gm := NewGitMapping()
			gm.ScriptsDir = filepath.Join(root, "scripts")
			gm.ContainerScriptsDir = "/scripts"
			pin, err := gm.pinGitDataFile(source)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			info, err := os.Stat(pin.FilePath)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(info.Mode().Perm()).To(gomega.Equal(mode))
			gomega.Expect(os.ReadFile(pin.FilePath)).To(gomega.Equal([]byte("contents")))
		},
		ginkgo.Entry("executable", os.FileMode(0o755)),
		ginkgo.Entry("group readable", os.FileMode(0o640)),
	)

	ginkgo.It("returns an error rather than a path to a missing input", func() {
		root := ginkgo.GinkgoT().TempDir()
		gm := NewGitMapping()
		gm.ScriptsDir = filepath.Join(root, "scripts")
		gm.ContainerScriptsDir = "/scripts"
		pin, err := gm.pinGitDataFile(filepath.Join(root, "missing"))
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("open git input")))
		gomega.Expect(pin).To(gomega.BeNil())
	})
})
