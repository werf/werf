package tmp_manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
)

func CreateDockerConfigDir(fromDockerConfig string) (string, error) {
	newDir, err := newTmpDir(DockerConfigDirPrefix)
	if err != nil {
		return "", err
	}

	if err := os.Chmod(newDir, 0700); err != nil {
		return "", err
	}

	if _, err := os.Stat(fromDockerConfig); !os.IsNotExist(err) {
		err := copy.Copy(fromDockerConfig, newDir)
		if err != nil {
			return "", fmt.Errorf("unable to copy %s to %s: %s", fromDockerConfig, newDir, err)
		}
	}

	if err := registerCreatedPath(newDir, filepath.Join(GetCreatedTmpDirs(), dockerConfigsServiceDir)); err != nil {
		os.RemoveAll(newDir)
		return "", err
	}

	shouldRunGC, err := checkShouldRunGC()
	if err != nil {
		return "", err
	}

	if shouldRunGC {
		err := runGC()
		if err != nil {
			return "", fmt.Errorf("tmp manager GC failed: %s", err)
		}
	}

	return newDir, nil
}
