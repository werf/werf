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

func GetContextTmpDir() string {
	return filepath.Join(werf.GetServiceDir(), "tmp", "context")
}

func GetTmpArchivePath() string {
	return filepath.Join(GetContextTmpDir(), uuid.NewV4().String())
}

func ContextAddFileChecksum(ctx context.Context, projectDir string, contextDir string, contextAddFile []string, matcher path_matcher.PathMatcher) (string, error) {
	logboek.Context(ctx).Debug().LogF("-- ContextAddFileChecksum %q %q\n", projectDir, contextAddFile)

	h := sha256.New()

	for _, addFileRelativeToContext := range contextAddFile {
		addFileRelativeToProject := filepath.Join(contextDir, addFileRelativeToContext)
		if !matcher.MatchPath(addFileRelativeToProject) {
			continue
		}

		h.Write([]byte(addFileRelativeToContext))

		addFileAbsolute := filepath.Join(projectDir, addFileRelativeToProject)
		if _, err := os.Stat(addFileAbsolute); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return "", fmt.Errorf("error accessing %q: %s", addFileAbsolute, err)
		}

		if err := func() error {
			f, err := os.Open(addFileAbsolute)
			if err != nil {
				return fmt.Errorf("unable to open %q: %s", addFileAbsolute, err)
			}
			defer f.Close()

			if _, err := io.Copy(h, f); err != nil {
				return fmt.Errorf("error reading %q: %s", addFileAbsolute, err)
			}

			return nil
		}(); err != nil {
			return "", err
		}
	}

	if h.Size() == 0 {
		return "", nil
	} else {
		return fmt.Sprintf("%x", h.Sum(nil)), nil
	}
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
