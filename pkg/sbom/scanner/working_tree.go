package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/werf/logboek"
)

type WorkingTree struct {
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

func NewWorkingTree() *WorkingTree {
	return &WorkingTree{
		billDir:   "sbom",
		billFiles: nil,

		containerfile: "Containerfile",
		containerfileContent: []byte(`
FROM scratch

COPY ./sbom /sbom
`),
	}
}

func (wt *WorkingTree) Create(_ context.Context, baseDir string, paths []string) error {
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

		if err = os.MkdirAll(billFileDir, 0o700); err != nil {
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

func (wt *WorkingTree) Cleanup(ctx context.Context) {
	for _, billFile := range wt.billFiles {
		if err := billFile.Close(); err != nil {
			logboek.Context(ctx).Warn().LogF("closing bill file %q\n", billFile.Name())
		}
	}
	if err := os.RemoveAll(wt.rootDir); err != nil {
		logboek.Context(ctx).Warn().LogF("removing sbom working tree %q\n", wt.rootDir)
	}
}

func (wt *WorkingTree) RootDir() string {
	return wt.rootDir
}

func (wt *WorkingTree) BillsDir() string {
	return wt.billDir
}

func (wt *WorkingTree) Containerfile() string {
	return wt.containerfile
}

func (wt *WorkingTree) ContainerfileContent() []byte {
	return slices.Clone(wt.containerfileContent)
}

func (wt *WorkingTree) BillFiles() []*os.File {
	return slices.Clone(wt.billFiles)
}

func (wt *WorkingTree) BillPaths() []string {
	return slices.Clone(wt.billPaths)
}

func (wt *WorkingTree) WriteBOMToFirstFile(bomJSON []byte) error {
	if len(wt.billFiles) == 0 {
		return fmt.Errorf("no bill files available")
	}

	file := wt.billFiles[0]

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("seek to start: %w", err)
	}

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("truncate file: %w", err)
	}

	if _, err := file.Write(bomJSON); err != nil {
		return fmt.Errorf("write BOM: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync file: %w", err)
	}

	return nil
}
