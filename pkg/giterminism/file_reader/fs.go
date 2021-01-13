package file_reader

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/util"
)

// Glob returns the hash of regular files and their contents for the paths that are matched pattern
// This function follows only symlinks pointed to a regular file (not to a directory)
func (r FileReader) filesGlob(pattern string) ([]string, error) {
	var result []string
	err := util.WalkByPattern(r.manager.ProjectDir(), pattern, func(path string, s os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if s.IsDir() {
			return nil
		}

		var filePath string
		if s.Mode()&os.ModeSymlink == os.ModeSymlink {
			link, err := filepath.EvalSymlinks(path)
			if err != nil {
				return fmt.Errorf("eval symlink %s failed: %s", path, err)
			}

			linkStat, err := os.Lstat(link)
			if err != nil {
				return fmt.Errorf("lstat %s failed: %s", linkStat, err)
			}

			if linkStat.IsDir() {
				return nil
			}

			filePath = link
		} else {
			filePath = path
		}

		if util.IsSubpathOfBasePath(r.manager.ProjectDir(), filePath) {
			relPath := util.GetRelativeToBaseFilepath(r.manager.ProjectDir(), filePath)
			result = append(result, relPath)
		} else {
			return fmt.Errorf("unable to handle the file %s which is located outside the project directory", filePath)
		}

		return nil
	})

	return result, err
}

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
