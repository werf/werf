package includes

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/werf/werf/v2/pkg/git_repo"
)

func (i *Include) getPathScope(ctx context.Context, commit string, repo *git_repo.Remote) (string, error) {
	var pathScope string

	gitAddIsDirOrSubmodule, err := repo.IsCommitTreeEntryDirectory(ctx, commit, i.Add)
	if err != nil {
		return "", fmt.Errorf("unable to determine whether ls tree entry for path %q on commit %q is directory or not: %w", i.Add, commit, err)
	}

	if gitAddIsDirOrSubmodule {
		pathScope = i.Add
	} else {
		pathScope = filepath.ToSlash(filepath.Dir(i.Add))
	}

	return pathScope, nil
}

func (i *Include) getFileRenames(ctx context.Context, commit string, repo *git_repo.Remote) (map[string]string, error) {
	gitAddIsDirOrSubmodule, err := repo.IsCommitTreeEntryDirectory(ctx, commit, i.Add)
	if err != nil {
		return nil, fmt.Errorf("unable to determine whether ls tree entry for path %q on commit %q is directory or not: %w", i.Add, commit, err)
	}

	fileRenames := make(map[string]string)
	if gitAddIsDirOrSubmodule {
		return fileRenames, nil
	}

	if filepath.Base(i.Add) != filepath.Base(i.To) {
		fileRenames[i.Add] = filepath.Base(i.To)
	}

	return fileRenames, nil
}

func extractTar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	tarReader := tar.NewReader(file)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		default:
			//
		}
	}

	return nil
}
