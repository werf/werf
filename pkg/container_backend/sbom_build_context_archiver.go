package container_backend

import (
	"archive/tar"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type sbomBuildContextArchiver struct {
	rootDir string
	tarPath string
}

func newSbomContextArchiver(rootDir string) *sbomBuildContextArchiver {
	return &sbomBuildContextArchiver{
		rootDir: rootDir,
	}
}

func (a *sbomBuildContextArchiver) Create(_ context.Context, opts BuildContextArchiveCreateOptions) error {
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

		fileContent := make([]byte, stat.Size())

		if _, err = file.Read(fileContent); err != nil {
			return fmt.Errorf("unable to read file %q: %w", file.Name(), err)
		}

		if _, err = tarWriter.Write(fileContent); err != nil {
			return fmt.Errorf("unable to write to tar %q: %w", tarFile.Name(), err)
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

func (a *sbomBuildContextArchiver) Path() string {
	return a.tarPath
}

func (a *sbomBuildContextArchiver) ExtractOrGetExtractedDir(ctx context.Context) (string, error) {
	return "", nil
}

func (a *sbomBuildContextArchiver) CalculatePathsChecksum(ctx context.Context, paths []string) (string, error) {
	return "", nil
}

func (a *sbomBuildContextArchiver) CalculateGlobsChecksum(ctx context.Context, globs []string, checkForArchive bool) (string, error) {
	return "", nil
}

func (a *sbomBuildContextArchiver) CleanupExtractedDir(ctx context.Context) {
}
