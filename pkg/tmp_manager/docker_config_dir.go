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

		pathsToSkip := []string{"run", "mutagen"}
		for _, file := range files {
			for _, pathToSkip := range pathsToSkip {
				if file.Name() == pathToSkip {
					continue
				}
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
