package cleanup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/werf"
)

func RemoveLostTmpWerfFiles() error {
	tmpFiles, err := ioutil.ReadDir(werf.GetTmpDir())
	if err != nil {
		return fmt.Errorf("unable to list tmp files in %s: %s", werf.GetTmpDir(), err)
	}

	filesToRemove := []string{}
	for _, finfo := range tmpFiles {
		if strings.HasPrefix(finfo.Name(), "werf") {
			filesToRemove = append(filesToRemove, filepath.Join(werf.GetTmpDir(), finfo.Name()))
		}
	}

	for _, file := range filesToRemove {
		err := os.RemoveAll(file)
		if err != nil {
			logger.LogWarningF("WARNING: unable to remove %s: %s\n", file, err)
		}
	}

	return nil
}
