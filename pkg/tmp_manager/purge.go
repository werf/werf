package tmp_manager

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"slices"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/werf"
)

func Purge(ctx context.Context, dryRun bool) error {
	return logboek.Context(ctx).LogProcess("Running purge for tmp data").DoError(func() error {
		return purge(ctx, dryRun)
	})
}

func purge(ctx context.Context, dryRun bool) error {
	tmpFiles, err := ioutil.ReadDir(werf.GetTmpDir())
	if err != nil {
		return fmt.Errorf("unable to list tmp files in %s: %w", werf.GetTmpDir(), err)
	}

	filesToRemove := make([]string, 0, len(tmpFiles))
	projectDirsToRemove := make([]string, 0, len(tmpFiles))

	for _, finfo := range tmpFiles {
		if strings.HasPrefix(finfo.Name(), projectDirPrefix) {
			projectDirsToRemove = append(projectDirsToRemove, filepath.Join(werf.GetTmpDir(), finfo.Name()))
		}

		if strings.HasPrefix(finfo.Name(), commonPrefix) {
			filesToRemove = append(filesToRemove, filepath.Join(werf.GetTmpDir(), finfo.Name()))
		}
	}

	return runGCForPaths(ctx, dryRun, slices.Concat(projectDirsToRemove, filesToRemove))
}
