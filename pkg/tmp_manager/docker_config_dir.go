package tmp_manager

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
)

func CreateDockerConfigDir(ctx context.Context, fromDockerConfig string) (string, error) {
	newDir, err := newTmpDir(DockerConfigDirPrefix)
	if err != nil {
		return "", err
	}

	if err := os.Chmod(newDir, 0o700); err != nil {
		return "", err
	}

	if _, err := os.Stat(fromDockerConfig); !os.IsNotExist(err) {
		files, err := ioutil.ReadDir(fromDockerConfig)
		if err != nil {
			return "", fmt.Errorf("unable to read %q", fromDockerConfig)
		}

		for _, file := range files {
			if file.Name() == "run" {
				continue
			}

			source := filepath.Join(fromDockerConfig, file.Name())
			destination := filepath.Join(newDir, file.Name())
			err := copy.Copy(source, destination)
			if err != nil {
				return "", fmt.Errorf("unable to copy %q to %q: %w", source, destination, err)
			}
		}
	}

	if err := registerCreatedPath(newDir, filepath.Join(GetCreatedTmpDirs(), dockerConfigsServiceDir)); err != nil {
		os.RemoveAll(newDir)
		return "", err
	}

	return newDir, nil
}
