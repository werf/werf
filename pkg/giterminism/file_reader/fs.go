package file_reader

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/werf/werf/pkg/util"
)

func (r FileReader) readFile(relPath string) ([]byte, error) {
	absPath := filepath.Join(r.manager.ProjectDir(), relPath)
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %s: %s", absPath, err)
	}

	return data, nil
}

func (r FileReader) isFileExist(relPath string) (bool, error) {
	absPath := filepath.Join(r.manager.ProjectDir(), relPath)
	exist, err := util.FileExists(absPath)
	if err != nil {
		return false, fmt.Errorf("unable to check existence of file %s: %s", absPath, err)
	}

	return exist, nil
}
