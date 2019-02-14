package tmp_manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/werf"
)

func Purge(dryRun bool) error {
	return logger.LogServiceProcess("Running tmp purge", logger.LogProcessOptions{}, func() error { return purge(dryRun) })
}

func purge(dryRun bool) error {
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

	filesToRemove = append(filesToRemove, GetServiceTmpDir())

	errors := []error{}

	for _, file := range filesToRemove {
		logger.LogF("Removing %s ...\n", file)
		if !dryRun {
			err := os.RemoveAll(file)
			if err != nil {
				errors = append(errors, fmt.Errorf("unable to remove %s: %s", file, err))
			}
		}
	}

	if len(errors) > 0 {
		msg := ""
		for _, err := range errors {
			msg += fmt.Sprintf("%s\n", err)
		}
		return fmt.Errorf("%s", msg)
	}

	return nil
}
