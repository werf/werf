package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/werf/logboek"
)

type workingTree struct {
	// rootDir is absolute path to dir.
	rootDir string

	// billDir is relative path to SBOM scanning artifacts.
	billDir string
	// billFiles is list of files.
	billFiles []*os.File
	// billPaths is list of relative file names.
	billPaths []string

	// containerfile is relative path to Containerfile.
	containerfile string
	// containerfileContent is content of Containerfile.
	containerfileContent []byte
}

func NewWorkingTree() *workingTree {
	return &workingTree{
		billDir:   "sbom",
		billFiles: nil,

		containerfile: "Containerfile",
		containerfileContent: []byte(`
FROM scratch

COPY ./sbom /sbom
`),
	}
}

func (wt *workingTree) Create(_ context.Context, baseDir string, paths []string) error {
	var err error

	if wt.rootDir, err = os.MkdirTemp(baseDir, fmt.Sprintf("sbom-")); err != nil {
		return fmt.Errorf("unable to create root directory fot sbom working tree: %w", err)
	}

	if err = os.Mkdir(filepath.Join(wt.rootDir, wt.billDir), 0o700); err != nil {
		return fmt.Errorf("unable to create artifacts directory in sbom working tree: %w", err)
	}

	l1 := len(paths)

	wt.billFiles = make([]*os.File, l1)
	wt.billPaths = make([]string, l1)

	for i, name := range paths {
		billFileDir := filepath.Join(wt.rootDir, wt.billDir, filepath.Dir(name))

		if err = os.Mkdir(billFileDir, 0o700); err != nil {
			return fmt.Errorf("unable to create %q: %w", billFileDir, err)
		}

		wt.billPaths[i] = name
		billAbsPath := filepath.Join(wt.rootDir, wt.billDir, wt.billPaths[i])

		if wt.billFiles[i], err = os.OpenFile(billAbsPath, os.O_CREATE|os.O_WRONLY, 0o666); err != nil {
			return fmt.Errorf("unable to create %q: %w", billAbsPath, err)
		}
	}

	var containerFile *os.File
	if containerFile, err = os.Create(filepath.Join(wt.rootDir, wt.containerfile)); err != nil {
		return fmt.Errorf("unable to create %q: %w", wt.containerfile, err)
	}
	defer containerFile.Close()

	if _, err = containerFile.Write(wt.containerfileContent); err != nil {
		return fmt.Errorf("unable to write into %q: %w", wt.containerfile, err)
	}

	return nil
}

func (wt *workingTree) Cleanup(ctx context.Context) {
	for _, billFile := range wt.billFiles {
		if err := billFile.Close(); err != nil {
			logboek.Context(ctx).Warn().LogF("closing bill file %q\n", billFile.Name())
		}
	}
	if err := os.RemoveAll(wt.rootDir); err != nil {
		logboek.Context(ctx).Warn().LogF("removing sbom working tree %q\n", wt.rootDir)
	}
}

func (wt *workingTree) RootDir() string {
	return wt.rootDir
}

func (wt *workingTree) BillsDir() string {
	return wt.billDir
}

func (wt *workingTree) Containerfile() string {
	return wt.containerfile
}

func (wt *workingTree) ContainerfileContent() []byte {
	return slices.Clone(wt.containerfileContent)
}

func (wt *workingTree) BillFiles() []*os.File {
	return slices.Clone(wt.billFiles)
}

func (wt *workingTree) BillPaths() []string {
	return slices.Clone(wt.billPaths)
}
