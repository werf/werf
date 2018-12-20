package cleanup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/logger"
)

func RemoveLostTmpDappFiles() error {
	tmpFiles, err := ioutil.ReadDir(dapp.GetTmpDir())
	if err != nil {
		return fmt.Errorf("unable to list tmp files in %s: %s", dapp.GetTmpDir(), err)
	}

	filesToRemove := []string{}
	for _, finfo := range tmpFiles {
		if strings.HasPrefix(finfo.Name(), "dapp") {
			filesToRemove = append(filesToRemove, filepath.Join(dapp.GetTmpDir(), finfo.Name()))
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
