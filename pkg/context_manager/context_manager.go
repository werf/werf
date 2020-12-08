package context_manager

import (
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
	var filePathListRelativeToContextToCalculate []string
	for _, addFileRelativeToContext := range contextAddFile {
		addFileRelativeToProject := filepath.Join(contextDir, addFileRelativeToContext)
		if !matcher.MatchPath(addFileRelativeToProject) {
			continue
		}

		filePathListRelativeToContextToCalculate = append(filePathListRelativeToContextToCalculate, addFileRelativeToProject)
	}

	if len(filePathListRelativeToContextToCalculate) == 0 {
		return "", nil
	}

	h := sha256.New()
	for _, pathRelativeToContext := range filePathListRelativeToContextToCalculate {
		h.Write([]byte(pathRelativeToContext))

		absolutePath := filepath.Join(projectDir, pathRelativeToContext)
		if exists, err := util.RegularFileExists(absolutePath); err != nil {
			return "", fmt.Errorf("error accessing %q: %s", absolutePath, err)
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
				return fmt.Errorf("error reading %q: %s", absolutePath, err)
			}

			return nil
		}(); err != nil {
			return "", err
		}

		logboek.Context(ctx).Debug().LogF("File was added: %q\n", pathRelativeToContext)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
