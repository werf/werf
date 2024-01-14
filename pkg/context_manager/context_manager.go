package context_manager

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

func GetTmpDir() string {
	return filepath.Join(werf.GetServiceDir(), "tmp", "context")
}

func GetTmpArchivePath() string {
	return filepath.Join(GetTmpDir(), uuid.NewString())
}

func GetContextAddFilesPaths(projectDir, contextDir string, contextAddFiles []string) ([]string, error) {
	var addFilePaths []string
	for _, addFile := range contextAddFiles {
		addFilePath := filepath.Join(projectDir, contextDir, addFile)

		addFileInfo, err := os.Lstat(addFilePath)
		if err != nil {
			return nil, fmt.Errorf("unable to get file info for contextAddFile %q: %w", addFilePath, err)
		}

		if addFileInfo.IsDir() {
			if err := filepath.Walk(addFilePath, func(path string, fileInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if fileInfo.IsDir() {
					return nil
				}
				addFilePaths = append(addFilePaths, path)
				return nil
			}); err != nil {
				return nil, fmt.Errorf("error occurred when recursively walking the contextAddFile dir %q: %w", addFilePath, err)
			}
		} else {
			addFilePaths = append(addFilePaths, addFilePath)
		}
	}

	return util.UniqStrings(addFilePaths), nil
}

func ContextAddFilesChecksum(ctx context.Context, projectDir, contextDir string, contextAddFiles []string, matcher path_matcher.PathMatcher) (string, error) {
	addFilePaths, err := GetContextAddFilesPaths(projectDir, contextDir, contextAddFiles)
	if err != nil {
		return "", err
	}

	var projectRelativeAddFilePaths []string
	for _, addFilePath := range addFilePaths {
		projectRelativeAddFilePath, err := filepath.Rel(projectDir, addFilePath)
		if err != nil {
			return "", fmt.Errorf("unable to get context relative path for %q: %w", addFilePath, err)
		}
		if !matcher.IsPathMatched(projectRelativeAddFilePath) {
			continue
		}
		projectRelativeAddFilePaths = append(projectRelativeAddFilePaths, projectRelativeAddFilePath)
	}

	if len(projectRelativeAddFilePaths) == 0 {
		return "", nil
	}

	h := sha256.New()
	for _, projectRelativeAddFilePath := range projectRelativeAddFilePaths {
		projectRelativeAddFilePath = filepath.ToSlash(projectRelativeAddFilePath)
		h.Write([]byte(projectRelativeAddFilePath))

		addFilePath := filepath.Join(projectDir, projectRelativeAddFilePath)
		if exists, err := util.RegularFileExists(addFilePath); err != nil {
			return "", fmt.Errorf("unable to check existence of file %q: %w", addFilePath, err)
		} else if !exists {
			continue
		}

		if err := func() error {
			f, err := os.Open(addFilePath)
			if err != nil {
				return fmt.Errorf("unable to open %q: %w", addFilePath, err)
			}
			defer f.Close()

			if _, err := io.Copy(h, f); err != nil {
				return fmt.Errorf("unable to copy file %q: %w", addFilePath, err)
			}

			return nil
		}(); err != nil {
			return "", err
		}

		logboek.Context(ctx).Debug().LogF("File was added: %q\n", projectRelativeAddFilePath)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func AddContextAddFilesToContextArchive(ctx context.Context, originalArchivePath, projectDir, contextDir string, contextAddFiles []string) (string, error) {
	destinationArchivePath := GetTmpArchivePath()

	pathsToExcludeFromSourceArchive := contextAddFiles
	if err := util.CreateArchiveBasedOnAnotherOne(ctx, originalArchivePath, destinationArchivePath, util.CreateArchiveOptions{
		CopyTarOptions: util.CopyTarOptions{ExcludePaths: pathsToExcludeFromSourceArchive},
		AfterCopyFunc: func(tw *tar.Writer) error {
			addFilePathsToCopy, err := GetContextAddFilesPaths(projectDir, contextDir, contextAddFiles)
			if err != nil {
				return err
			}

			for _, addFilePathToCopy := range addFilePathsToCopy {
				tarEntryName, err := filepath.Rel(filepath.Join(projectDir, contextDir), addFilePathToCopy)
				if err != nil {
					return fmt.Errorf("unable to get context relative path for %q: %w", addFilePathToCopy, err)
				}
				tarEntryName = filepath.ToSlash(tarEntryName)
				if err := util.CopyFileIntoTar(tw, tarEntryName, addFilePathToCopy); err != nil {
					return fmt.Errorf("unable to add contextAddFile %q to archive %q: %w", addFilePathToCopy, destinationArchivePath, err)
				}
				logboek.Context(ctx).Debug().LogF("Extra file was added to the current context: %q\n", tarEntryName)
			}
			return nil
		},
	}); err != nil {
		return "", err
	}

	return destinationArchivePath, nil
}
