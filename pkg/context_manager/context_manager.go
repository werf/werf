package context_manager

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

func GetTmpDir() string {
	return filepath.Join(werf.GetServiceDir(), "tmp", "context")
}

func GetTmpArchivePath() string {
	return filepath.Join(GetTmpDir(), uuid.NewV4().String())
}

func ContextAddFileChecksum(ctx context.Context, projectDir string, contextDir string, contextAddFile []string, matcher path_matcher.PathMatcher) (string, error) {
	var filePathListRelativeToProject []string
	for _, addFileRelativeToContext := range contextAddFile {
		addFileRelativeToProject := filepath.Join(contextDir, addFileRelativeToContext)
		if !matcher.IsPathMatched(addFileRelativeToProject) {
			continue
		}

		filePathListRelativeToProject = append(filePathListRelativeToProject, addFileRelativeToProject)
	}

	if len(filePathListRelativeToProject) == 0 {
		return "", nil
	}

	h := sha256.New()
	for _, pathRelativeToProject := range filePathListRelativeToProject {
		pathWithSlashes := filepath.ToSlash(pathRelativeToProject)
		h.Write([]byte(pathWithSlashes))

		absolutePath := filepath.Join(projectDir, pathRelativeToProject)
		if exists, err := util.RegularFileExists(absolutePath); err != nil {
			return "", fmt.Errorf("unable to check existence of file %q: %s", absolutePath, err)
		} else if !exists {
			continue
		}

		if err := func() error {
			f, err := os.Open(absolutePath)
			if err != nil {
				return fmt.Errorf("unable to open %q: %s", absolutePath, err)
			}
			defer f.Close()

			if _, err := io.Copy(h, f); err != nil {
				return fmt.Errorf("unable to copy file %q: %s", absolutePath, err)
			}

			return nil
		}(); err != nil {
			return "", err
		}

		logboek.Context(ctx).Debug().LogF("File was added: %q\n", pathWithSlashes)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func AddContextAddFileToContextArchive(ctx context.Context, originalArchivePath string, projectDir string, contextDir string, contextAddFile []string) (string, error) {
	destinationArchivePath := GetTmpArchivePath()

	pathsToExcludeFromSourceArchive := contextAddFile
	if err := util.CreateArchiveBasedOnAnotherOne(ctx, originalArchivePath, destinationArchivePath, pathsToExcludeFromSourceArchive, func(tw *tar.Writer) error {
		for _, contextAddFile := range contextAddFile {
			sourceFilePath := filepath.Join(projectDir, contextDir, contextAddFile)
			tarEntryName := filepath.ToSlash(contextAddFile)
			if err := util.CopyFileIntoTar(tw, tarEntryName, sourceFilePath); err != nil {
				return fmt.Errorf("unable to add contextAddFile %q to archive %q: %s", sourceFilePath, destinationArchivePath, err)
			}

			logboek.Context(ctx).Debug().LogF("Extra file was added: %q\n", tarEntryName)
		}

		return nil
	}); err != nil {
		return "", err
	}

	return destinationArchivePath, nil
}
