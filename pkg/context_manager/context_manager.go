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

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/pkg/tmp_manager"
)

func getTmpArchivePath() string {
	return filepath.Join(tmp_manager.GetContextTmpDir(), uuid.NewString())
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

type AddContextAddFilesToContextArchiveOpts struct {
	OriginalArchivePath    string
	ProjectDir             string
	ContextDir             string
	ContextAddFiles        []string
	ContextAddFilesFromMem map[string][]byte // map[tarEntryName]data
}

func AddContextAddFilesToContextArchive(ctx context.Context, opts *AddContextAddFilesToContextArchiveOpts) (string, error) {
	destinationArchivePath := getTmpArchivePath()

	pathsToExcludeFromSourceArchive := opts.ContextAddFiles
	if err := util.CreateArchiveBasedOnAnotherOne(ctx, opts.OriginalArchivePath, destinationArchivePath, util.CreateArchiveOptions{
		CopyTarOptions: util.CopyTarOptions{ExcludePaths: pathsToExcludeFromSourceArchive},
		AfterCopyFunc: func(tw *tar.Writer) error {
			addFilePathsToCopy, err := GetContextAddFilesPaths(opts.ProjectDir, opts.ContextDir, opts.ContextAddFiles)
			if err != nil {
				return err
			}

			for _, addFilePathToCopy := range addFilePathsToCopy {
				tarEntryName, err := filepath.Rel(filepath.Join(opts.ProjectDir, opts.ContextDir), addFilePathToCopy)
				if err != nil {
					return fmt.Errorf("unable to get context relative path for %q: %w", addFilePathToCopy, err)
				}
				tarEntryName = filepath.ToSlash(tarEntryName)
				if err := util.CopyFileIntoTar(tw, tarEntryName, addFilePathToCopy); err != nil {
					return fmt.Errorf("unable to add contextAddFile %q to archive %q: %w", addFilePathToCopy, destinationArchivePath, err)
				}
				logboek.Context(ctx).Debug().LogF("Extra file was added to the current context: %q\n", tarEntryName)
			}

			for tarEntryName, data := range opts.ContextAddFilesFromMem {
				if err := addFileToTarFromMem(tw, tarEntryName, data); err != nil {
					return fmt.Errorf("unable to add contextAddFile from memory %q to archive %q: %w", tarEntryName, destinationArchivePath, err)
				}
				logboek.Context(ctx).Debug().LogF("Extra file from memory was added to the current context: %q\n", tarEntryName)
			}
			return nil
		},
	}); err != nil {
		return "", err
	}

	return destinationArchivePath, nil
}

func addFileToTarFromMem(tw *tar.Writer, tarEntryName string, data []byte) error {
	header := &tar.Header{
		Name: tarEntryName,
		Mode: 0o600,
		Size: int64(len(data)),
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write tar header for file %s: %w", tarEntryName, err)
	}

	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("unable to write data to tar archive from file %q: %w", tarEntryName, err)
	}
	return nil
}
