package sbom

import (
	"bytes"
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
	// billList is list of relative file paths.
	billList []string

	// containerfile is relative path to Containerfile.
	containerfile string
	// containerfileContent is content of Containerfile.
	containerfileContent []byte
}

func NewWorkingTree() *workingTree {
	return &workingTree{
		billDir:  "sbom",
		billList: nil,

		containerfile: "Containerfile",
		containerfileContent: []byte(`
FROM scratch

COPY ./sbom /sbom
`),
	}
}

func (wt *workingTree) Create(_ context.Context, baseDir string, names []string) error {
	var err error

	if wt.rootDir, err = os.MkdirTemp(baseDir, fmt.Sprintf("sbom-")); err != nil {
		return fmt.Errorf("unable to create root directory fot sbom working tree: %w", err)
	}

	if err = os.Mkdir(filepath.Join(wt.rootDir, wt.billDir), 0o700); err != nil {
		return fmt.Errorf("unable to create artifacts directory in sbom working tree: %w", err)
	}

	wt.billList = make([]string, len(names))

	for i, name := range names {
		wt.billList[i] = filepath.Join(wt.billDir, name)

		var file *os.File
		if file, err = os.Create(wt.billList[i]); err != nil {
			return fmt.Errorf("unable to create %q: %w", wt.billList[i], err)
		}
		file.Close()
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
	if err := os.RemoveAll(wt.rootDir); err != nil {
		logboek.Context(ctx).Warn().LogF("removing sbom working tree %q\n", wt.rootDir)
	}
}

func (wt *workingTree) RootDir() string {
	return wt.rootDir
}

func (wt *workingTree) Containerfile() string {
	return wt.containerfile
}

func (wt *workingTree) ContainerfileContent() []byte {
	return bytes.Clone(wt.containerfileContent)
}

func (wt *workingTree) BillList() []string {
	return slices.Clone(wt.billList)
}
