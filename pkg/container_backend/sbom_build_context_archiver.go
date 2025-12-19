package container_backend

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type SbomBuildContextArchiver struct {
	rootDir string
	tarPath string
}

func NewSbomContextArchiver(rootDir string) *SbomBuildContextArchiver {
	return &SbomBuildContextArchiver{
		rootDir: rootDir,
	}
}

func (a *SbomBuildContextArchiver) Create(_ context.Context, opts BuildContextArchiveCreateOptions) error {
	tarFile, err := os.Create(filepath.Join(a.rootDir, "sbom-docker.tar"))
	if err != nil {
		return fmt.Errorf("unable to create tar file: %w", err)
	}
	defer tarFile.Close()

	a.tarPath = tarFile.Name()

	tarWriter := tar.NewWriter(tarFile)

	for _, addFileName := range opts.ContextAddFiles {
		file, err := os.Open(filepath.Join(a.rootDir, addFileName))
		if err != nil {
			return fmt.Errorf("unable to open file for adding: %w", err)
		}
		stat, err := file.Stat()
		if err != nil {
			return fmt.Errorf("unable to stat file %q: %w", file.Name(), err)
		}

		hdr := &tar.Header{
			Name: addFileName,
			Mode: 0o600,
			Size: stat.Size(),
		}

		if err = tarWriter.WriteHeader(hdr); err != nil {
			return fmt.Errorf("unable to write tar header %v: %w", hdr, err)
		}

		if _, err = io.Copy(tarWriter, file); err != nil {
			return fmt.Errorf("unable to copy file %q: %w", file.Name(), err)
		}

		if err = file.Close(); err != nil {
			return fmt.Errorf("unable to close %q: %w", file.Name(), err)
		}
	}

	if err = tarWriter.Close(); err != nil {
		return fmt.Errorf("unable to close tar writer: %w", err)
	}

	return nil
}

func (a *SbomBuildContextArchiver) Path() string {
	return a.tarPath
}

func (a *SbomBuildContextArchiver) ExtractOrGetExtractedDir(ctx context.Context) (string, error) {
	return "", nil
}

func (a *SbomBuildContextArchiver) CalculatePathsChecksum(ctx context.Context, paths []string) (string, error) {
	return "", nil
}

func (a *SbomBuildContextArchiver) CalculateGlobsChecksum(ctx context.Context, globs []string, checkForArchive bool) (string, error) {
	return "", nil
}

func (a *SbomBuildContextArchiver) CleanupExtractedDir(ctx context.Context) {
}
